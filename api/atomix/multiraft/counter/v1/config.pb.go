// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: atomix/multiraft/counter/v1/config.proto

package v1

import (
	fmt "fmt"
	proto "github.com/gogo/protobuf/proto"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type CounterConfig struct {
}

func (m *CounterConfig) Reset()         { *m = CounterConfig{} }
func (m *CounterConfig) String() string { return proto.CompactTextString(m) }
func (*CounterConfig) ProtoMessage()    {}
func (*CounterConfig) Descriptor() ([]byte, []int) {
	return fileDescriptor_59331e77389511cc, []int{0}
}
func (m *CounterConfig) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *CounterConfig) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_CounterConfig.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *CounterConfig) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CounterConfig.Merge(m, src)
}
func (m *CounterConfig) XXX_Size() int {
	return m.Size()
}
func (m *CounterConfig) XXX_DiscardUnknown() {
	xxx_messageInfo_CounterConfig.DiscardUnknown(m)
}

var xxx_messageInfo_CounterConfig proto.InternalMessageInfo

func init() {
	proto.RegisterType((*CounterConfig)(nil), "atomix.multiraft.counter.v1.CounterConfig")
}

func init() {
	proto.RegisterFile("atomix/multiraft/counter/v1/config.proto", fileDescriptor_59331e77389511cc)
}

var fileDescriptor_59331e77389511cc = []byte{
	// 121 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xd2, 0x48, 0x2c, 0xc9, 0xcf,
	0xcd, 0xac, 0xd0, 0xcf, 0x2d, 0xcd, 0x29, 0xc9, 0x2c, 0x4a, 0x4c, 0x2b, 0xd1, 0x4f, 0xce, 0x2f,
	0xcd, 0x2b, 0x49, 0x2d, 0xd2, 0x2f, 0x33, 0xd4, 0x4f, 0xce, 0xcf, 0x4b, 0xcb, 0x4c, 0xd7, 0x2b,
	0x28, 0xca, 0x2f, 0xc9, 0x17, 0x92, 0x86, 0xa8, 0xd4, 0x83, 0xab, 0xd4, 0x83, 0xaa, 0xd4, 0x2b,
	0x33, 0x54, 0xe2, 0xe7, 0xe2, 0x75, 0x86, 0xf0, 0x9c, 0xc1, 0x7a, 0x9c, 0x24, 0x4e, 0x3c, 0x92,
	0x63, 0xbc, 0xf0, 0x48, 0x8e, 0xf1, 0xc1, 0x23, 0x39, 0xc6, 0x09, 0x8f, 0xe5, 0x18, 0x2e, 0x3c,
	0x96, 0x63, 0xb8, 0xf1, 0x58, 0x8e, 0x21, 0x89, 0x0d, 0x6c, 0x9c, 0x31, 0x20, 0x00, 0x00, 0xff,
	0xff, 0x3e, 0x06, 0x7a, 0xc2, 0x7a, 0x00, 0x00, 0x00,
}

func (m *CounterConfig) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *CounterConfig) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *CounterConfig) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func encodeVarintConfig(dAtA []byte, offset int, v uint64) int {
	offset -= sovConfig(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *CounterConfig) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func sovConfig(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozConfig(x uint64) (n int) {
	return sovConfig(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *CounterConfig) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowConfig
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
			return fmt.Errorf("proto: CounterConfig: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: CounterConfig: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipConfig(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthConfig
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
func skipConfig(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowConfig
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
					return 0, ErrIntOverflowConfig
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
					return 0, ErrIntOverflowConfig
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
				return 0, ErrInvalidLengthConfig
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupConfig
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthConfig
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthConfig        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowConfig          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupConfig = fmt.Errorf("proto: unexpected end of group")
)
