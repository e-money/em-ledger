// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: em/inflation/v1beta1/inflation.proto

package types

import (
	fmt "fmt"
	_ "github.com/cosmos/cosmos-sdk/types"
	github_com_cosmos_cosmos_sdk_types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	github_com_gogo_protobuf_types "github.com/gogo/protobuf/types"
	_ "github.com/golang/protobuf/ptypes/timestamp"
	io "io"
	math "math"
	math_bits "math/bits"
	time "time"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf
var _ = time.Kitchen

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type InflationAsset struct {
	Denom     string                                 `protobuf:"bytes,1,opt,name=denom,proto3" json:"denom,omitempty" yaml:"denom"`
	Inflation github_com_cosmos_cosmos_sdk_types.Dec `protobuf:"bytes,2,opt,name=inflation,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Dec" json:"inflation" yaml:"inflation"`
	Accum     github_com_cosmos_cosmos_sdk_types.Dec `protobuf:"bytes,3,opt,name=accum,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Dec" json:"accum" yaml:"accum"`
}

func (m *InflationAsset) Reset()         { *m = InflationAsset{} }
func (m *InflationAsset) String() string { return proto.CompactTextString(m) }
func (*InflationAsset) ProtoMessage()    {}
func (*InflationAsset) Descriptor() ([]byte, []int) {
	return fileDescriptor_7e4db7953b5ef153, []int{0}
}
func (m *InflationAsset) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *InflationAsset) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_InflationAsset.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *InflationAsset) XXX_Merge(src proto.Message) {
	xxx_messageInfo_InflationAsset.Merge(m, src)
}
func (m *InflationAsset) XXX_Size() int {
	return m.Size()
}
func (m *InflationAsset) XXX_DiscardUnknown() {
	xxx_messageInfo_InflationAsset.DiscardUnknown(m)
}

var xxx_messageInfo_InflationAsset proto.InternalMessageInfo

func (m *InflationAsset) GetDenom() string {
	if m != nil {
		return m.Denom
	}
	return ""
}

type InflationState struct {
	LastAppliedTime   time.Time                              `protobuf:"bytes,1,opt,name=last_applied_time,json=lastAppliedTime,proto3,stdtime" json:"last_applied_time" yaml:"last_applied"`
	LastAppliedHeight github_com_cosmos_cosmos_sdk_types.Int `protobuf:"bytes,2,opt,name=last_applied_height,json=lastAppliedHeight,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Int" json:"last_applied_height" yaml:"last_applied_height"`
	InflationAssets   []InflationAsset                       `protobuf:"bytes,3,rep,name=inflation_assets,json=inflationAssets,proto3" json:"assets" yaml:"assets"`
}

func (m *InflationState) Reset()      { *m = InflationState{} }
func (*InflationState) ProtoMessage() {}
func (*InflationState) Descriptor() ([]byte, []int) {
	return fileDescriptor_7e4db7953b5ef153, []int{1}
}
func (m *InflationState) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *InflationState) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_InflationState.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *InflationState) XXX_Merge(src proto.Message) {
	xxx_messageInfo_InflationState.Merge(m, src)
}
func (m *InflationState) XXX_Size() int {
	return m.Size()
}
func (m *InflationState) XXX_DiscardUnknown() {
	xxx_messageInfo_InflationState.DiscardUnknown(m)
}

var xxx_messageInfo_InflationState proto.InternalMessageInfo

func (m *InflationState) GetLastAppliedTime() time.Time {
	if m != nil {
		return m.LastAppliedTime
	}
	return time.Time{}
}

func (m *InflationState) GetInflationAssets() []InflationAsset {
	if m != nil {
		return m.InflationAssets
	}
	return nil
}

func init() {
	proto.RegisterType((*InflationAsset)(nil), "em.inflation.v1beta1.InflationAsset")
	proto.RegisterType((*InflationState)(nil), "em.inflation.v1beta1.InflationState")
}

func init() {
	proto.RegisterFile("em/inflation/v1beta1/inflation.proto", fileDescriptor_7e4db7953b5ef153)
}

