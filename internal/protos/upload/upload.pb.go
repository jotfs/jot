// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.20.1
// 	protoc        (unknown)
// source: internal/protos/upload/upload.proto

package upload

import (
	proto "github.com/golang/protobuf/proto"
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

type ChunksExistRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Sums [][]byte `protobuf:"bytes,1,rep,name=sums,proto3" json:"sums,omitempty"`
}

func (x *ChunksExistRequest) Reset() {
	*x = ChunksExistRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_protos_upload_upload_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ChunksExistRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ChunksExistRequest) ProtoMessage() {}

func (x *ChunksExistRequest) ProtoReflect() protoreflect.Message {
	mi := &file_internal_protos_upload_upload_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ChunksExistRequest.ProtoReflect.Descriptor instead.
func (*ChunksExistRequest) Descriptor() ([]byte, []int) {
	return file_internal_protos_upload_upload_proto_rawDescGZIP(), []int{0}
}

func (x *ChunksExistRequest) GetSums() [][]byte {
	if x != nil {
		return x.Sums
	}
	return nil
}

type ChunksExistResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Exists []bool `protobuf:"varint,1,rep,packed,name=exists,proto3" json:"exists,omitempty"`
}

func (x *ChunksExistResponse) Reset() {
	*x = ChunksExistResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_protos_upload_upload_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ChunksExistResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ChunksExistResponse) ProtoMessage() {}

func (x *ChunksExistResponse) ProtoReflect() protoreflect.Message {
	mi := &file_internal_protos_upload_upload_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ChunksExistResponse.ProtoReflect.Descriptor instead.
func (*ChunksExistResponse) Descriptor() ([]byte, []int) {
	return file_internal_protos_upload_upload_proto_rawDescGZIP(), []int{1}
}

func (x *ChunksExistResponse) GetExists() []bool {
	if x != nil {
		return x.Exists
	}
	return nil
}

type File struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Sums [][]byte `protobuf:"bytes,2,rep,name=sums,proto3" json:"sums,omitempty"`
}

func (x *File) Reset() {
	*x = File{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_protos_upload_upload_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *File) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*File) ProtoMessage() {}

func (x *File) ProtoReflect() protoreflect.Message {
	mi := &file_internal_protos_upload_upload_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use File.ProtoReflect.Descriptor instead.
func (*File) Descriptor() ([]byte, []int) {
	return file_internal_protos_upload_upload_proto_rawDescGZIP(), []int{2}
}

func (x *File) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *File) GetSums() [][]byte {
	if x != nil {
		return x.Sums
	}
	return nil
}

type CopyRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SrcId []byte `protobuf:"bytes,1,opt,name=src_id,json=srcId,proto3" json:"src_id,omitempty"`
	Dst   string `protobuf:"bytes,2,opt,name=dst,proto3" json:"dst,omitempty"`
}

func (x *CopyRequest) Reset() {
	*x = CopyRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_protos_upload_upload_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CopyRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CopyRequest) ProtoMessage() {}

func (x *CopyRequest) ProtoReflect() protoreflect.Message {
	mi := &file_internal_protos_upload_upload_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CopyRequest.ProtoReflect.Descriptor instead.
func (*CopyRequest) Descriptor() ([]byte, []int) {
	return file_internal_protos_upload_upload_proto_rawDescGZIP(), []int{3}
}

func (x *CopyRequest) GetSrcId() []byte {
	if x != nil {
		return x.SrcId
	}
	return nil
}

func (x *CopyRequest) GetDst() string {
	if x != nil {
		return x.Dst
	}
	return ""
}

type FileID struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Sum []byte `protobuf:"bytes,1,opt,name=sum,proto3" json:"sum,omitempty"`
}

func (x *FileID) Reset() {
	*x = FileID{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_protos_upload_upload_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FileID) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FileID) ProtoMessage() {}

func (x *FileID) ProtoReflect() protoreflect.Message {
	mi := &file_internal_protos_upload_upload_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FileID.ProtoReflect.Descriptor instead.
func (*FileID) Descriptor() ([]byte, []int) {
	return file_internal_protos_upload_upload_proto_rawDescGZIP(), []int{4}
}

func (x *FileID) GetSum() []byte {
	if x != nil {
		return x.Sum
	}
	return nil
}

type Prefix struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Prefix string `protobuf:"bytes,1,opt,name=prefix,proto3" json:"prefix,omitempty"`
}

