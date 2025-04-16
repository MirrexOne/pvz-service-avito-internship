// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v6.30.2
// source: pvz.proto

package pvz_v1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Enum ReceptionStatus не используется в GetPVZList, но может быть полезен для других методов
type ReceptionStatus int32

const (
	ReceptionStatus_RECEPTION_STATUS_UNSPECIFIED ReceptionStatus = 0 // Стандартное значение по умолчанию
	ReceptionStatus_RECEPTION_STATUS_IN_PROGRESS ReceptionStatus = 1
	ReceptionStatus_RECEPTION_STATUS_CLOSED      ReceptionStatus = 2
)

// Enum value maps for ReceptionStatus.
var (
	ReceptionStatus_name = map[int32]string{
		0: "RECEPTION_STATUS_UNSPECIFIED",
		1: "RECEPTION_STATUS_IN_PROGRESS",
		2: "RECEPTION_STATUS_CLOSED",
	}
	ReceptionStatus_value = map[string]int32{
		"RECEPTION_STATUS_UNSPECIFIED": 0,
		"RECEPTION_STATUS_IN_PROGRESS": 1,
		"RECEPTION_STATUS_CLOSED":      2,
	}
)

func (x ReceptionStatus) Enum() *ReceptionStatus {
	p := new(ReceptionStatus)
	*p = x
	return p
}

func (x ReceptionStatus) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ReceptionStatus) Descriptor() protoreflect.EnumDescriptor {
	return file_pvz_proto_enumTypes[0].Descriptor()
}

func (ReceptionStatus) Type() protoreflect.EnumType {
	return &file_pvz_proto_enumTypes[0]
}

func (x ReceptionStatus) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ReceptionStatus.Descriptor instead.
func (ReceptionStatus) EnumDescriptor() ([]byte, []int) {
	return file_pvz_proto_rawDescGZIP(), []int{0}
}

type PVZ struct {
	state            protoimpl.MessageState `protogen:"open.v1"`
	Id               string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"` // UUID как строка
	RegistrationDate *timestamppb.Timestamp `protobuf:"bytes,2,opt,name=registration_date,json=registrationDate,proto3" json:"registration_date,omitempty"`
	City             string                 `protobuf:"bytes,3,opt,name=city,proto3" json:"city,omitempty"`
	unknownFields    protoimpl.UnknownFields
	sizeCache        protoimpl.SizeCache
}

func (x *PVZ) Reset() {
	*x = PVZ{}
	mi := &file_pvz_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PVZ) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PVZ) ProtoMessage() {}

func (x *PVZ) ProtoReflect() protoreflect.Message {
	mi := &file_pvz_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PVZ.ProtoReflect.Descriptor instead.
func (*PVZ) Descriptor() ([]byte, []int) {
	return file_pvz_proto_rawDescGZIP(), []int{0}
}

func (x *PVZ) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *PVZ) GetRegistrationDate() *timestamppb.Timestamp {
	if x != nil {
		return x.RegistrationDate
	}
	return nil
}

func (x *PVZ) GetCity() string {
	if x != nil {
		return x.City
	}
	return ""
}

type GetPVZListRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetPVZListRequest) Reset() {
	*x = GetPVZListRequest{}
	mi := &file_pvz_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetPVZListRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetPVZListRequest) ProtoMessage() {}

func (x *GetPVZListRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pvz_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetPVZListRequest.ProtoReflect.Descriptor instead.
func (*GetPVZListRequest) Descriptor() ([]byte, []int) {
	return file_pvz_proto_rawDescGZIP(), []int{1}
}

type GetPVZListResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Pvzs          []*PVZ                 `protobuf:"bytes,1,rep,name=pvzs,proto3" json:"pvzs,omitempty"` // Изменено с pvzs на pvz для консистентности
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetPVZListResponse) Reset() {
	*x = GetPVZListResponse{}
	mi := &file_pvz_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetPVZListResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetPVZListResponse) ProtoMessage() {}

