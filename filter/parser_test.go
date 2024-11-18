package filter

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parserReg(t *testing.T) {
	tests := []struct {
		reg   *regexp.Regexp
		input string
		want  bool
	}{
		{
			reg:   reRowPrefix,
			input: "PrefixFilter( 'age' )",
			want:  true,
		},
		{
			reg:   reValueFilter,
			input: "ValueFilter( = , 'substring:18' )",
			want:  true,
		},
		{
			reg:   reValueFilter,
			input: "ValueFilter(=,'substring:18')",
			want:  true,
		},
		{
			reg:   reSingleColumnValueFilter,
			input: "SingleColumnValueFilter('cf1', 'col1', =, 'binary:14')",
			want:  true,
		},
		{
			reg:   reSingleColumnValueFilter,
			input: "SingleColumnValueFilter('cf1', 'col1', =, 'binary:14', true, true)",
			want:  true,
		},
		{
			reg:   reSingleColumnValueFilter,
			input: "SingleColumnValueFilter('cf1', 'col1', =, 'binary:14',True, FALSE)",
			want:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := tt.reg.MatchString(tt.input)
			assert.Equalf(t, tt.want, got, "input:%s want:%t", tt.input, tt.want)
		})
	}
}

func Test_parserFilter(t *testing.T) {
	tests := []struct {
		name      string
		filterStr string
		want      Filter
		wantErr   bool
	}{
		{
			name:      "PrefixFilter('age')",
			filterStr: "PrefixFilter",
			want:      nil,
			wantErr:   true,
		},
		{
			name:      "PrefixFilter('age')",
			filterStr: "PrefixFilter('age')",
			want:      NewPrefixFilter([]byte("age")),
			wantErr:   false,
		},
		{
			name:      "ValueFilter(=,'substring:18')",
			filterStr: "ValueFilter(=,'substring:18')",
			want:      NewValueFilter(NewCompareFilter(Equal, NewSubstringComparator("18"))),
			wantErr:   false,
		},
		{
			name:      "PrefixFilter('age') AND ValueFilter(=,'substring:18')",
			filterStr: "PrefixFilter('age') AND ValueFilter(=,'substring:18')",
			want: NewList(MustPassAll,
				NewPrefixFilter([]byte("age")),
				NewValueFilter(NewCompareFilter(Equal, NewSubstringComparator("18"))),
			),
			wantErr: false,
		},
		{
			name:      "PrefixFilter('age') OR ValueFilter(=,'substring:18')",
			filterStr: "PrefixFilter('age') OR ValueFilter(=,'substring:18')",
			want: NewList(MustPassOne,
				NewPrefixFilter([]byte("age")),
				NewValueFilter(NewCompareFilter(Equal, NewSubstringComparator("18"))),
			),
			wantErr: false,
		},
		{
			name:      "PrefixFilter('age') AND ( PrefixFilter('age') OR ValueFilter(=,'substring:18') )",
			filterStr: "PrefixFilter('age') AND ( PrefixFilter('age') OR ValueFilter(=,'substring:18') )",
			want: NewList(MustPassAll,
				NewPrefixFilter([]byte("age")),
				NewList(MustPassOne,
					NewPrefixFilter([]byte("age")),
					NewValueFilter(NewCompareFilter(Equal, NewSubstringComparator("18"))),
				)),
			wantErr: false,
		},
		{
			name:      "PrefixFilter('age') OR ( PrefixFilter('age') AND ValueFilter(=,'substring:18') )",
			filterStr: "PrefixFilter('age') OR ( PrefixFilter('age') AND ValueFilter(=,'substring:18') )",
			want: NewList(MustPassOne,
				NewPrefixFilter([]byte("age")),
				NewList(MustPassAll,
					NewPrefixFilter([]byte("age")),
					NewValueFilter(NewCompareFilter(Equal, NewSubstringComparator("18"))),
				)),
			wantErr: false,
		},
		{
			name:      "PrefixFilter('age') OR ( PrefixFilter('age') AND ValueFilter(=,'substring:18') AND ValueFilter(=,'substring:18') )",
			filterStr: "PrefixFilter('age') OR ( PrefixFilter('age') AND ValueFilter(=,'substring:18') AND ValueFilter(=,'substring:18') )",
			want: NewList(MustPassOne,
				NewPrefixFilter([]byte("age")),
				NewList(MustPassAll,
					NewPrefixFilter([]byte("age")),
					NewValueFilter(NewCompareFilter(Equal, NewSubstringComparator("18"))),
					NewValueFilter(NewCompareFilter(Equal, NewSubstringComparator("18"))),
				)),
			wantErr: false,
		},
		{
			name:      "( PrefixFilter('age') OR ValueFilter(=,'substring:18') ) AND ( PrefixFilter('age') OR ValueFilter(=,'substring:18') )",
			filterStr: "( PrefixFilter('age') OR ValueFilter(=,'substring:18') ) AND ( PrefixFilter('age') OR ValueFilter(=,'substring:18') )",
			want: NewList(MustPassAll,
				NewList(MustPassOne,
					NewPrefixFilter([]byte("age")),
					NewValueFilter(NewCompareFilter(Equal, NewSubstringComparator("18"))),
				),
				NewList(MustPassOne,
					NewPrefixFilter([]byte("age")),
					NewValueFilter(NewCompareFilter(Equal, NewSubstringComparator("18"))),
				)),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p Parser
			got, err := p.Parse(tt.filterStr)
			if tt.wantErr && err == nil {
				t.Errorf("parseFilter(%v), wantErr: %v, got: %v", tt.filterStr, tt.wantErr, got)
				return
			}
			assert.Equalf(t, tt.want, got, "parseFilter(%v)", tt.filterStr)
		})
	}
}