func (x *Prefix) Reset() {
	*x = Prefix{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_protos_upload_upload_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Prefix) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Prefix) ProtoMessage() {}

func (x *Prefix) ProtoReflect() protoreflect.Message {
	mi := &file_internal_protos_upload_upload_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Prefix.ProtoReflect.Descriptor instead.
func (*Prefix) Descriptor() ([]byte, []int) {
	return file_internal_protos_upload_upload_proto_rawDescGZIP(), []int{5}
}

func (x *Prefix) GetPrefix() string {
	if x != nil {
		return x.Prefix
	}
	return ""
}

type HeadFileRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name          string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Limit         uint64 `protobuf:"varint,2,opt,name=limit,proto3" json:"limit,omitempty"`
	NextPageToken int64  `protobuf:"varint,3,opt,name=next_page_token,json=nextPageToken,proto3" json:"next_page_token,omitempty"`
}

func (x *HeadFileRequest) Reset() {
	*x = HeadFileRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_protos_upload_upload_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HeadFileRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HeadFileRequest) ProtoMessage() {}

func (x *HeadFileRequest) ProtoReflect() protoreflect.Message {
	mi := &file_internal_protos_upload_upload_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HeadFileRequest.ProtoReflect.Descriptor instead.
func (*HeadFileRequest) Descriptor() ([]byte, []int) {
	return file_internal_protos_upload_upload_proto_rawDescGZIP(), []int{6}
}

func (x *HeadFileRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *HeadFileRequest) GetLimit() uint64 {
	if x != nil {
		return x.Limit
	}
	return 0
}

func (x *HeadFileRequest) GetNextPageToken() int64 {
	if x != nil {
		return x.NextPageToken
	}
	return 0
}

type HeadFileResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Info          []*FileInfo `protobuf:"bytes,1,rep,name=info,proto3" json:"info,omitempty"`
	NextPageToken int64       `protobuf:"varint,2,opt,name=next_page_token,json=nextPageToken,proto3" json:"next_page_token,omitempty"`
}

func (x *HeadFileResponse) Reset() {
	*x = HeadFileResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_protos_upload_upload_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HeadFileResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HeadFileResponse) ProtoMessage() {}

func (x *HeadFileResponse) ProtoReflect() protoreflect.Message {
	mi := &file_internal_protos_upload_upload_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HeadFileResponse.ProtoReflect.Descriptor instead.
func (*HeadFileResponse) Descriptor() ([]byte, []int) {
	return file_internal_protos_upload_upload_proto_rawDescGZIP(), []int{7}
}

func (x *HeadFileResponse) GetInfo() []*FileInfo {
	if x != nil {
		return x.Info
	}
	return nil
}

func (x *HeadFileResponse) GetNextPageToken() int64 {
	if x != nil {
		return x.NextPageToken
	}
	return 0
}

type Files struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Infos []*FileInfo `protobuf:"bytes,1,rep,name=infos,proto3" json:"infos,omitempty"`
}

func (x *Files) Reset() {
	*x = Files{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_protos_upload_upload_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Files) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Files) ProtoMessage() {}

