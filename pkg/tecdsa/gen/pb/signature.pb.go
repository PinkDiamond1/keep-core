// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.7.1
// source: pkg/tecdsa/gen/pb/signature.proto

package pb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Signature struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	R          []byte `protobuf:"bytes,1,opt,name=r,proto3" json:"r,omitempty"`
	S          []byte `protobuf:"bytes,2,opt,name=s,proto3" json:"s,omitempty"`
	RecoveryID int32  `protobuf:"varint,3,opt,name=recoveryID,proto3" json:"recoveryID,omitempty"`
}

func (x *Signature) Reset() {
	*x = Signature{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_tecdsa_gen_pb_signature_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Signature) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Signature) ProtoMessage() {}

func (x *Signature) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_tecdsa_gen_pb_signature_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Signature.ProtoReflect.Descriptor instead.
func (*Signature) Descriptor() ([]byte, []int) {
	return file_pkg_tecdsa_gen_pb_signature_proto_rawDescGZIP(), []int{0}
}

func (x *Signature) GetR() []byte {
	if x != nil {
		return x.R
	}
	return nil
}

func (x *Signature) GetS() []byte {
	if x != nil {
		return x.S
	}
	return nil
}

func (x *Signature) GetRecoveryID() int32 {
	if x != nil {
		return x.RecoveryID
	}
	return 0
}

var File_pkg_tecdsa_gen_pb_signature_proto protoreflect.FileDescriptor

var file_pkg_tecdsa_gen_pb_signature_proto_rawDesc = []byte{
	0x0a, 0x21, 0x70, 0x6b, 0x67, 0x2f, 0x74, 0x65, 0x63, 0x64, 0x73, 0x61, 0x2f, 0x67, 0x65, 0x6e,
	0x2f, 0x70, 0x62, 0x2f, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x06, 0x74, 0x65, 0x63, 0x64, 0x73, 0x61, 0x22, 0x47, 0x0a, 0x09, 0x53,
	0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x12, 0x0c, 0x0a, 0x01, 0x72, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x01, 0x72, 0x12, 0x0c, 0x0a, 0x01, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0c, 0x52, 0x01, 0x73, 0x12, 0x1e, 0x0a, 0x0a, 0x72, 0x65, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x79,
	0x49, 0x44, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0a, 0x72, 0x65, 0x63, 0x6f, 0x76, 0x65,
	0x72, 0x79, 0x49, 0x44, 0x42, 0x06, 0x5a, 0x04, 0x2e, 0x2f, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_pkg_tecdsa_gen_pb_signature_proto_rawDescOnce sync.Once
	file_pkg_tecdsa_gen_pb_signature_proto_rawDescData = file_pkg_tecdsa_gen_pb_signature_proto_rawDesc
)

func file_pkg_tecdsa_gen_pb_signature_proto_rawDescGZIP() []byte {
	file_pkg_tecdsa_gen_pb_signature_proto_rawDescOnce.Do(func() {
		file_pkg_tecdsa_gen_pb_signature_proto_rawDescData = protoimpl.X.CompressGZIP(file_pkg_tecdsa_gen_pb_signature_proto_rawDescData)
	})
	return file_pkg_tecdsa_gen_pb_signature_proto_rawDescData
}

var file_pkg_tecdsa_gen_pb_signature_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_pkg_tecdsa_gen_pb_signature_proto_goTypes = []interface{}{
	(*Signature)(nil), // 0: tecdsa.Signature
}
var file_pkg_tecdsa_gen_pb_signature_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_pkg_tecdsa_gen_pb_signature_proto_init() }
func file_pkg_tecdsa_gen_pb_signature_proto_init() {
	if File_pkg_tecdsa_gen_pb_signature_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_pkg_tecdsa_gen_pb_signature_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Signature); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_pkg_tecdsa_gen_pb_signature_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_pkg_tecdsa_gen_pb_signature_proto_goTypes,
		DependencyIndexes: file_pkg_tecdsa_gen_pb_signature_proto_depIdxs,
		MessageInfos:      file_pkg_tecdsa_gen_pb_signature_proto_msgTypes,
	}.Build()
	File_pkg_tecdsa_gen_pb_signature_proto = out.File
	file_pkg_tecdsa_gen_pb_signature_proto_rawDesc = nil
	file_pkg_tecdsa_gen_pb_signature_proto_goTypes = nil
	file_pkg_tecdsa_gen_pb_signature_proto_depIdxs = nil
}