var fileDescriptor_7e4db7953b5ef153 = []byte{
	// 475 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x93, 0x31, 0x6f, 0xd3, 0x40,
	0x14, 0xc7, 0xed, 0x96, 0x56, 0xaa, 0x0b, 0x34, 0x75, 0x3b, 0x44, 0x1e, 0x7c, 0xd5, 0xa9, 0xaa,
	0xba, 0xe4, 0x4e, 0x2d, 0x03, 0x52, 0x07, 0xa4, 0x46, 0x0c, 0x54, 0x62, 0xc1, 0x74, 0x62, 0x09,
	0x67, 0xe7, 0xd5, 0x39, 0xe1, 0xf3, 0x59, 0xb9, 0x0b, 0x22, 0x12, 0x1f, 0xa2, 0x23, 0x23, 0x1f,
	0xa7, 0x63, 0x47, 0xc4, 0x60, 0x50, 0xb2, 0xb1, 0x80, 0xfc, 0x09, 0x90, 0xef, 0x9c, 0xc4, 0x11,
	0x2c, 0x9d, 0xec, 0xfb, 0xfb, 0xdd, 0xef, 0xbd, 0xff, 0xff, 0xce, 0xde, 0x31, 0x08, 0xca, 0xf3,
	0x9b, 0x8c, 0x69, 0x2e, 0x73, 0xfa, 0xf1, 0x2c, 0x06, 0xcd, 0xce, 0x56, 0x0a, 0x29, 0xc6, 0x52,
	0x4b, 0xff, 0x10, 0x04, 0x59, 0x69, 0x4d, 0x55, 0x70, 0x98, 0xca, 0x54, 0x9a, 0x02, 0x5a, 0xbf,
	0xd9, 0xda, 0x20, 0x4c, 0xa4, 0x12, 0x52, 0xd1, 0x98, 0x29, 0x58, 0x02, 0x13, 0xc9, 0x1b, 0x56,
	0x80, 0x52, 0x29, 0xd3, 0x0c, 0xa8, 0x59, 0xc5, 0x93, 0x1b, 0xaa, 0xb9, 0x00, 0xa5, 0x99, 0x28,
	0x6c, 0x01, 0xfe, 0xe3, 0x7a, 0x4f, 0xaf, 0x16, 0xcd, 0x2e, 0x95, 0x02, 0xed, 0x9f, 0x78, 0x5b,
	0x43, 0xc8, 0xa5, 0xe8, 0xba, 0x47, 0xee, 0xe9, 0x4e, 0xbf, 0x53, 0x95, 0xe8, 0xf1, 0x94, 0x89,
	0xec, 0x02, 0x1b, 0x19, 0x47, 0xf6, 0xb3, 0xff, 0xde, 0xdb, 0x59, 0x8e, 0xd9, 0xdd, 0x30, 0xb5,
	0xfd, 0xbb, 0x12, 0x39, 0xdf, 0x4b, 0x74, 0x92, 0x72, 0x3d, 0x9a, 0xc4, 0x24, 0x91, 0x82, 0x36,
	0x13, 0xda, 0x47, 0x4f, 0x0d, 0x3f, 0x50, 0x3d, 0x2d, 0x40, 0x91, 0x97, 0x90, 0x54, 0x25, 0xea,
	0x58, 0xf2, 0x12, 0x84, 0xa3, 0x15, 0xd4, 0xbf, 0xf6, 0xb6, 0x58, 0x92, 0x4c, 0x44, 0x77, 0xd3,
	0xd0, 0x5f, 0x3c, 0x98, 0xde, 0xcc, 0x6d, 0x20, 0x38, 0xb2, 0x30, 0xfc, 0x7b, 0xa3, 0x65, 0xf9,
	0xad, 0x66, 0x1a, 0xfc, 0xd4, 0xdb, 0xcf, 0x98, 0xd2, 0x03, 0x56, 0x14, 0x19, 0x87, 0xe1, 0xa0,
	0x4e, 0xc9, 0xd8, 0xdf, 0x3d, 0x0f, 0x88, 0x8d, 0x90, 0x2c, 0x22, 0x24, 0xd7, 0x8b, 0x08, 0xfb,
	0xa8, 0x1e, 0xa8, 0x2a, 0xd1, 0x81, 0x6d, 0xd3, 0x46, 0xe0, 0xdb, 0x1f, 0xc8, 0x8d, 0xf6, 0x6a,
	0xe9, 0xd2, 0x2a, 0xf5, 0x36, 0xff, 0xb3, 0x77, 0xb0, 0xd6, 0x68, 0x04, 0x3c, 0x1d, 0xe9, 0x26,
	0xbd, 0xd7, 0x0f, 0xf0, 0x77, 0x95, 0xeb, 0xaa, 0x44, 0xc1, 0xbf, 0x8d, 0x1b, 0x24, 0x8e, 0xf6,
	0x5b, 0xbd, 0x5f, 0x19, 0xcd, 0x2f, 0xbc, 0xce, 0x32, 0xdc, 0x01, 0xab, 0x0f, 0x5b, 0x75, 0x37,
	0x8f, 0x36, 0x4f, 0x77, 0xcf, 0x8f, 0xc9, 0xff, 0x2e, 0x1d, 0x59, 0xbf, 0x19, 0xd6, 0xef, 0xaf,
	0x12, 0x6d, 0xdb, 0xbd, 0x55, 0x89, 0x9e, 0x34, 0x01, 0x9b, 0x35, 0x8e, 0xf6, 0xf8, 0xda, 0x06,
	0x75, 0xf1, 0xe8, 0xcb, 0x57, 0xe4, 0xf4, 0xdf, 0xdc, 0xcd, 0x42, 0xf7, 0x7e, 0x16, 0xba, 0x3f,
	0x67, 0xa1, 0x7b, 0x3b, 0x0f, 0x9d, 0xfb, 0x79, 0xe8, 0x7c, 0x9b, 0x87, 0xce, 0xbb, 0xe7, 0x2d,
	0xab, 0xd0, 0x13, 0x32, 0x87, 0x29, 0x05, 0xd1, 0xcb, 0x60, 0x98, 0xc2, 0x98, 0x7e, 0x6a, 0xfd,
	0x2d, 0x3c, 0xd7, 0x30, 0xce, 0x59, 0x66, 0xfd, 0xc7, 0xdb, 0xe6, 0x38, 0x9e, 0xfd, 0x0d, 0x00,
	0x00, 0xff, 0xff, 0x94, 0x17, 0xfd, 0x8f, 0x53, 0x03, 0x00, 0x00,
}

