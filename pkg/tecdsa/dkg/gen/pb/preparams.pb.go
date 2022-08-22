// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: pb/preparams.proto

package pb

import (
	bytes "bytes"
	fmt "fmt"
	io "io"
	math "math"
	math_bits "math/bits"
	reflect "reflect"
	strings "strings"

	proto "github.com/gogo/protobuf/proto"
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

type PreParams struct {
	Data *PreParams_LocalPreParams `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (m *PreParams) Reset()      { *m = PreParams{} }
func (*PreParams) ProtoMessage() {}
func (*PreParams) Descriptor() ([]byte, []int) {
	return fileDescriptor_fd924d802897c3eb, []int{0}
}
func (m *PreParams) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *PreParams) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_PreParams.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *PreParams) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PreParams.Merge(m, src)
}
func (m *PreParams) XXX_Size() int {
	return m.Size()
}
func (m *PreParams) XXX_DiscardUnknown() {
	xxx_messageInfo_PreParams.DiscardUnknown(m)
}

var xxx_messageInfo_PreParams proto.InternalMessageInfo

func (m *PreParams) GetData() *PreParams_LocalPreParams {
	if m != nil {
		return m.Data
	}
	return nil
}

type PreParams_LocalPreParams struct {
	NTilde []byte `protobuf:"bytes,2,opt,name=nTilde,proto3" json:"nTilde,omitempty"`
	H1I    []byte `protobuf:"bytes,3,opt,name=h1i,proto3" json:"h1i,omitempty"`
	H2I    []byte `protobuf:"bytes,4,opt,name=h2i,proto3" json:"h2i,omitempty"`
	Alpha  []byte `protobuf:"bytes,5,opt,name=alpha,proto3" json:"alpha,omitempty"`
	Beta   []byte `protobuf:"bytes,6,opt,name=beta,proto3" json:"beta,omitempty"`
	P      []byte `protobuf:"bytes,7,opt,name=p,proto3" json:"p,omitempty"`
	Q      []byte `protobuf:"bytes,8,opt,name=q,proto3" json:"q,omitempty"`
}

func (m *PreParams_LocalPreParams) Reset()      { *m = PreParams_LocalPreParams{} }
func (*PreParams_LocalPreParams) ProtoMessage() {}
func (*PreParams_LocalPreParams) Descriptor() ([]byte, []int) {
	return fileDescriptor_fd924d802897c3eb, []int{0, 0}
}
func (m *PreParams_LocalPreParams) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *PreParams_LocalPreParams) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_PreParams_LocalPreParams.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *PreParams_LocalPreParams) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PreParams_LocalPreParams.Merge(m, src)
}
func (m *PreParams_LocalPreParams) XXX_Size() int {
	return m.Size()
}
func (m *PreParams_LocalPreParams) XXX_DiscardUnknown() {
	xxx_messageInfo_PreParams_LocalPreParams.DiscardUnknown(m)
}

var xxx_messageInfo_PreParams_LocalPreParams proto.InternalMessageInfo

func (m *PreParams_LocalPreParams) GetNTilde() []byte {
	if m != nil {
		return m.NTilde
	}
	return nil
}

func (m *PreParams_LocalPreParams) GetH1I() []byte {
	if m != nil {
		return m.H1I
	}
	return nil
}

func (m *PreParams_LocalPreParams) GetH2I() []byte {
	if m != nil {
		return m.H2I
	}
	return nil
}

func (m *PreParams_LocalPreParams) GetAlpha() []byte {
	if m != nil {
		return m.Alpha
	}
	return nil
}

func (m *PreParams_LocalPreParams) GetBeta() []byte {
	if m != nil {
		return m.Beta
	}
	return nil
}

func (m *PreParams_LocalPreParams) GetP() []byte {
	if m != nil {
		return m.P
	}
	return nil
}

func (m *PreParams_LocalPreParams) GetQ() []byte {
	if m != nil {
		return m.Q
	}
	return nil
}

func init() {
	proto.RegisterType((*PreParams)(nil), "dkg.PreParams")
	proto.RegisterType((*PreParams_LocalPreParams)(nil), "dkg.PreParams.LocalPreParams")
}

func init() { proto.RegisterFile("pb/preparams.proto", fileDescriptor_fd924d802897c3eb) }

var fileDescriptor_fd924d802897c3eb = []byte{
	// 252 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x54, 0x90, 0xb1, 0x4a, 0xc3, 0x50,
	0x14, 0x86, 0xef, 0x69, 0xd2, 0xa8, 0xc7, 0x22, 0x72, 0x10, 0xb9, 0x08, 0x1e, 0x8a, 0x53, 0xa7,
	0x48, 0xe3, 0xe2, 0xec, 0xec, 0x50, 0x8a, 0x93, 0xdb, 0x8d, 0x09, 0x6d, 0x68, 0x34, 0xb7, 0xd7,
	0x3c, 0x80, 0x8f, 0x20, 0x3e, 0x85, 0x8f, 0xe2, 0x18, 0x70, 0xe9, 0x68, 0x6e, 0x16, 0xc7, 0x3e,
	0x82, 0xf4, 0xb4, 0x14, 0xdc, 0xfe, 0xef, 0xfb, 0xce, 0x74, 0x90, 0x6c, 0x7a, 0x6d, 0x5d, 0x6e,
	0x8d, 0x33, 0xcf, 0xaf, 0xb1, 0x75, 0x55, 0x5d, 0x51, 0x90, 0x2d, 0x66, 0x57, 0xdf, 0x80, 0x47,
	0x13, 0x97, 0x4f, 0x24, 0xd0, 0x18, 0xc3, 0xcc, 0xd4, 0x46, 0xc3, 0x10, 0x46, 0xc7, 0xc9, 0x65,
	0x9c, 0x2d, 0x66, 0xf1, 0xbe, 0xc6, 0xf7, 0xd5, 0x93, 0x29, 0xf7, 0x38, 0x95, 0xd3, 0x8b, 0x0f,
	0xc0, 0x93, 0xff, 0x81, 0xce, 0x31, 0x7a, 0x79, 0x28, 0xca, 0x2c, 0xd7, 0xbd, 0x21, 0x8c, 0x06,
	0xd3, 0x1d, 0xd1, 0x29, 0x06, 0xf3, 0x71, 0xa1, 0x03, 0x91, 0x9b, 0x29, 0x26, 0x29, 0x74, 0xb8,
	0x33, 0x49, 0x41, 0x67, 0xd8, 0x37, 0xa5, 0x9d, 0x1b, 0xdd, 0x17, 0xb7, 0x05, 0x22, 0x0c, 0xd3,
	0xbc, 0x36, 0x3a, 0x12, 0x29, 0x9b, 0x06, 0x08, 0x56, 0x1f, 0x88, 0x00, 0xbb, 0xa1, 0xa5, 0x3e,
	0xdc, 0xd2, 0xf2, 0xee, 0xb6, 0x69, 0x59, 0xad, 0x5a, 0x56, 0xeb, 0x96, 0xe1, 0xcd, 0x33, 0x7c,
	0x7a, 0x86, 0x2f, 0xcf, 0xd0, 0x78, 0x86, 0x1f, 0xcf, 0xf0, 0xeb, 0x59, 0xad, 0x3d, 0xc3, 0x7b,
	0xc7, 0xaa, 0xe9, 0x58, 0xad, 0x3a, 0x56, 0x8f, 0x3d, 0x9b, 0xa6, 0x91, 0xfc, 0xe6, 0xe6, 0x2f,
	0x00, 0x00, 0xff, 0xff, 0xe8, 0x5d, 0xb0, 0xb3, 0x31, 0x01, 0x00, 0x00,
}

func (this *PreParams) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*PreParams)
	if !ok {
		that2, ok := that.(PreParams)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if !this.Data.Equal(that1.Data) {
		return false
	}
	return true
}
func (this *PreParams_LocalPreParams) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*PreParams_LocalPreParams)
	if !ok {
		that2, ok := that.(PreParams_LocalPreParams)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if !bytes.Equal(this.NTilde, that1.NTilde) {
		return false
	}
	if !bytes.Equal(this.H1I, that1.H1I) {
		return false
	}
	if !bytes.Equal(this.H2I, that1.H2I) {
		return false
	}
	if !bytes.Equal(this.Alpha, that1.Alpha) {
		return false
	}
	if !bytes.Equal(this.Beta, that1.Beta) {
		return false
	}
	if !bytes.Equal(this.P, that1.P) {
		return false
	}
	if !bytes.Equal(this.Q, that1.Q) {
		return false
	}
	return true
}
func (this *PreParams) GoString() string {
	if this == nil {
		return "nil"
	}
	s := make([]string, 0, 5)
	s = append(s, "&pb.PreParams{")
	if this.Data != nil {
		s = append(s, "Data: "+fmt.Sprintf("%#v", this.Data)+",\n")
	}
	s = append(s, "}")
	return strings.Join(s, "")
}
func (this *PreParams_LocalPreParams) GoString() string {
	if this == nil {
		return "nil"
	}
	s := make([]string, 0, 11)
	s = append(s, "&pb.PreParams_LocalPreParams{")
	s = append(s, "NTilde: "+fmt.Sprintf("%#v", this.NTilde)+",\n")
	s = append(s, "H1I: "+fmt.Sprintf("%#v", this.H1I)+",\n")
	s = append(s, "H2I: "+fmt.Sprintf("%#v", this.H2I)+",\n")
	s = append(s, "Alpha: "+fmt.Sprintf("%#v", this.Alpha)+",\n")
	s = append(s, "Beta: "+fmt.Sprintf("%#v", this.Beta)+",\n")
	s = append(s, "P: "+fmt.Sprintf("%#v", this.P)+",\n")
	s = append(s, "Q: "+fmt.Sprintf("%#v", this.Q)+",\n")
	s = append(s, "}")
	return strings.Join(s, "")
}
func valueToGoStringPreparams(v interface{}, typ string) string {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect.Indirect(rv).Interface()
	return fmt.Sprintf("func(v %v) *%v { return &v } ( %#v )", typ, typ, pv)
}
func (m *PreParams) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *PreParams) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *PreParams) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Data != nil {
		{
			size, err := m.Data.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintPreparams(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *PreParams_LocalPreParams) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *PreParams_LocalPreParams) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *PreParams_LocalPreParams) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Q) > 0 {
		i -= len(m.Q)
		copy(dAtA[i:], m.Q)
		i = encodeVarintPreparams(dAtA, i, uint64(len(m.Q)))
		i--
		dAtA[i] = 0x42
	}
	if len(m.P) > 0 {
		i -= len(m.P)
		copy(dAtA[i:], m.P)
		i = encodeVarintPreparams(dAtA, i, uint64(len(m.P)))
		i--
		dAtA[i] = 0x3a
	}
	if len(m.Beta) > 0 {
		i -= len(m.Beta)
		copy(dAtA[i:], m.Beta)
		i = encodeVarintPreparams(dAtA, i, uint64(len(m.Beta)))
		i--
		dAtA[i] = 0x32
	}
	if len(m.Alpha) > 0 {
		i -= len(m.Alpha)
		copy(dAtA[i:], m.Alpha)
		i = encodeVarintPreparams(dAtA, i, uint64(len(m.Alpha)))
		i--
		dAtA[i] = 0x2a
	}
	if len(m.H2I) > 0 {
		i -= len(m.H2I)
		copy(dAtA[i:], m.H2I)
		i = encodeVarintPreparams(dAtA, i, uint64(len(m.H2I)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.H1I) > 0 {
		i -= len(m.H1I)
		copy(dAtA[i:], m.H1I)
		i = encodeVarintPreparams(dAtA, i, uint64(len(m.H1I)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.NTilde) > 0 {
		i -= len(m.NTilde)
		copy(dAtA[i:], m.NTilde)
		i = encodeVarintPreparams(dAtA, i, uint64(len(m.NTilde)))
		i--
		dAtA[i] = 0x12
	}
	return len(dAtA) - i, nil
}

func encodeVarintPreparams(dAtA []byte, offset int, v uint64) int {
	offset -= sovPreparams(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *PreParams) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Data != nil {
		l = m.Data.Size()
		n += 1 + l + sovPreparams(uint64(l))
	}
	return n
}

func (m *PreParams_LocalPreParams) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.NTilde)
	if l > 0 {
		n += 1 + l + sovPreparams(uint64(l))
	}
	l = len(m.H1I)
	if l > 0 {
		n += 1 + l + sovPreparams(uint64(l))
	}
	l = len(m.H2I)
	if l > 0 {
		n += 1 + l + sovPreparams(uint64(l))
	}
	l = len(m.Alpha)
	if l > 0 {
		n += 1 + l + sovPreparams(uint64(l))
	}
	l = len(m.Beta)
	if l > 0 {
		n += 1 + l + sovPreparams(uint64(l))
	}
	l = len(m.P)
	if l > 0 {
		n += 1 + l + sovPreparams(uint64(l))
	}
	l = len(m.Q)
	if l > 0 {
		n += 1 + l + sovPreparams(uint64(l))
	}
	return n
}

func sovPreparams(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozPreparams(x uint64) (n int) {
	return sovPreparams(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (this *PreParams) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&PreParams{`,
		`Data:` + strings.Replace(fmt.Sprintf("%v", this.Data), "PreParams_LocalPreParams", "PreParams_LocalPreParams", 1) + `,`,
		`}`,
	}, "")
	return s
}
func (this *PreParams_LocalPreParams) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&PreParams_LocalPreParams{`,
		`NTilde:` + fmt.Sprintf("%v", this.NTilde) + `,`,
		`H1I:` + fmt.Sprintf("%v", this.H1I) + `,`,
		`H2I:` + fmt.Sprintf("%v", this.H2I) + `,`,
		`Alpha:` + fmt.Sprintf("%v", this.Alpha) + `,`,
		`Beta:` + fmt.Sprintf("%v", this.Beta) + `,`,
		`P:` + fmt.Sprintf("%v", this.P) + `,`,
		`Q:` + fmt.Sprintf("%v", this.Q) + `,`,
		`}`,
	}, "")
	return s
}
func valueToStringPreparams(v interface{}) string {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect.Indirect(rv).Interface()
	return fmt.Sprintf("*%v", pv)
}
func (m *PreParams) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowPreparams
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
			return fmt.Errorf("proto: PreParams: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: PreParams: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Data", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPreparams
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
				return ErrInvalidLengthPreparams
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthPreparams
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Data == nil {
				m.Data = &PreParams_LocalPreParams{}
			}
			if err := m.Data.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipPreparams(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthPreparams
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
func (m *PreParams_LocalPreParams) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowPreparams
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
			return fmt.Errorf("proto: LocalPreParams: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: LocalPreParams: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field NTilde", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPreparams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthPreparams
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthPreparams
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.NTilde = append(m.NTilde[:0], dAtA[iNdEx:postIndex]...)
			if m.NTilde == nil {
				m.NTilde = []byte{}
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field H1I", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPreparams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthPreparams
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthPreparams
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.H1I = append(m.H1I[:0], dAtA[iNdEx:postIndex]...)
			if m.H1I == nil {
				m.H1I = []byte{}
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field H2I", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPreparams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthPreparams
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthPreparams
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.H2I = append(m.H2I[:0], dAtA[iNdEx:postIndex]...)
			if m.H2I == nil {
				m.H2I = []byte{}
			}
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Alpha", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPreparams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthPreparams
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthPreparams
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Alpha = append(m.Alpha[:0], dAtA[iNdEx:postIndex]...)
			if m.Alpha == nil {
				m.Alpha = []byte{}
			}
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Beta", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPreparams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthPreparams
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthPreparams
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Beta = append(m.Beta[:0], dAtA[iNdEx:postIndex]...)
			if m.Beta == nil {
				m.Beta = []byte{}
			}
			iNdEx = postIndex
		case 7:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field P", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPreparams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthPreparams
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthPreparams
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.P = append(m.P[:0], dAtA[iNdEx:postIndex]...)
			if m.P == nil {
				m.P = []byte{}
			}
			iNdEx = postIndex
		case 8:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Q", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPreparams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthPreparams
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthPreparams
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Q = append(m.Q[:0], dAtA[iNdEx:postIndex]...)
			if m.Q == nil {
				m.Q = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipPreparams(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthPreparams
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
func skipPreparams(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowPreparams
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
					return 0, ErrIntOverflowPreparams
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
					return 0, ErrIntOverflowPreparams
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
				return 0, ErrInvalidLengthPreparams
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupPreparams
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthPreparams
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthPreparams        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowPreparams          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupPreparams = fmt.Errorf("proto: unexpected end of group")
)