func (x *Files) ProtoReflect() protoreflect.Message {
	mi := &file_internal_protos_upload_upload_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Files.ProtoReflect.Descriptor instead.
func (*Files) Descriptor() ([]byte, []int) {
	return file_internal_protos_upload_upload_proto_rawDescGZIP(), []int{8}
}

func (x *Files) GetInfos() []*FileInfo {
	if x != nil {
		return x.Infos
	}
	return nil
}

type FileInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name      string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	CreatedAt int64  `protobuf:"varint,2,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	Size      uint64 `protobuf:"varint,3,opt,name=size,proto3" json:"size,omitempty"`
	Sum       []byte `protobuf:"bytes,4,opt,name=sum,proto3" json:"sum,omitempty"`
}

func (x *FileInfo) Reset() {
	*x = FileInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_protos_upload_upload_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FileInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FileInfo) ProtoMessage() {}

func (x *FileInfo) ProtoReflect() protoreflect.Message {
	mi := &file_internal_protos_upload_upload_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FileInfo.ProtoReflect.Descriptor instead.
func (*FileInfo) Descriptor() ([]byte, []int) {
	return file_internal_protos_upload_upload_proto_rawDescGZIP(), []int{9}
}

func (x *FileInfo) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *FileInfo) GetCreatedAt() int64 {
	if x != nil {
		return x.CreatedAt
	}
	return 0
}

func (x *FileInfo) GetSize() uint64 {
	if x != nil {
		return x.Size
	}
	return 0
}

func (x *FileInfo) GetSum() []byte {
	if x != nil {
		return x.Sum
	}
	return nil
}

type Empty struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Empty) Reset() {
	*x = Empty{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_protos_upload_upload_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Empty) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Empty) ProtoMessage() {}

func (x *Empty) ProtoReflect() protoreflect.Message {
	mi := &file_internal_protos_upload_upload_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Empty.ProtoReflect.Descriptor instead.
func (*Empty) Descriptor() ([]byte, []int) {
	return file_internal_protos_upload_upload_proto_rawDescGZIP(), []int{10}
}

type Filename struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *Filename) Reset() {
	*x = Filename{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_protos_upload_upload_proto_msgTypes[11]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Filename) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Filename) ProtoMessage() {}

func (x *Filename) ProtoReflect() protoreflect.Message {
	mi := &file_internal_protos_upload_upload_proto_msgTypes[11]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Filename.ProtoReflect.Descriptor instead.
func (*Filename) Descriptor() ([]byte, []int) {
	return file_internal_protos_upload_upload_proto_rawDescGZIP(), []int{11}
}

func (x *Filename) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type SectionChunk struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Sequence    uint64 `protobuf:"varint,1,opt,name=sequence,proto3" json:"sequence,omitempty"`
	Size        uint64 `protobuf:"varint,2,opt,name=size,proto3" json:"size,omitempty"`
	Sum         []byte `protobuf:"bytes,3,opt,name=sum,proto3" json:"sum,omitempty"`
	BlockOffset uint64 `protobuf:"varint,4,opt,name=block_offset,json=blockOffset,proto3" json:"block_offset,omitempty"`
}

func (x *SectionChunk) Reset() {
	*x = SectionChunk{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_protos_upload_upload_proto_msgTypes[12]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SectionChunk) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SectionChunk) ProtoMessage() {}

func (x *SectionChunk) ProtoReflect() protoreflect.Message {
	mi := &file_internal_protos_upload_upload_proto_msgTypes[12]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SectionChunk.ProtoReflect.Descriptor instead.
func (*SectionChunk) Descriptor() ([]byte, []int) {
	return file_internal_protos_upload_upload_proto_rawDescGZIP(), []int{12}
}

func (x *SectionChunk) GetSequence() uint64 {
	if x != nil {
		return x.Sequence
	}
	return 0
}

func (x *SectionChunk) GetSize() uint64 {
	if x != nil {
		return x.Size
	}
	return 0
}

func (x *SectionChunk) GetSum() []byte {
	if x != nil {
		return x.Sum
	}
	return nil
}

func (x *SectionChunk) GetBlockOffset() uint64 {
	if x != nil {
		return x.BlockOffset
	}
	return 0
}

type Section struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Chunks     []*SectionChunk `protobuf:"bytes,1,rep,name=chunks,proto3" json:"chunks,omitempty"`
	Url        string          `protobuf:"bytes,2,opt,name=url,proto3" json:"url,omitempty"`
	RangeStart uint64          `protobuf:"varint,3,opt,name=range_start,json=rangeStart,proto3" json:"range_start,omitempty"`
	RangeEnd   uint64          `protobuf:"varint,4,opt,name=range_end,json=rangeEnd,proto3" json:"range_end,omitempty"`
}

func (x *Section) Reset() {
	*x = Section{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_protos_upload_upload_proto_msgTypes[13]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Section) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Section) ProtoMessage() {}

func (x *Section) ProtoReflect() protoreflect.Message {
	mi := &file_internal_protos_upload_upload_proto_msgTypes[13]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Section.ProtoReflect.Descriptor instead.
func (*Section) Descriptor() ([]byte, []int) {
	return file_internal_protos_upload_upload_proto_rawDescGZIP(), []int{13}
}

func (x *Section) GetChunks() []*SectionChunk {
	if x != nil {
		return x.Chunks
	}
	return nil
}

func (x *Section) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

func (x *Section) GetRangeStart() uint64 {
	if x != nil {
		return x.RangeStart
	}
	return 0
}

func (x *Section) GetRangeEnd() uint64 {
	if x != nil {
		return x.RangeEnd
	}
	return 0
}

type DownloadResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Sections []*Section `protobuf:"bytes,1,rep,name=sections,proto3" json:"sections,omitempty"`
}

func (x *DownloadResponse) Reset() {
	*x = DownloadResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_protos_upload_upload_proto_msgTypes[14]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DownloadResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DownloadResponse) ProtoMessage() {}

func (x *DownloadResponse) ProtoReflect() protoreflect.Message {
	mi := &file_internal_protos_upload_upload_proto_msgTypes[14]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DownloadResponse.ProtoReflect.Descriptor instead.
func (*DownloadResponse) Descriptor() ([]byte, []int) {
	return file_internal_protos_upload_upload_proto_rawDescGZIP(), []int{14}
}

func (x *DownloadResponse) GetSections() []*Section {
	if x != nil {
		return x.Sections
	}
	return nil
}

var File_internal_protos_upload_upload_proto protoreflect.FileDescriptor

var file_internal_protos_upload_upload_proto_rawDesc = []byte{
	0x0a, 0x23, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x73, 0x2f, 0x75, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x2f, 0x75, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06, 0x75, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x22, 0x28, 0x0a,
	0x12, 0x43, 0x68, 0x75, 0x6e, 0x6b, 0x73, 0x45, 0x78, 0x69, 0x73, 0x74, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x73, 0x75, 0x6d, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28,
	0x0c, 0x52, 0x04, 0x73, 0x75, 0x6d, 0x73, 0x22, 0x2d, 0x0a, 0x13, 0x43, 0x68, 0x75, 0x6e, 0x6b,
	0x73, 0x45, 0x78, 0x69, 0x73, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x16,
	0x0a, 0x06, 0x65, 0x78, 0x69, 0x73, 0x74, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x08, 0x52, 0x06,
	0x65, 0x78, 0x69, 0x73, 0x74, 0x73, 0x22, 0x2e, 0x0a, 0x04, 0x46, 0x69, 0x6c, 0x65, 0x12, 0x12,
	0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61,
	0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x73, 0x75, 0x6d, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0c,
	0x52, 0x04, 0x73, 0x75, 0x6d, 0x73, 0x22, 0x36, 0x0a, 0x0b, 0x43, 0x6f, 0x70, 0x79, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x15, 0x0a, 0x06, 0x73, 0x72, 0x63, 0x5f, 0x69, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05, 0x73, 0x72, 0x63, 0x49, 0x64, 0x12, 0x10, 0x0a, 0x03,
	0x64, 0x73, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x64, 0x73, 0x74, 0x22, 0x1a,
	0x0a, 0x06, 0x46, 0x69, 0x6c, 0x65, 0x49, 0x44, 0x12, 0x10, 0x0a, 0x03, 0x73, 0x75, 0x6d, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x03, 0x73, 0x75, 0x6d, 0x22, 0x20, 0x0a, 0x06, 0x50, 0x72,
	0x65, 0x66, 0x69, 0x78, 0x12, 0x16, 0x0a, 0x06, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78, 0x22, 0x63, 0x0a, 0x0f,
	0x48, 0x65, 0x61, 0x64, 0x46, 0x69, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e,
	0x61, 0x6d, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x6c, 0x69, 0x6d, 0x69, 0x74, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x04, 0x52, 0x05, 0x6c, 0x69, 0x6d, 0x69, 0x74, 0x12, 0x26, 0x0a, 0x0f, 0x6e, 0x65, 0x78,
	0x74, 0x5f, 0x70, 0x61, 0x67, 0x65, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x03, 0x52, 0x0d, 0x6e, 0x65, 0x78, 0x74, 0x50, 0x61, 0x67, 0x65, 0x54, 0x6f, 0x6b, 0x65,
	0x6e, 0x22, 0x60, 0x0a, 0x10, 0x48, 0x65, 0x61, 0x64, 0x46, 0x69, 0x6c, 0x65, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x24, 0x0a, 0x04, 0x69, 0x6e, 0x66, 0x6f, 0x18, 0x01, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x75, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x2e, 0x46, 0x69, 0x6c,
	0x65, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x04, 0x69, 0x6e, 0x66, 0x6f, 0x12, 0x26, 0x0a, 0x0f, 0x6e,
	0x65, 0x78, 0x74, 0x5f, 0x70, 0x61, 0x67, 0x65, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x0d, 0x6e, 0x65, 0x78, 0x74, 0x50, 0x61, 0x67, 0x65, 0x54, 0x6f,
	0x6b, 0x65, 0x6e, 0x22, 0x2f, 0x0a, 0x05, 0x46, 0x69, 0x6c, 0x65, 0x73, 0x12, 0x26, 0x0a, 0x05,
	0x69, 0x6e, 0x66, 0x6f, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x75, 0x70,
	0x6c, 0x6f, 0x61, 0x64, 0x2e, 0x46, 0x69, 0x6c, 0x65, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x05, 0x69,
	0x6e, 0x66, 0x6f, 0x73, 0x22, 0x63, 0x0a, 0x08, 0x46, 0x69, 0x6c, 0x65, 0x49, 0x6e, 0x66, 0x6f,
	0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04,
	0x6e, 0x61, 0x6d, 0x65, 0x12, 0x1d, 0x0a, 0x0a, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x5f,
	0x61, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x09, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65,
	0x64, 0x41, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x04, 0x52, 0x04, 0x73, 0x69, 0x7a, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x73, 0x75, 0x6d, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x03, 0x73, 0x75, 0x6d, 0x22, 0x07, 0x0a, 0x05, 0x45, 0x6d, 0x70,
	0x74, 0x79, 0x22, 0x1e, 0x0a, 0x08, 0x46, 0x69, 0x6c, 0x65, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x12,
	0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61,
	0x6d, 0x65, 0x22, 0x73, 0x0a, 0x0c, 0x53, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x43, 0x68, 0x75,
	0x6e, 0x6b, 0x12, 0x1a, 0x0a, 0x08, 0x73, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x04, 0x52, 0x08, 0x73, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x65, 0x12, 0x12,
	0x0a, 0x04, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x04, 0x73, 0x69,
	0x7a, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x73, 0x75, 0x6d, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52,
	0x03, 0x73, 0x75, 0x6d, 0x12, 0x21, 0x0a, 0x0c, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x5f, 0x6f, 0x66,
	0x66, 0x73, 0x65, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0b, 0x62, 0x6c, 0x6f, 0x63,
	0x6b, 0x4f, 0x66, 0x66, 0x73, 0x65, 0x74, 0x22, 0x87, 0x01, 0x0a, 0x07, 0x53, 0x65, 0x63, 0x74,
	0x69, 0x6f, 0x6e, 0x12, 0x2c, 0x0a, 0x06, 0x63, 0x68, 0x75, 0x6e, 0x6b, 0x73, 0x18, 0x01, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x75, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x2e, 0x53, 0x65, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x43, 0x68, 0x75, 0x6e, 0x6b, 0x52, 0x06, 0x63, 0x68, 0x75, 0x6e, 0x6b,
	0x73, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x72, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03,
	0x75, 0x72, 0x6c, 0x12, 0x1f, 0x0a, 0x0b, 0x72, 0x61, 0x6e, 0x67, 0x65, 0x5f, 0x73, 0x74, 0x61,
	0x72, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0a, 0x72, 0x61, 0x6e, 0x67, 0x65, 0x53,
	0x74, 0x61, 0x72, 0x74, 0x12, 0x1b, 0x0a, 0x09, 0x72, 0x61, 0x6e, 0x67, 0x65, 0x5f, 0x65, 0x6e,
	0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x04, 0x52, 0x08, 0x72, 0x61, 0x6e, 0x67, 0x65, 0x45, 0x6e,
	0x64, 0x22, 0x3f, 0x0a, 0x10, 0x44, 0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2b, 0x0a, 0x08, 0x73, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x75, 0x70, 0x6c, 0x6f, 0x61, 0x64,
	0x2e, 0x53, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x08, 0x73, 0x65, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x32, 0xca, 0x02, 0x0a, 0x06, 0x49, 0x6f, 0x74, 0x61, 0x46, 0x53, 0x12, 0x46, 0x0a,
	0x0b, 0x43, 0x68, 0x75, 0x6e, 0x6b, 0x73, 0x45, 0x78, 0x69, 0x73, 0x74, 0x12, 0x1a, 0x2e, 0x75,
	0x70, 0x6c, 0x6f, 0x61, 0x64, 0x2e, 0x43, 0x68, 0x75, 0x6e, 0x6b, 0x73, 0x45, 0x78, 0x69, 0x73,
	0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1b, 0x2e, 0x75, 0x70, 0x6c, 0x6f, 0x61,
	0x64, 0x2e, 0x43, 0x68, 0x75, 0x6e, 0x6b, 0x73, 0x45, 0x78, 0x69, 0x73, 0x74, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2a, 0x0a, 0x0a, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x46,
	0x69, 0x6c, 0x65, 0x12, 0x0c, 0x2e, 0x75, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x2e, 0x46, 0x69, 0x6c,
	0x65, 0x1a, 0x0e, 0x2e, 0x75, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x2e, 0x46, 0x69, 0x6c, 0x65, 0x49,
	0x44, 0x12, 0x2a, 0x0a, 0x09, 0x4c, 0x69, 0x73, 0x74, 0x46, 0x69, 0x6c, 0x65, 0x73, 0x12, 0x0e,
	0x2e, 0x75, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x2e, 0x50, 0x72, 0x65, 0x66, 0x69, 0x78, 0x1a, 0x0d,
	0x2e, 0x75, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x2e, 0x46, 0x69, 0x6c, 0x65, 0x73, 0x12, 0x3d, 0x0a,
	0x08, 0x48, 0x65, 0x61, 0x64, 0x46, 0x69, 0x6c, 0x65, 0x12, 0x17, 0x2e, 0x75, 0x70, 0x6c, 0x6f,
	0x61, 0x64, 0x2e, 0x48, 0x65, 0x61, 0x64, 0x46, 0x69, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x18, 0x2e, 0x75, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x2e, 0x48, 0x65, 0x61, 0x64,
	0x46, 0x69, 0x6c, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x34, 0x0a, 0x08,
	0x44, 0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64, 0x12, 0x0e, 0x2e, 0x75, 0x70, 0x6c, 0x6f, 0x61,
	0x64, 0x2e, 0x46, 0x69, 0x6c, 0x65, 0x49, 0x44, 0x1a, 0x18, 0x2e, 0x75, 0x70, 0x6c, 0x6f, 0x61,
	0x64, 0x2e, 0x44, 0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x2b, 0x0a, 0x04, 0x43, 0x6f, 0x70, 0x79, 0x12, 0x13, 0x2e, 0x75, 0x70, 0x6c,
	0x6f, 0x61, 0x64, 0x2e, 0x43, 0x6f, 0x70, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x0e, 0x2e, 0x75, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x2e, 0x46, 0x69, 0x6c, 0x65, 0x49, 0x44, 0x42,
	0x18, 0x5a, 0x16, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x73, 0x2f, 0x75, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_internal_protos_upload_upload_proto_rawDescOnce sync.Once
	file_internal_protos_upload_upload_proto_rawDescData = file_internal_protos_upload_upload_proto_rawDesc
)

func file_internal_protos_upload_upload_proto_rawDescGZIP() []byte {
	file_internal_protos_upload_upload_proto_rawDescOnce.Do(func() {
		file_internal_protos_upload_upload_proto_rawDescData = protoimpl.X.CompressGZIP(file_internal_protos_upload_upload_proto_rawDescData)
	})
	return file_internal_protos_upload_upload_proto_rawDescData
}

var file_internal_protos_upload_upload_proto_msgTypes = make([]protoimpl.MessageInfo, 15)
var file_internal_protos_upload_upload_proto_goTypes = []interface{}{
	(*ChunksExistRequest)(nil),  // 0: upload.ChunksExistRequest
	(*ChunksExistResponse)(nil), // 1: upload.ChunksExistResponse
	(*File)(nil),                // 2: upload.File
	(*CopyRequest)(nil),         // 3: upload.CopyRequest
	(*FileID)(nil),              // 4: upload.FileID
	(*Prefix)(nil),              // 5: upload.Prefix
	(*HeadFileRequest)(nil),     // 6: upload.HeadFileRequest
	(*HeadFileResponse)(nil),    // 7: upload.HeadFileResponse
	(*Files)(nil),               // 8: upload.Files
	(*FileInfo)(nil),            // 9: upload.FileInfo
	(*Empty)(nil),               // 10: upload.Empty
	(*Filename)(nil),            // 11: upload.Filename
	(*SectionChunk)(nil),        // 12: upload.SectionChunk
	(*Section)(nil),             // 13: upload.Section
	(*DownloadResponse)(nil),    // 14: upload.DownloadResponse
}
var file_internal_protos_upload_upload_proto_depIdxs = []int32{
	9,  // 0: upload.HeadFileResponse.info:type_name -> upload.FileInfo
	9,  // 1: upload.Files.infos:type_name -> upload.FileInfo
	12, // 2: upload.Section.chunks:type_name -> upload.SectionChunk
	13, // 3: upload.DownloadResponse.sections:type_name -> upload.Section
	0,  // 4: upload.IotaFS.ChunksExist:input_type -> upload.ChunksExistRequest
	2,  // 5: upload.IotaFS.CreateFile:input_type -> upload.File
	5,  // 6: upload.IotaFS.ListFiles:input_type -> upload.Prefix
	6,  // 7: upload.IotaFS.HeadFile:input_type -> upload.HeadFileRequest
	4,  // 8: upload.IotaFS.Download:input_type -> upload.FileID
	3,  // 9: upload.IotaFS.Copy:input_type -> upload.CopyRequest
	1,  // 10: upload.IotaFS.ChunksExist:output_type -> upload.ChunksExistResponse
	4,  // 11: upload.IotaFS.CreateFile:output_type -> upload.FileID
	8,  // 12: upload.IotaFS.ListFiles:output_type -> upload.Files
	7,  // 13: upload.IotaFS.HeadFile:output_type -> upload.HeadFileResponse
	14, // 14: upload.IotaFS.Download:output_type -> upload.DownloadResponse
	4,  // 15: upload.IotaFS.Copy:output_type -> upload.FileID
	10, // [10:16] is the sub-list for method output_type
	4,  // [4:10] is the sub-list for method input_type
	4,  // [4:4] is the sub-list for extension type_name
	4,  // [4:4] is the sub-list for extension extendee
	0,  // [0:4] is the sub-list for field type_name
}

func init() { file_internal_protos_upload_upload_proto_init() }
func file_internal_protos_upload_upload_proto_init() {
	if File_internal_protos_upload_upload_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_internal_protos_upload_upload_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ChunksExistRequest); i {
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
		file_internal_protos_upload_upload_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ChunksExistResponse); i {
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
		file_internal_protos_upload_upload_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*File); i {
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
		file_internal_protos_upload_upload_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CopyRequest); i {
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
		file_internal_protos_upload_upload_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FileID); i {
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
		file_internal_protos_upload_upload_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Prefix); i {
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
		file_internal_protos_upload_upload_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*HeadFileRequest); i {
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
		file_internal_protos_upload_upload_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*HeadFileResponse); i {
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
		file_internal_protos_upload_upload_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Files); i {
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
		file_internal_protos_upload_upload_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FileInfo); i {
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
		file_internal_protos_upload_upload_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Empty); i {
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
		file_internal_protos_upload_upload_proto_msgTypes[11].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Filename); i {
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
		file_internal_protos_upload_upload_proto_msgTypes[12].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SectionChunk); i {
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
		file_internal_protos_upload_upload_proto_msgTypes[13].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Section); i {
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
		file_internal_protos_upload_upload_proto_msgTypes[14].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DownloadResponse); i {
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
			RawDescriptor: file_internal_protos_upload_upload_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   15,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_internal_protos_upload_upload_proto_goTypes,
		DependencyIndexes: file_internal_protos_upload_upload_proto_depIdxs,
		MessageInfos:      file_internal_protos_upload_upload_proto_msgTypes,
	}.Build()
	File_internal_protos_upload_upload_proto = out.File
	file_internal_protos_upload_upload_proto_rawDesc = nil
	file_internal_protos_upload_upload_proto_goTypes = nil
	file_internal_protos_upload_upload_proto_depIdxs = nil
}
