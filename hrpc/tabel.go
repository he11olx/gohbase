package hrpc

import (
	"context"

	"github.com/tsuna/gohbase/pb"
	"google.golang.org/protobuf/proto"
)

// GetTableDescriptor models a GetTableDescriptor pb call
type GetTableDescriptor struct {
	base
	tableName string
	namespace string

	regex            string
	includeSysTables bool
}

func (tn *GetTableDescriptor) Description() string {
	return tn.Name()
}

func NewGetTableDescriptor(ctx context.Context, namespace, tableName string) (*GetTableDescriptor, error) {
	tn := &GetTableDescriptor{
		base: base{
			ctx:      ctx,
			resultch: make(chan RPCResult, 1),
		},
		tableName:        tableName,
		namespace:        namespace,
		regex:            ".*",
		includeSysTables: false,
	}
	return tn, nil
}

// Name returns the name of this RPC call.
func (tn *GetTableDescriptor) Name() string {
	return "GetTableDescriptors"
}

// ToProto converts the RPC into a protobuf message.
func (tn *GetTableDescriptor) ToProto() proto.Message {
	return &pb.GetTableDescriptorsRequest{
		TableNames: []*pb.TableName{
			{
				Namespace: []byte(tn.namespace),
				Qualifier: []byte(tn.tableName),
			},
		},
		Regex:            proto.String(tn.regex),
		IncludeSysTables: proto.Bool(tn.includeSysTables),
		Namespace:        proto.String(tn.namespace),
	}
}

// NewResponse creates an empty protobuf message to read the response of this
// RPC.
func (tn *GetTableDescriptor) NewResponse() proto.Message {
	return &pb.GetTableDescriptorsResponse{}
}
