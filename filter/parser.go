package filter

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// define regexp
var (
	reFilter                  = regexp.MustCompile(`FILTER\((\d+)\)`)
	reFilterWrap              = regexp.MustCompile(`\(\s*(FILTER\((\d+)\))\s*\)`)
	reListFilters             = regexp.MustCompile(`FILTER\(\d+\)(?:\s+(AND|OR)\s+FILTER\(\d+\))+`)
	reRowPrefix               = regexp.MustCompile(`PrefixFilter\s*\(\s*'(\w+)'\s*\)`)
	reValueFilter             = regexp.MustCompile(`ValueFilter\s*\(\s*([=<>])\s*,\s*'(\w+):(.*?)'\s*\)`)
	reSingleColumnValueFilter = regexp.MustCompile(`SingleColumnValueFilter\s*\('(\w+)'\s*,\s*'(\w+)'\s*,\s*([=<>])\s*,\s*'(\w+):(.*?)'(?:\s*,\s*(?i)(true|false)\s*,\s*(true|false))?\s*\)`)
	compareOpMap              = map[string]CompareType{
		"=":  Equal,
		"!=": NotEqual,
		"<":  Less,
		"<=": LessOrEqual,
		">":  Greater,
		">=": LessOrEqual,
	}
	matchTypeMap = map[string]func(string) Comparator{
		"binary": func(s string) Comparator {
			return NewBinaryComparator(NewByteArrayComparable([]byte(s)))
		},
		"binaryprefix": func(s string) Comparator {
			return NewBinaryPrefixComparator(NewByteArrayComparable([]byte(s)))
		},
		"null": func(s string) Comparator {
			return NewNullComparator()
		},
		"regexstring": func(s string) Comparator {
			return NewRegexStringComparator(s, 0, "", "")
		},
		"substring": func(s string) Comparator {
			return NewSubstringComparator(s)
		},
	}
)

type Parser struct {
	filters []Filter
}

func (p *Parser) addFilter(f Filter) string {
	p.filters = append(p.filters, f)
	return fmt.Sprintf("FILTER(%d)", len(p.filters)-1)
}

func (p *Parser) parseList(filterStr string, op ListOperator) string {
	var filters []Filter
	for _, match := range reFilter.FindAllStringSubmatch(filterStr, -1) {
		idx, _ := strconv.Atoi(match[1])
		filters = append(filters, p.filters[idx])
	}
	newFilter := NewList(op, filters...)
	return p.addFilter(newFilter)
}

func (p *Parser) Parse(filterStr string) (Filter, error) {
	// try PrefixFilter
	for _, match := range reRowPrefix.FindAllStringSubmatch(filterStr, -1) {
		f := NewPrefixFilter([]byte(match[1]))
		s := p.addFilter(f)
		filterStr = strings.Replace(filterStr, match[0], s, 1)
	}

	// try reValueFilter
	for _, match := range reValueFilter.FindAllStringSubmatch(filterStr, -1) {
		op := match[1]
		matchType := match[2]
		value := match[3]

		if _, ok := compareOpMap[op]; !ok {
			return nil, fmt.Errorf("unsupported operator: %s", op)
		}

		if _, ok := matchTypeMap[matchType]; !ok {
			return nil, fmt.Errorf("unsupported Comparator: %s", matchType)
		}

		f := NewValueFilter(NewCompareFilter(compareOpMap[op], matchTypeMap[matchType](value)))
		s := p.addFilter(f)
		filterStr = strings.Replace(filterStr, match[0], s, 1)
	}

	// try reSingleColumnValueFilter
	for _, match := range reSingleColumnValueFilter.FindAllStringSubmatch(filterStr, -1) {
		cq := match[1]
		cf := match[2]
		op := match[3]
		matchType := match[4]
		value := match[5]
		var filterIfMissing, latestVersionOnly bool = false, true
		if len(match) >= 7 && match[6] != "" {
			filterIfMissing = strings.ToLower(match[6]) == "true"
			latestVersionOnly = strings.ToLower(match[7]) == "true"
		}

		if _, ok := compareOpMap[op]; !ok {
			return nil, fmt.Errorf("unsupported operator: %s", op)
		}

		if _, ok := matchTypeMap[matchType]; !ok {
			return nil, fmt.Errorf("unsupported Comparator: %s", matchType)
		}

		f := NewSingleColumnValueFilter([]byte(cq), []byte(cf),
			compareOpMap[op], matchTypeMap[matchType](value),
			filterIfMissing, latestVersionOnly,
		)
		s := p.addFilter(f)
		filterStr = strings.Replace(filterStr, match[0], s, 1)
	}

	// drop ( )
	for match := reFilterWrap.FindStringSubmatch(filterStr); match != nil; match = reFilterWrap.FindStringSubmatch(filterStr) {
		filterStr = strings.Replace(filterStr, match[0], match[1], 1)
	}

	// drop AND OR
	for match := reListFilters.FindStringSubmatch(filterStr); match != nil; match = reListFilters.FindStringSubmatch(filterStr) {
		var f string
		switch match[1] {
		case "AND":
			f = p.parseList(match[0], MustPassAll)
		case "OR":
			f = p.parseList(match[0], MustPassOne)
		}
		filterStr = strings.Replace(filterStr, match[0], f, 1)

		// drop ( ) again
		if match := reFilterWrap.FindStringSubmatch(filterStr); match != nil {
			filterStr = strings.Replace(filterStr, match[0], match[1], 1)
		}
	}

	if match := reFilter.FindStringSubmatch(filterStr); match != nil {
		idx, _ := strconv.Atoi(match[1])
		return p.filters[idx], nil
	}

	return nil, errors.New("unable to Parse filter: " + filterStr)
}
