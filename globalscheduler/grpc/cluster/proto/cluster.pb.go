/*
Copyright 2020 Authors of Arktos.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.14.0
// source: cluster.proto

package proto

import (
	proto "github.com/golang/protobuf/proto3"
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

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type ClusterProfile_ClusterSpecInfo_StorageType int32

const (
	ClusterProfile_ClusterSpecInfo_SATA ClusterProfile_ClusterSpecInfo_StorageType = 0
	ClusterProfile_ClusterSpecInfo_SAS  ClusterProfile_ClusterSpecInfo_StorageType = 1
	ClusterProfile_ClusterSpecInfo_SSD  ClusterProfile_ClusterSpecInfo_StorageType = 2
)

// Enum value maps for ClusterProfile_ClusterSpecInfo_StorageType.
var (
	ClusterProfile_ClusterSpecInfo_StorageType_name = map[int32]string{
		0: "SATA",
		1: "SAS",
		2: "SSD",
	}
	ClusterProfile_ClusterSpecInfo_StorageType_value = map[string]int32{
		"SATA": 0,
		"SAS":  1,
		"SSD":  2,
	}
)

func (x ClusterProfile_ClusterSpecInfo_StorageType) Enum() *ClusterProfile_ClusterSpecInfo_StorageType {
	p := new(ClusterProfile_ClusterSpecInfo_StorageType)
	*p = x
	return p
}

func (x ClusterProfile_ClusterSpecInfo_StorageType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ClusterProfile_ClusterSpecInfo_StorageType) Descriptor() protoreflect.EnumDescriptor {
	return file_cluster_proto_enumTypes[0].Descriptor()
}

func (ClusterProfile_ClusterSpecInfo_StorageType) Type() protoreflect.EnumType {
	return &file_cluster_proto_enumTypes[0]
}

func (x ClusterProfile_ClusterSpecInfo_StorageType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ClusterProfile_ClusterSpecInfo_StorageType.Descriptor instead.
func (ClusterProfile_ClusterSpecInfo_StorageType) EnumDescriptor() ([]byte, []int) {
	return file_cluster_proto_rawDescGZIP(), []int{0, 0, 0}
}

type ReturnMessageClusterProfile_CodeType int32

const (
	ReturnMessageClusterProfile_Error ReturnMessageClusterProfile_CodeType = 0
	ReturnMessageClusterProfile_OK    ReturnMessageClusterProfile_CodeType = 1
)

// Enum value maps for ReturnMessageClusterProfile_CodeType.
var (
	ReturnMessageClusterProfile_CodeType_name = map[int32]string{
		0: "Error",
		1: "OK",
	}
	ReturnMessageClusterProfile_CodeType_value = map[string]int32{
		"Error": 0,
		"OK":    1,
	}
)

func (x ReturnMessageClusterProfile_CodeType) Enum() *ReturnMessageClusterProfile_CodeType {
	p := new(ReturnMessageClusterProfile_CodeType)
	*p = x
	return p
}

func (x ReturnMessageClusterProfile_CodeType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ReturnMessageClusterProfile_CodeType) Descriptor() protoreflect.EnumDescriptor {
	return file_cluster_proto_enumTypes[1].Descriptor()
}

func (ReturnMessageClusterProfile_CodeType) Type() protoreflect.EnumType {
	return &file_cluster_proto_enumTypes[1]
}

func (x ReturnMessageClusterProfile_CodeType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ReturnMessageClusterProfile_CodeType.Descriptor instead.
func (ReturnMessageClusterProfile_CodeType) EnumDescriptor() ([]byte, []int) {
	return file_cluster_proto_rawDescGZIP(), []int{1, 0}
}

//message from ClusterController to ResourceCollector
type ClusterProfile struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ClusterNameSpace string                          `protobuf:"bytes,1,opt,name=ClusterNameSpace,proto3" json:"ClusterNameSpace,omitempty"`
	ClusterName      string                          `protobuf:"bytes,2,opt,name=ClusterName,proto3" json:"ClusterName,omitempty"`
	ClusterSpec      *ClusterProfile_ClusterSpecInfo `protobuf:"bytes,3,opt,name=ClusterSpec,proto3" json:"ClusterSpec,omitempty"`
}

func (x *ClusterProfile) Reset() {
	*x = ClusterProfile{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cluster_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ClusterProfile) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClusterProfile) ProtoMessage() {}

func (x *ClusterProfile) ProtoReflect() protoreflect.Message {
	mi := &file_cluster_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ClusterProfile.ProtoReflect.Descriptor instead.
func (*ClusterProfile) Descriptor() ([]byte, []int) {
	return file_cluster_proto_rawDescGZIP(), []int{0}
}

func (x *ClusterProfile) GetClusterNameSpace() string {
	if x != nil {
		return x.ClusterNameSpace
	}
	return ""
}

func (x *ClusterProfile) GetClusterName() string {
	if x != nil {
		return x.ClusterName
	}
	return ""
}

func (x *ClusterProfile) GetClusterSpec() *ClusterProfile_ClusterSpecInfo {
	if x != nil {
		return x.ClusterSpec
	}
	return nil
}

//Message from ResourceCollector, Cluster Controller should get response from ResourceCollector.
type ReturnMessageClusterProfile struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ClusterNameSpace string                               `protobuf:"bytes,1,opt,name=ClusterNameSpace,proto3" json:"ClusterNameSpace,omitempty"`
	ClusterName      string                               `protobuf:"bytes,2,opt,name=ClusterName,proto3" json:"ClusterName,omitempty"`
	ReturnCode       ReturnMessageClusterProfile_CodeType `protobuf:"varint,3,opt,name=ReturnCode,proto3,enum=proto.ReturnMessageClusterProfile_CodeType" json:"ReturnCode,omitempty"`
	Message          string                               `protobuf:"bytes,4,opt,name=Message,proto3" json:"Message,omitempty"`
}

func (x *ReturnMessageClusterProfile) Reset() {
	*x = ReturnMessageClusterProfile{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cluster_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReturnMessageClusterProfile) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReturnMessageClusterProfile) ProtoMessage() {}

func (x *ReturnMessageClusterProfile) ProtoReflect() protoreflect.Message {
	mi := &file_cluster_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReturnMessageClusterProfile.ProtoReflect.Descriptor instead.
func (*ReturnMessageClusterProfile) Descriptor() ([]byte, []int) {
	return file_cluster_proto_rawDescGZIP(), []int{1}
}

func (x *ReturnMessageClusterProfile) GetClusterNameSpace() string {
	if x != nil {
		return x.ClusterNameSpace
	}
	return ""
}

func (x *ReturnMessageClusterProfile) GetClusterName() string {
	if x != nil {
		return x.ClusterName
	}
	return ""
}

func (x *ReturnMessageClusterProfile) GetReturnCode() ReturnMessageClusterProfile_CodeType {
	if x != nil {
		return x.ReturnCode
	}
	return ReturnMessageClusterProfile_Error
}

func (x *ReturnMessageClusterProfile) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

type ClusterProfile_ClusterSpecInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ClusterIpAddress string                                          `protobuf:"bytes,1,opt,name=ClusterIpAddress,proto3" json:"ClusterIpAddress,omitempty"`
	GeoLocation      *ClusterProfile_ClusterSpecInfo_GeoLocationInfo `protobuf:"bytes,2,opt,name=GeoLocation,proto3" json:"GeoLocation,omitempty"`
	Region           *ClusterProfile_ClusterSpecInfo_RegionInfo      `protobuf:"bytes,3,opt,name=Region,proto3" json:"Region,omitempty"`
	Operator         *ClusterProfile_ClusterSpecInfo_OperatorInfo    `protobuf:"bytes,4,opt,name=Operator,proto3" json:"Operator,omitempty"`
	Flavor           []*ClusterProfile_ClusterSpecInfo_FlavorInfo    `protobuf:"bytes,5,rep,name=Flavor,proto3" json:"Flavor,omitempty"`
	Storage          []*ClusterProfile_ClusterSpecInfo_StorageInfo   `protobuf:"bytes,6,rep,name=Storage,proto3" json:"Storage,omitempty"`
	EipCapacity      int64                                           `protobuf:"varint,7,opt,name=EipCapacity,proto3" json:"EipCapacity,omitempty"`
	CPUCapacity      int64                                           `protobuf:"varint,8,opt,name=CPUCapacity,proto3" json:"CPUCapacity,omitempty"`
	MemCapacity      int64                                           `protobuf:"varint,9,opt,name=MemCapacity,proto3" json:"MemCapacity,omitempty"`
	ServerPrice      int64                                           `protobuf:"varint,10,opt,name=ServerPrice,proto3" json:"ServerPrice,omitempty"`
	HomeScheduler    string                                          `protobuf:"bytes,11,opt,name=HomeScheduler,proto3" json:"HomeScheduler,omitempty"`
}

func (x *ClusterProfile_ClusterSpecInfo) Reset() {
	*x = ClusterProfile_ClusterSpecInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cluster_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ClusterProfile_ClusterSpecInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClusterProfile_ClusterSpecInfo) ProtoMessage() {}

func (x *ClusterProfile_ClusterSpecInfo) ProtoReflect() protoreflect.Message {
	mi := &file_cluster_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ClusterProfile_ClusterSpecInfo.ProtoReflect.Descriptor instead.
func (*ClusterProfile_ClusterSpecInfo) Descriptor() ([]byte, []int) {
	return file_cluster_proto_rawDescGZIP(), []int{0, 0}
}

func (x *ClusterProfile_ClusterSpecInfo) GetClusterIpAddress() string {
	if x != nil {
		return x.ClusterIpAddress
	}
	return ""
}

func (x *ClusterProfile_ClusterSpecInfo) GetGeoLocation() *ClusterProfile_ClusterSpecInfo_GeoLocationInfo {
	if x != nil {
		return x.GeoLocation
	}
	return nil
}

func (x *ClusterProfile_ClusterSpecInfo) GetRegion() *ClusterProfile_ClusterSpecInfo_RegionInfo {
	if x != nil {
		return x.Region
	}
	return nil
}

func (x *ClusterProfile_ClusterSpecInfo) GetOperator() *ClusterProfile_ClusterSpecInfo_OperatorInfo {
	if x != nil {
		return x.Operator
	}
	return nil
}

func (x *ClusterProfile_ClusterSpecInfo) GetFlavor() []*ClusterProfile_ClusterSpecInfo_FlavorInfo {
	if x != nil {
		return x.Flavor
	}
	return nil
}

func (x *ClusterProfile_ClusterSpecInfo) GetStorage() []*ClusterProfile_ClusterSpecInfo_StorageInfo {
	if x != nil {
		return x.Storage
	}
	return nil
}

func (x *ClusterProfile_ClusterSpecInfo) GetEipCapacity() int64 {
	if x != nil {
		return x.EipCapacity
	}
	return 0
}

func (x *ClusterProfile_ClusterSpecInfo) GetCPUCapacity() int64 {
	if x != nil {
		return x.CPUCapacity
	}
	return 0
}

func (x *ClusterProfile_ClusterSpecInfo) GetMemCapacity() int64 {
	if x != nil {
		return x.MemCapacity
	}
	return 0
}

func (x *ClusterProfile_ClusterSpecInfo) GetServerPrice() int64 {
	if x != nil {
		return x.ServerPrice
	}
	return 0
}

func (x *ClusterProfile_ClusterSpecInfo) GetHomeScheduler() string {
	if x != nil {
		return x.HomeScheduler
	}
	return ""
}

type ClusterProfile_ClusterSpecInfo_GeoLocationInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	City     string `protobuf:"bytes,1,opt,name=City,proto3" json:"City,omitempty"`
	Province string `protobuf:"bytes,2,opt,name=Province,proto3" json:"Province,omitempty"`
	Area     string `protobuf:"bytes,3,opt,name=Area,proto3" json:"Area,omitempty"`
	Country  string `protobuf:"bytes,4,opt,name=Country,proto3" json:"Country,omitempty"`
}

func (x *ClusterProfile_ClusterSpecInfo_GeoLocationInfo) Reset() {
	*x = ClusterProfile_ClusterSpecInfo_GeoLocationInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cluster_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ClusterProfile_ClusterSpecInfo_GeoLocationInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClusterProfile_ClusterSpecInfo_GeoLocationInfo) ProtoMessage() {}

func (x *ClusterProfile_ClusterSpecInfo_GeoLocationInfo) ProtoReflect() protoreflect.Message {
	mi := &file_cluster_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ClusterProfile_ClusterSpecInfo_GeoLocationInfo.ProtoReflect.Descriptor instead.
func (*ClusterProfile_ClusterSpecInfo_GeoLocationInfo) Descriptor() ([]byte, []int) {
	return file_cluster_proto_rawDescGZIP(), []int{0, 0, 0}
}

func (x *ClusterProfile_ClusterSpecInfo_GeoLocationInfo) GetCity() string {
	if x != nil {
		return x.City
	}
	return ""
}

func (x *ClusterProfile_ClusterSpecInfo_GeoLocationInfo) GetProvince() string {
	if x != nil {
		return x.Province
	}
	return ""
}

func (x *ClusterProfile_ClusterSpecInfo_GeoLocationInfo) GetArea() string {
	if x != nil {
		return x.Area
	}
	return ""
}

func (x *ClusterProfile_ClusterSpecInfo_GeoLocationInfo) GetCountry() string {
	if x != nil {
		return x.Country
	}
	return ""
}

type ClusterProfile_ClusterSpecInfo_RegionInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Region           string `protobuf:"bytes,1,opt,name=Region,proto3" json:"Region,omitempty"`
	AvailabilityZone string `protobuf:"bytes,2,opt,name=AvailabilityZone,proto3" json:"AvailabilityZone,omitempty"`
}

func (x *ClusterProfile_ClusterSpecInfo_RegionInfo) Reset() {
	*x = ClusterProfile_ClusterSpecInfo_RegionInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cluster_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ClusterProfile_ClusterSpecInfo_RegionInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClusterProfile_ClusterSpecInfo_RegionInfo) ProtoMessage() {}

func (x *ClusterProfile_ClusterSpecInfo_RegionInfo) ProtoReflect() protoreflect.Message {
	mi := &file_cluster_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ClusterProfile_ClusterSpecInfo_RegionInfo.ProtoReflect.Descriptor instead.
func (*ClusterProfile_ClusterSpecInfo_RegionInfo) Descriptor() ([]byte, []int) {
	return file_cluster_proto_rawDescGZIP(), []int{0, 0, 1}
}

func (x *ClusterProfile_ClusterSpecInfo_RegionInfo) GetRegion() string {
	if x != nil {
		return x.Region
	}
	return ""
}

func (x *ClusterProfile_ClusterSpecInfo_RegionInfo) GetAvailabilityZone() string {
	if x != nil {
		return x.AvailabilityZone
	}
	return ""
}

type ClusterProfile_ClusterSpecInfo_OperatorInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Operator string `protobuf:"bytes,1,opt,name=Operator,proto3" json:"Operator,omitempty"`
}

func (x *ClusterProfile_ClusterSpecInfo_OperatorInfo) Reset() {
	*x = ClusterProfile_ClusterSpecInfo_OperatorInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cluster_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ClusterProfile_ClusterSpecInfo_OperatorInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClusterProfile_ClusterSpecInfo_OperatorInfo) ProtoMessage() {}

func (x *ClusterProfile_ClusterSpecInfo_OperatorInfo) ProtoReflect() protoreflect.Message {
	mi := &file_cluster_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ClusterProfile_ClusterSpecInfo_OperatorInfo.ProtoReflect.Descriptor instead.
func (*ClusterProfile_ClusterSpecInfo_OperatorInfo) Descriptor() ([]byte, []int) {
	return file_cluster_proto_rawDescGZIP(), []int{0, 0, 2}
}

func (x *ClusterProfile_ClusterSpecInfo_OperatorInfo) GetOperator() string {
	if x != nil {
		return x.Operator
	}
	return ""
}

type ClusterProfile_ClusterSpecInfo_FlavorInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	FlavorID      int64 `protobuf:"varint,1,opt,name=FlavorID,proto3" json:"FlavorID,omitempty"` //1:Small, 2:Medium, 3:Large, 4:Xlarge, 5:2xLarge
	TotalCapacity int64 `protobuf:"varint,2,opt,name=TotalCapacity,proto3" json:"TotalCapacity,omitempty"`
}

func (x *ClusterProfile_ClusterSpecInfo_FlavorInfo) Reset() {
	*x = ClusterProfile_ClusterSpecInfo_FlavorInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cluster_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ClusterProfile_ClusterSpecInfo_FlavorInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClusterProfile_ClusterSpecInfo_FlavorInfo) ProtoMessage() {}

func (x *ClusterProfile_ClusterSpecInfo_FlavorInfo) ProtoReflect() protoreflect.Message {
	mi := &file_cluster_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ClusterProfile_ClusterSpecInfo_FlavorInfo.ProtoReflect.Descriptor instead.
func (*ClusterProfile_ClusterSpecInfo_FlavorInfo) Descriptor() ([]byte, []int) {
	return file_cluster_proto_rawDescGZIP(), []int{0, 0, 3}
}

func (x *ClusterProfile_ClusterSpecInfo_FlavorInfo) GetFlavorID() int64 {
	if x != nil {
		return x.FlavorID
	}
	return 0
}

func (x *ClusterProfile_ClusterSpecInfo_FlavorInfo) GetTotalCapacity() int64 {
	if x != nil {
		return x.TotalCapacity
	}
	return 0
}

type ClusterProfile_ClusterSpecInfo_StorageInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TypeID          ClusterProfile_ClusterSpecInfo_StorageType `protobuf:"varint,1,opt,name=TypeID,proto3,enum=proto.ClusterProfile_ClusterSpecInfo_StorageType" json:"TypeID,omitempty"`
	StorageCapacity int64                                      `protobuf:"varint,2,opt,name=StorageCapacity,proto3" json:"StorageCapacity,omitempty"`
}

func (x *ClusterProfile_ClusterSpecInfo_StorageInfo) Reset() {
	*x = ClusterProfile_ClusterSpecInfo_StorageInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cluster_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ClusterProfile_ClusterSpecInfo_StorageInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClusterProfile_ClusterSpecInfo_StorageInfo) ProtoMessage() {}

func (x *ClusterProfile_ClusterSpecInfo_StorageInfo) ProtoReflect() protoreflect.Message {
	mi := &file_cluster_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ClusterProfile_ClusterSpecInfo_StorageInfo.ProtoReflect.Descriptor instead.
func (*ClusterProfile_ClusterSpecInfo_StorageInfo) Descriptor() ([]byte, []int) {
	return file_cluster_proto_rawDescGZIP(), []int{0, 0, 4}
}

func (x *ClusterProfile_ClusterSpecInfo_StorageInfo) GetTypeID() ClusterProfile_ClusterSpecInfo_StorageType {
	if x != nil {
		return x.TypeID
	}
	return ClusterProfile_ClusterSpecInfo_SATA
}

func (x *ClusterProfile_ClusterSpecInfo_StorageInfo) GetStorageCapacity() int64 {
	if x != nil {
		return x.StorageCapacity
	}
	return 0
}

var File_cluster_proto protoreflect.FileDescriptor

var file_cluster_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x05, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x8e, 0x0a, 0x0a, 0x0e, 0x43, 0x6c, 0x75, 0x73, 0x74,
	0x65, 0x72, 0x50, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65, 0x12, 0x2a, 0x0a, 0x10, 0x43, 0x6c, 0x75,
	0x73, 0x74, 0x65, 0x72, 0x4e, 0x61, 0x6d, 0x65, 0x53, 0x70, 0x61, 0x63, 0x65, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x10, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x4e, 0x61, 0x6d, 0x65,
	0x53, 0x70, 0x61, 0x63, 0x65, 0x12, 0x20, 0x0a, 0x0b, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72,
	0x4e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x43, 0x6c, 0x75, 0x73,
	0x74, 0x65, 0x72, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x47, 0x0a, 0x0b, 0x43, 0x6c, 0x75, 0x73, 0x74,
	0x65, 0x72, 0x53, 0x70, 0x65, 0x63, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x25, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x50, 0x72, 0x6f, 0x66,
	0x69, 0x6c, 0x65, 0x2e, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x53, 0x70, 0x65, 0x63, 0x49,
	0x6e, 0x66, 0x6f, 0x52, 0x0b, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x53, 0x70, 0x65, 0x63,
	0x1a, 0xe4, 0x08, 0x0a, 0x0f, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x53, 0x70, 0x65, 0x63,
	0x49, 0x6e, 0x66, 0x6f, 0x12, 0x2a, 0x0a, 0x10, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x49,
	0x70, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x10,
	0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x49, 0x70, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73,
	0x12, 0x57, 0x0a, 0x0b, 0x47, 0x65, 0x6f, 0x4c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x35, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x43, 0x6c,
	0x75, 0x73, 0x74, 0x65, 0x72, 0x50, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65, 0x2e, 0x43, 0x6c, 0x75,
	0x73, 0x74, 0x65, 0x72, 0x53, 0x70, 0x65, 0x63, 0x49, 0x6e, 0x66, 0x6f, 0x2e, 0x47, 0x65, 0x6f,
	0x4c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x0b, 0x47, 0x65,
	0x6f, 0x4c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x48, 0x0a, 0x06, 0x52, 0x65, 0x67,
	0x69, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x30, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x2e, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x50, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65,
	0x2e, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x53, 0x70, 0x65, 0x63, 0x49, 0x6e, 0x66, 0x6f,
	0x2e, 0x52, 0x65, 0x67, 0x69, 0x6f, 0x6e, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x06, 0x52, 0x65, 0x67,
	0x69, 0x6f, 0x6e, 0x12, 0x4e, 0x0a, 0x08, 0x4f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x32, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x43, 0x6c,
	0x75, 0x73, 0x74, 0x65, 0x72, 0x50, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65, 0x2e, 0x43, 0x6c, 0x75,
	0x73, 0x74, 0x65, 0x72, 0x53, 0x70, 0x65, 0x63, 0x49, 0x6e, 0x66, 0x6f, 0x2e, 0x4f, 0x70, 0x65,
	0x72, 0x61, 0x74, 0x6f, 0x72, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x08, 0x4f, 0x70, 0x65, 0x72, 0x61,
	0x74, 0x6f, 0x72, 0x12, 0x48, 0x0a, 0x06, 0x46, 0x6c, 0x61, 0x76, 0x6f, 0x72, 0x18, 0x05, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x30, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x43, 0x6c, 0x75, 0x73,
	0x74, 0x65, 0x72, 0x50, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65, 0x2e, 0x43, 0x6c, 0x75, 0x73, 0x74,
	0x65, 0x72, 0x53, 0x70, 0x65, 0x63, 0x49, 0x6e, 0x66, 0x6f, 0x2e, 0x46, 0x6c, 0x61, 0x76, 0x6f,
	0x72, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x06, 0x46, 0x6c, 0x61, 0x76, 0x6f, 0x72, 0x12, 0x4b, 0x0a,
	0x07, 0x53, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x18, 0x06, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x31,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x50, 0x72,
	0x6f, 0x66, 0x69, 0x6c, 0x65, 0x2e, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x53, 0x70, 0x65,
	0x63, 0x49, 0x6e, 0x66, 0x6f, 0x2e, 0x53, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x49, 0x6e, 0x66,
	0x6f, 0x52, 0x07, 0x53, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x12, 0x20, 0x0a, 0x0b, 0x45, 0x69,
	0x70, 0x43, 0x61, 0x70, 0x61, 0x63, 0x69, 0x74, 0x79, 0x18, 0x07, 0x20, 0x01, 0x28, 0x03, 0x52,
	0x0b, 0x45, 0x69, 0x70, 0x43, 0x61, 0x70, 0x61, 0x63, 0x69, 0x74, 0x79, 0x12, 0x20, 0x0a, 0x0b,
	0x43, 0x50, 0x55, 0x43, 0x61, 0x70, 0x61, 0x63, 0x69, 0x74, 0x79, 0x18, 0x08, 0x20, 0x01, 0x28,
	0x03, 0x52, 0x0b, 0x43, 0x50, 0x55, 0x43, 0x61, 0x70, 0x61, 0x63, 0x69, 0x74, 0x79, 0x12, 0x20,
	0x0a, 0x0b, 0x4d, 0x65, 0x6d, 0x43, 0x61, 0x70, 0x61, 0x63, 0x69, 0x74, 0x79, 0x18, 0x09, 0x20,
	0x01, 0x28, 0x03, 0x52, 0x0b, 0x4d, 0x65, 0x6d, 0x43, 0x61, 0x70, 0x61, 0x63, 0x69, 0x74, 0x79,
	0x12, 0x20, 0x0a, 0x0b, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x50, 0x72, 0x69, 0x63, 0x65, 0x18,
	0x0a, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0b, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x50, 0x72, 0x69,
	0x63, 0x65, 0x12, 0x24, 0x0a, 0x0d, 0x48, 0x6f, 0x6d, 0x65, 0x53, 0x63, 0x68, 0x65, 0x64, 0x75,
	0x6c, 0x65, 0x72, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x48, 0x6f, 0x6d, 0x65, 0x53,
	0x63, 0x68, 0x65, 0x64, 0x75, 0x6c, 0x65, 0x72, 0x1a, 0x6f, 0x0a, 0x0f, 0x47, 0x65, 0x6f, 0x4c,
	0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x12, 0x0a, 0x04, 0x43,
	0x69, 0x74, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x43, 0x69, 0x74, 0x79, 0x12,
	0x1a, 0x0a, 0x08, 0x50, 0x72, 0x6f, 0x76, 0x69, 0x6e, 0x63, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x08, 0x50, 0x72, 0x6f, 0x76, 0x69, 0x6e, 0x63, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x41,
	0x72, 0x65, 0x61, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x41, 0x72, 0x65, 0x61, 0x12,
	0x18, 0x0a, 0x07, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x79, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x07, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x79, 0x1a, 0x50, 0x0a, 0x0a, 0x52, 0x65, 0x67,
	0x69, 0x6f, 0x6e, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x16, 0x0a, 0x06, 0x52, 0x65, 0x67, 0x69, 0x6f,
	0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x52, 0x65, 0x67, 0x69, 0x6f, 0x6e, 0x12,
	0x2a, 0x0a, 0x10, 0x41, 0x76, 0x61, 0x69, 0x6c, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x79, 0x5a,
	0x6f, 0x6e, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x10, 0x41, 0x76, 0x61, 0x69, 0x6c,
	0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x79, 0x5a, 0x6f, 0x6e, 0x65, 0x1a, 0x2a, 0x0a, 0x0c, 0x4f,
	0x70, 0x65, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x1a, 0x0a, 0x08, 0x4f,
	0x70, 0x65, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x4f,
	0x70, 0x65, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x1a, 0x4e, 0x0a, 0x0a, 0x46, 0x6c, 0x61, 0x76, 0x6f,
	0x72, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x1a, 0x0a, 0x08, 0x46, 0x6c, 0x61, 0x76, 0x6f, 0x72, 0x49,
	0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x08, 0x46, 0x6c, 0x61, 0x76, 0x6f, 0x72, 0x49,
	0x44, 0x12, 0x24, 0x0a, 0x0d, 0x54, 0x6f, 0x74, 0x61, 0x6c, 0x43, 0x61, 0x70, 0x61, 0x63, 0x69,
	0x74, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0d, 0x54, 0x6f, 0x74, 0x61, 0x6c, 0x43,
	0x61, 0x70, 0x61, 0x63, 0x69, 0x74, 0x79, 0x1a, 0x82, 0x01, 0x0a, 0x0b, 0x53, 0x74, 0x6f, 0x72,
	0x61, 0x67, 0x65, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x49, 0x0a, 0x06, 0x54, 0x79, 0x70, 0x65, 0x49,
	0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x31, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e,
	0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x50, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65, 0x2e, 0x43,
	0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x53, 0x70, 0x65, 0x63, 0x49, 0x6e, 0x66, 0x6f, 0x2e, 0x53,
	0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x54, 0x79, 0x70, 0x65, 0x52, 0x06, 0x54, 0x79, 0x70, 0x65,
	0x49, 0x44, 0x12, 0x28, 0x0a, 0x0f, 0x53, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x43, 0x61, 0x70,
	0x61, 0x63, 0x69, 0x74, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0f, 0x53, 0x74, 0x6f,
	0x72, 0x61, 0x67, 0x65, 0x43, 0x61, 0x70, 0x61, 0x63, 0x69, 0x74, 0x79, 0x22, 0x29, 0x0a, 0x0b,
	0x53, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x54, 0x79, 0x70, 0x65, 0x12, 0x08, 0x0a, 0x04, 0x53,
	0x41, 0x54, 0x41, 0x10, 0x00, 0x12, 0x07, 0x0a, 0x03, 0x53, 0x41, 0x53, 0x10, 0x01, 0x12, 0x07,
	0x0a, 0x03, 0x53, 0x53, 0x44, 0x10, 0x02, 0x22, 0xf1, 0x01, 0x0a, 0x1b, 0x52, 0x65, 0x74, 0x75,
	0x72, 0x6e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72,
	0x50, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65, 0x12, 0x2a, 0x0a, 0x10, 0x43, 0x6c, 0x75, 0x73, 0x74,
	0x65, 0x72, 0x4e, 0x61, 0x6d, 0x65, 0x53, 0x70, 0x61, 0x63, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x10, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x4e, 0x61, 0x6d, 0x65, 0x53, 0x70,
	0x61, 0x63, 0x65, 0x12, 0x20, 0x0a, 0x0b, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x4e, 0x61,
	0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65,
	0x72, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x4b, 0x0a, 0x0a, 0x52, 0x65, 0x74, 0x75, 0x72, 0x6e, 0x43,
	0x6f, 0x64, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x2b, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x2e, 0x52, 0x65, 0x74, 0x75, 0x72, 0x6e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x43,
	0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x50, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65, 0x2e, 0x43, 0x6f,
	0x64, 0x65, 0x54, 0x79, 0x70, 0x65, 0x52, 0x0a, 0x52, 0x65, 0x74, 0x75, 0x72, 0x6e, 0x43, 0x6f,
	0x64, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x07, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x22, 0x1d, 0x0a, 0x08,
	0x43, 0x6f, 0x64, 0x65, 0x54, 0x79, 0x70, 0x65, 0x12, 0x09, 0x0a, 0x05, 0x45, 0x72, 0x72, 0x6f,
	0x72, 0x10, 0x00, 0x12, 0x06, 0x0a, 0x02, 0x4f, 0x4b, 0x10, 0x01, 0x32, 0x64, 0x0a, 0x0f, 0x43,
	0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x12, 0x51,
	0x0a, 0x12, 0x53, 0x65, 0x6e, 0x64, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x50, 0x72, 0x6f,
	0x66, 0x69, 0x6c, 0x65, 0x12, 0x15, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x43, 0x6c, 0x75,
	0x73, 0x74, 0x65, 0x72, 0x50, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65, 0x1a, 0x22, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x2e, 0x52, 0x65, 0x74, 0x75, 0x72, 0x6e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x50, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65, 0x22,
	0x00, 0x42, 0x36, 0x5a, 0x34, 0x6b, 0x38, 0x73, 0x2e, 0x69, 0x6f, 0x2f, 0x6b, 0x75, 0x62, 0x65,
	0x72, 0x6e, 0x65, 0x74, 0x65, 0x73, 0x2f, 0x67, 0x6c, 0x6f, 0x62, 0x61, 0x6c, 0x73, 0x63, 0x68,
	0x65, 0x64, 0x75, 0x6c, 0x65, 0x72, 0x2f, 0x67, 0x72, 0x70, 0x63, 0x2f, 0x63, 0x6c, 0x75, 0x73,
	0x74, 0x65, 0x72, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_cluster_proto_rawDescOnce sync.Once
	file_cluster_proto_rawDescData = file_cluster_proto_rawDesc
)

func file_cluster_proto_rawDescGZIP() []byte {
	file_cluster_proto_rawDescOnce.Do(func() {
		file_cluster_proto_rawDescData = protoimpl.X.CompressGZIP(file_cluster_proto_rawDescData)
	})
	return file_cluster_proto_rawDescData
}

var file_cluster_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_cluster_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_cluster_proto_goTypes = []interface{}{
	(ClusterProfile_ClusterSpecInfo_StorageType)(0),        // 0: proto.ClusterProfile.ClusterSpecInfo.StorageType
	(ReturnMessageClusterProfile_CodeType)(0),              // 1: proto.ReturnMessageClusterProfile.CodeType
	(*ClusterProfile)(nil),                                 // 2: proto.ClusterProfile
	(*ReturnMessageClusterProfile)(nil),                    // 3: proto.ReturnMessageClusterProfile
	(*ClusterProfile_ClusterSpecInfo)(nil),                 // 4: proto.ClusterProfile.ClusterSpecInfo
	(*ClusterProfile_ClusterSpecInfo_GeoLocationInfo)(nil), // 5: proto.ClusterProfile.ClusterSpecInfo.GeoLocationInfo
	(*ClusterProfile_ClusterSpecInfo_RegionInfo)(nil),      // 6: proto.ClusterProfile.ClusterSpecInfo.RegionInfo
	(*ClusterProfile_ClusterSpecInfo_OperatorInfo)(nil),    // 7: proto.ClusterProfile.ClusterSpecInfo.OperatorInfo
	(*ClusterProfile_ClusterSpecInfo_FlavorInfo)(nil),      // 8: proto.ClusterProfile.ClusterSpecInfo.FlavorInfo
	(*ClusterProfile_ClusterSpecInfo_StorageInfo)(nil),     // 9: proto.ClusterProfile.ClusterSpecInfo.StorageInfo
}
var file_cluster_proto_depIdxs = []int32{
	4, // 0: proto.ClusterProfile.ClusterSpec:type_name -> proto.ClusterProfile.ClusterSpecInfo
	1, // 1: proto.ReturnMessageClusterProfile.ReturnCode:type_name -> proto.ReturnMessageClusterProfile.CodeType
	5, // 2: proto.ClusterProfile.ClusterSpecInfo.GeoLocation:type_name -> proto.ClusterProfile.ClusterSpecInfo.GeoLocationInfo
	6, // 3: proto.ClusterProfile.ClusterSpecInfo.Region:type_name -> proto.ClusterProfile.ClusterSpecInfo.RegionInfo
	7, // 4: proto.ClusterProfile.ClusterSpecInfo.Operator:type_name -> proto.ClusterProfile.ClusterSpecInfo.OperatorInfo
	8, // 5: proto.ClusterProfile.ClusterSpecInfo.Flavor:type_name -> proto.ClusterProfile.ClusterSpecInfo.FlavorInfo
	9, // 6: proto.ClusterProfile.ClusterSpecInfo.Storage:type_name -> proto.ClusterProfile.ClusterSpecInfo.StorageInfo
	0, // 7: proto.ClusterProfile.ClusterSpecInfo.StorageInfo.TypeID:type_name -> proto.ClusterProfile.ClusterSpecInfo.StorageType
	2, // 8: proto.ClusterProtocol.SendClusterProfile:input_type -> proto.ClusterProfile
	3, // 9: proto.ClusterProtocol.SendClusterProfile:output_type -> proto.ReturnMessageClusterProfile
	9, // [9:10] is the sub-list for method output_type
	8, // [8:9] is the sub-list for method input_type
	8, // [8:8] is the sub-list for extension type_name
	8, // [8:8] is the sub-list for extension extendee
	0, // [0:8] is the sub-list for field type_name
}

func init() { file_cluster_proto_init() }
func file_cluster_proto_init() {
	if File_cluster_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_cluster_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ClusterProfile); i {
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
		file_cluster_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReturnMessageClusterProfile); i {
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
		file_cluster_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ClusterProfile_ClusterSpecInfo); i {
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
		file_cluster_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ClusterProfile_ClusterSpecInfo_GeoLocationInfo); i {
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
		file_cluster_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ClusterProfile_ClusterSpecInfo_RegionInfo); i {
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
		file_cluster_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ClusterProfile_ClusterSpecInfo_OperatorInfo); i {
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
		file_cluster_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ClusterProfile_ClusterSpecInfo_FlavorInfo); i {
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
		file_cluster_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ClusterProfile_ClusterSpecInfo_StorageInfo); i {
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
			RawDescriptor: file_cluster_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_cluster_proto_goTypes,
		DependencyIndexes: file_cluster_proto_depIdxs,
		EnumInfos:         file_cluster_proto_enumTypes,
		MessageInfos:      file_cluster_proto_msgTypes,
	}.Build()
	File_cluster_proto = out.File
	file_cluster_proto_rawDesc = nil
	file_cluster_proto_goTypes = nil
	file_cluster_proto_depIdxs = nil
}