func (x *GetPVZListResponse) ProtoReflect() protoreflect.Message {
	mi := &file_pvz_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetPVZListResponse.ProtoReflect.Descriptor instead.
func (*GetPVZListResponse) Descriptor() ([]byte, []int) {
	return file_pvz_proto_rawDescGZIP(), []int{2}
}

func (x *GetPVZListResponse) GetPvzs() []*PVZ {
	if x != nil {
		return x.Pvzs
	}
	return nil
}

var File_pvz_proto protoreflect.FileDescriptor

const file_pvz_proto_rawDesc = "" +
	"\n" +
	"\tpvz.proto\x12\x06pvz.v1\x1a\x1fgoogle/protobuf/timestamp.proto\"r\n" +
	"\x03PVZ\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\tR\x02id\x12G\n" +
	"\x11registration_date\x18\x02 \x01(\v2\x1a.google.protobuf.TimestampR\x10registrationDate\x12\x12\n" +
	"\x04city\x18\x03 \x01(\tR\x04city\"\x13\n" +
	"\x11GetPVZListRequest\"5\n" +
	"\x12GetPVZListResponse\x12\x1f\n" +
	"\x04pvzs\x18\x01 \x03(\v2\v.pvz.v1.PVZR\x04pvzs*r\n" +
	"\x0fReceptionStatus\x12 \n" +
	"\x1cRECEPTION_STATUS_UNSPECIFIED\x10\x00\x12 \n" +
	"\x1cRECEPTION_STATUS_IN_PROGRESS\x10\x01\x12\x1b\n" +
	"\x17RECEPTION_STATUS_CLOSED\x10\x022Q\n" +
	"\n" +
	"PVZService\x12C\n" +
	"\n" +
	"GetPVZList\x12\x19.pvz.v1.GetPVZListRequest\x1a\x1a.pvz.v1.GetPVZListResponseB5Z3pvz-service-avito-internship/pkg/grpc/pvz/v1;pvz_v1b\x06proto3"

var (
	file_pvz_proto_rawDescOnce sync.Once
	file_pvz_proto_rawDescData []byte
)

func file_pvz_proto_rawDescGZIP() []byte {
	file_pvz_proto_rawDescOnce.Do(func() {
		file_pvz_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_pvz_proto_rawDesc), len(file_pvz_proto_rawDesc)))
	})
	return file_pvz_proto_rawDescData
}

var file_pvz_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_pvz_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_pvz_proto_goTypes = []any{
	(ReceptionStatus)(0),          // 0: pvz.v1.ReceptionStatus
	(*PVZ)(nil),                   // 1: pvz.v1.PVZ
	(*GetPVZListRequest)(nil),     // 2: pvz.v1.GetPVZListRequest
	(*GetPVZListResponse)(nil),    // 3: pvz.v1.GetPVZListResponse
	(*timestamppb.Timestamp)(nil), // 4: google.protobuf.Timestamp
}
var file_pvz_proto_depIdxs = []int32{
	4, // 0: pvz.v1.PVZ.registration_date:type_name -> google.protobuf.Timestamp
	1, // 1: pvz.v1.GetPVZListResponse.pvzs:type_name -> pvz.v1.PVZ
	2, // 2: pvz.v1.PVZService.GetPVZList:input_type -> pvz.v1.GetPVZListRequest
	3, // 3: pvz.v1.PVZService.GetPVZList:output_type -> pvz.v1.GetPVZListResponse
	3, // [3:4] is the sub-list for method output_type
	2, // [2:3] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_pvz_proto_init() }
func file_pvz_proto_init() {
	if File_pvz_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_pvz_proto_rawDesc), len(file_pvz_proto_rawDesc)),
			NumEnums:      1,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_pvz_proto_goTypes,
		DependencyIndexes: file_pvz_proto_depIdxs,
		EnumInfos:         file_pvz_proto_enumTypes,
		MessageInfos:      file_pvz_proto_msgTypes,
	}.Build()
	File_pvz_proto = out.File
	file_pvz_proto_goTypes = nil
	file_pvz_proto_depIdxs = nil
}
