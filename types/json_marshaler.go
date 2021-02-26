package types

import (
	"bytes"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
)

var _ codec.JSONMarshaler = (*JsonMarshaller)(nil)

// JsonMarshaller that omits empty values
type JsonMarshaller struct {
	interfaceRegistry types.InterfaceRegistry
}

func NewMarshaller(ctx client.Context) *JsonMarshaller {
	return &JsonMarshaller{interfaceRegistry: ctx.InterfaceRegistry}
}

// MarshalJSON implements JSONMarshaler.MarshalJSON method,
// it marshals to JSON using proto codec.
func (pc *JsonMarshaller) MarshalJSON(o proto.Message) ([]byte, error) {
	m, ok := o.(codec.ProtoMarshaler)
	if !ok {
		return nil, fmt.Errorf("cannot protobuf JSON encode unsupported type: %T", o)
	}

	return ProtoMarshalJSON(m, pc.interfaceRegistry)
}

// MustMarshalJSON implements JSONMarshaler.MustMarshalJSON method,
// it executes MarshalJSON except it panics upon failure.
func (pc *JsonMarshaller) MustMarshalJSON(o proto.Message) []byte {
	bz, err := pc.MarshalJSON(o)
	if err != nil {
		panic(err)
	}

	return bz
}

func (j JsonMarshaller) MarshalInterfaceJSON(i proto.Message) ([]byte, error) {
	panic("not implemented")
}

func (j JsonMarshaller) UnmarshalInterfaceJSON(bz []byte, ptr interface{}) error {
	panic("not implemented")
}

func (j JsonMarshaller) UnmarshalJSON(bz []byte, ptr proto.Message) error {
	panic("not implemented")
}

func (j JsonMarshaller) MustUnmarshalJSON(bz []byte, ptr proto.Message) {
	panic("not implemented")
}

// ProtoMarshalJSON provides an auxiliary function to return Proto3 JSON encoded
// bytes of a message.
func ProtoMarshalJSON(msg proto.Message, resolver jsonpb.AnyResolver) ([]byte, error) {
	// copied from sdk with EmitDefaults: false

	// We use the OrigName because camel casing fields just doesn't make sense.
	// EmitDefaults is also often the more expected behavior for CLI users
	jm := &jsonpb.Marshaler{OrigName: true, EmitDefaults: false, AnyResolver: resolver}
	err := types.UnpackInterfaces(msg, types.ProtoJSONPacker{JSONPBMarshaler: jm})
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)

	if err := jm.Marshal(buf, msg); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