func (m *InflationAsset) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *InflationAsset) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *InflationAsset) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size := m.Accum.Size()
		i -= size
		if _, err := m.Accum.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintInflation(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x1a
	{
		size := m.Inflation.Size()
		i -= size
		if _, err := m.Inflation.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintInflation(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
	if len(m.Denom) > 0 {
		i -= len(m.Denom)
		copy(dAtA[i:], m.Denom)
		i = encodeVarintInflation(dAtA, i, uint64(len(m.Denom)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *InflationState) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *InflationState) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *InflationState) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.InflationAssets) > 0 {
		for iNdEx := len(m.InflationAssets) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.InflationAssets[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintInflation(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x1a
		}
	}
	{
		size := m.LastAppliedHeight.Size()
		i -= size
		if _, err := m.LastAppliedHeight.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintInflation(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
	n1, err1 := github_com_gogo_protobuf_types.StdTimeMarshalTo(m.LastAppliedTime, dAtA[i-github_com_gogo_protobuf_types.SizeOfStdTime(m.LastAppliedTime):])
	if err1 != nil {
		return 0, err1
	}
	i -= n1
	i = encodeVarintInflation(dAtA, i, uint64(n1))
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func encodeVarintInflation(dAtA []byte, offset int, v uint64) int {
	offset -= sovInflation(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *InflationAsset) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Denom)
	if l > 0 {
		n += 1 + l + sovInflation(uint64(l))
	}
	l = m.Inflation.Size()
	n += 1 + l + sovInflation(uint64(l))
	l = m.Accum.Size()
	n += 1 + l + sovInflation(uint64(l))
	return n
}

func (m *InflationState) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = github_com_gogo_protobuf_types.SizeOfStdTime(m.LastAppliedTime)
	n += 1 + l + sovInflation(uint64(l))
	l = m.LastAppliedHeight.Size()
	n += 1 + l + sovInflation(uint64(l))
	if len(m.InflationAssets) > 0 {
		for _, e := range m.InflationAssets {
			l = e.Size()
			n += 1 + l + sovInflation(uint64(l))
		}
	}
	return n
}

func sovInflation(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozInflation(x uint64) (n int) {
	return sovInflation(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *InflationAsset) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowInflation
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: InflationAsset: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: InflationAsset: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Denom", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowInflation
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthInflation
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthInflation
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Denom = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Inflation", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowInflation
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthInflation
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthInflation
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Inflation.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Accum", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowInflation
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthInflation
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthInflation
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Accum.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipInflation(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthInflation
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *InflationState) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowInflation
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: InflationState: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: InflationState: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field LastAppliedTime", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowInflation
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthInflation
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthInflation
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := github_com_gogo_protobuf_types.StdTimeUnmarshal(&m.LastAppliedTime, dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field LastAppliedHeight", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowInflation
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthInflation
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthInflation
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.LastAppliedHeight.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field InflationAssets", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowInflation
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthInflation
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthInflation
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.InflationAssets = append(m.InflationAssets, InflationAsset{})
			if err := m.InflationAssets[len(m.InflationAssets)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipInflation(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthInflation
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipInflation(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowInflation
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowInflation
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowInflation
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthInflation
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupInflation
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthInflation
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthInflation        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowInflation          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupInflation = fmt.Errorf("proto: unexpected end of group")
)