// Code generated by protoc-gen-go. DO NOT EDIT.
// source: pkg/catalogapi/catalog.proto

/*
Package catalogapi is a generated protocol buffer package.

It is generated from these files:
	pkg/catalogapi/catalog.proto

It has these top-level messages:
	PutRequest
	PutResponse
	SearchRequest
	SearchResponse
	PublicationReceipt
	Date
*/
package catalogapi

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type PutRequest struct {
	Pub *PublicationReceipt `protobuf:"bytes,1,opt,name=pub" json:"pub,omitempty"`
}

func (m *PutRequest) Reset()                    { *m = PutRequest{} }
func (m *PutRequest) String() string            { return proto.CompactTextString(m) }
func (*PutRequest) ProtoMessage()               {}
func (*PutRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *PutRequest) GetPub() *PublicationReceipt {
	if m != nil {
		return m.Pub
	}
	return nil
}

type PutResponse struct {
}

func (m *PutResponse) Reset()                    { *m = PutResponse{} }
func (m *PutResponse) String() string            { return proto.CompactTextString(m) }
func (*PutResponse) ProtoMessage()               {}
func (*PutResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

type SearchRequest struct {
	EntryKey        []byte `protobuf:"bytes,1,opt,name=entry_key,json=entryKey,proto3" json:"entry_key,omitempty"`
	AuthorPublicKey []byte `protobuf:"bytes,2,opt,name=author_public_key,json=authorPublicKey,proto3" json:"author_public_key,omitempty"`
	ReaderPublicKey []byte `protobuf:"bytes,3,opt,name=reader_public_key,json=readerPublicKey,proto3" json:"reader_public_key,omitempty"`
	Before          int64  `protobuf:"varint,4,opt,name=before" json:"before,omitempty"`
	Limit           uint32 `protobuf:"varint,5,opt,name=limit" json:"limit,omitempty"`
}

func (m *SearchRequest) Reset()                    { *m = SearchRequest{} }
func (m *SearchRequest) String() string            { return proto.CompactTextString(m) }
func (*SearchRequest) ProtoMessage()               {}
func (*SearchRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *SearchRequest) GetEntryKey() []byte {
	if m != nil {
		return m.EntryKey
	}
	return nil
}

func (m *SearchRequest) GetAuthorPublicKey() []byte {
	if m != nil {
		return m.AuthorPublicKey
	}
	return nil
}

func (m *SearchRequest) GetReaderPublicKey() []byte {
	if m != nil {
		return m.ReaderPublicKey
	}
	return nil
}

func (m *SearchRequest) GetBefore() int64 {
	if m != nil {
		return m.Before
	}
	return 0
}

func (m *SearchRequest) GetLimit() uint32 {
	if m != nil {
		return m.Limit
	}
	return 0
}

type SearchResponse struct {
	Result []*PublicationReceipt `protobuf:"bytes,1,rep,name=result" json:"result,omitempty"`
}

func (m *SearchResponse) Reset()                    { *m = SearchResponse{} }
func (m *SearchResponse) String() string            { return proto.CompactTextString(m) }
func (*SearchResponse) ProtoMessage()               {}
func (*SearchResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *SearchResponse) GetResult() []*PublicationReceipt {
	if m != nil {
		return m.Result
	}
	return nil
}

// Publication is a libri publication, intentionally defined separately to give us the flexibility
// down the road to add things to one and not the other.
type PublicationReceipt struct {
	EnvelopeKey     []byte `protobuf:"bytes,1,opt,name=envelope_key,json=envelopeKey,proto3" json:"envelope_key,omitempty"`
	EntryKey        []byte `protobuf:"bytes,2,opt,name=entry_key,json=entryKey,proto3" json:"entry_key,omitempty"`
	AuthorPublicKey []byte `protobuf:"bytes,3,opt,name=author_public_key,json=authorPublicKey,proto3" json:"author_public_key,omitempty"`
	ReaderPublicKey []byte `protobuf:"bytes,4,opt,name=reader_public_key,json=readerPublicKey,proto3" json:"reader_public_key,omitempty"`
	// received_time is epoch-time (microseconds since Jan 1, 1970) when the publication was
	// received.
	ReceivedTime int64 `protobuf:"varint,5,opt,name=received_time,json=receivedTime" json:"received_time,omitempty"`
}

func (m *PublicationReceipt) Reset()                    { *m = PublicationReceipt{} }
func (m *PublicationReceipt) String() string            { return proto.CompactTextString(m) }
func (*PublicationReceipt) ProtoMessage()               {}
func (*PublicationReceipt) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *PublicationReceipt) GetEnvelopeKey() []byte {
	if m != nil {
		return m.EnvelopeKey
	}
	return nil
}

func (m *PublicationReceipt) GetEntryKey() []byte {
	if m != nil {
		return m.EntryKey
	}
	return nil
}

func (m *PublicationReceipt) GetAuthorPublicKey() []byte {
	if m != nil {
		return m.AuthorPublicKey
	}
	return nil
}

func (m *PublicationReceipt) GetReaderPublicKey() []byte {
	if m != nil {
		return m.ReaderPublicKey
	}
	return nil
}

func (m *PublicationReceipt) GetReceivedTime() int64 {
	if m != nil {
		return m.ReceivedTime
	}
	return 0
}

// Date is a straightforward date.
type Date struct {
	Year  int32 `protobuf:"varint,1,opt,name=year" json:"year,omitempty"`
	Month int32 `protobuf:"varint,2,opt,name=month" json:"month,omitempty"`
	Day   int32 `protobuf:"varint,3,opt,name=day" json:"day,omitempty"`
}

func (m *Date) Reset()                    { *m = Date{} }
func (m *Date) String() string            { return proto.CompactTextString(m) }
func (*Date) ProtoMessage()               {}
func (*Date) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *Date) GetYear() int32 {
	if m != nil {
		return m.Year
	}
	return 0
}

func (m *Date) GetMonth() int32 {
	if m != nil {
		return m.Month
	}
	return 0
}

func (m *Date) GetDay() int32 {
	if m != nil {
		return m.Day
	}
	return 0
}

func init() {
	proto.RegisterType((*PutRequest)(nil), "catalogapi.PutRequest")
	proto.RegisterType((*PutResponse)(nil), "catalogapi.PutResponse")
	proto.RegisterType((*SearchRequest)(nil), "catalogapi.SearchRequest")
	proto.RegisterType((*SearchResponse)(nil), "catalogapi.SearchResponse")
	proto.RegisterType((*PublicationReceipt)(nil), "catalogapi.PublicationReceipt")
	proto.RegisterType((*Date)(nil), "catalogapi.Date")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Catalog service

type CatalogClient interface {
	Put(ctx context.Context, in *PutRequest, opts ...grpc.CallOption) (*PutResponse, error)
	Search(ctx context.Context, in *SearchRequest, opts ...grpc.CallOption) (*SearchResponse, error)
}

type catalogClient struct {
	cc *grpc.ClientConn
}

func NewCatalogClient(cc *grpc.ClientConn) CatalogClient {
	return &catalogClient{cc}
}

func (c *catalogClient) Put(ctx context.Context, in *PutRequest, opts ...grpc.CallOption) (*PutResponse, error) {
	out := new(PutResponse)
	err := grpc.Invoke(ctx, "/catalogapi.Catalog/Put", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *catalogClient) Search(ctx context.Context, in *SearchRequest, opts ...grpc.CallOption) (*SearchResponse, error) {
	out := new(SearchResponse)
	err := grpc.Invoke(ctx, "/catalogapi.Catalog/Search", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Catalog service

type CatalogServer interface {
	Put(context.Context, *PutRequest) (*PutResponse, error)
	Search(context.Context, *SearchRequest) (*SearchResponse, error)
}

func RegisterCatalogServer(s *grpc.Server, srv CatalogServer) {
	s.RegisterService(&_Catalog_serviceDesc, srv)
}

func _Catalog_Put_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PutRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CatalogServer).Put(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/catalogapi.Catalog/Put",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CatalogServer).Put(ctx, req.(*PutRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Catalog_Search_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SearchRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CatalogServer).Search(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/catalogapi.Catalog/Search",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CatalogServer).Search(ctx, req.(*SearchRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Catalog_serviceDesc = grpc.ServiceDesc{
	ServiceName: "catalogapi.Catalog",
	HandlerType: (*CatalogServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Put",
			Handler:    _Catalog_Put_Handler,
		},
		{
			MethodName: "Search",
			Handler:    _Catalog_Search_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pkg/catalogapi/catalog.proto",
}

func init() { proto.RegisterFile("pkg/catalogapi/catalog.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 399 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x93, 0xcf, 0x8e, 0xd3, 0x30,
	0x10, 0xc6, 0xf1, 0x3a, 0x09, 0x30, 0x6d, 0xf8, 0x63, 0xa1, 0x25, 0x14, 0x84, 0x42, 0xb8, 0x54,
	0x1c, 0x0a, 0x5a, 0x24, 0xc4, 0x09, 0x89, 0x3f, 0x07, 0x24, 0x2e, 0x95, 0xe1, 0x5e, 0x39, 0xe9,
	0xb0, 0xb5, 0x36, 0x89, 0x8d, 0xe3, 0xac, 0x94, 0x17, 0xe0, 0x81, 0x78, 0x13, 0xde, 0x08, 0xc5,
	0x4e, 0xd4, 0x8d, 0xda, 0x43, 0xb9, 0xcd, 0xcc, 0xf7, 0x75, 0xfa, 0xe5, 0x97, 0x09, 0x3c, 0xd3,
	0x57, 0x97, 0xaf, 0x0b, 0x61, 0x45, 0xa9, 0x2e, 0x85, 0x96, 0x63, 0xb9, 0xd2, 0x46, 0x59, 0xc5,
	0x60, 0xaf, 0x64, 0x1f, 0x00, 0xd6, 0xad, 0xe5, 0xf8, 0xab, 0xc5, 0xc6, 0xb2, 0x37, 0x40, 0x75,
	0x9b, 0x27, 0x24, 0x25, 0xcb, 0xd9, 0xc5, 0xf3, 0xd5, 0xde, 0xb7, 0x5a, 0xb7, 0x79, 0x29, 0x0b,
	0x61, 0xa5, 0xaa, 0x39, 0x16, 0x28, 0xb5, 0xe5, 0xbd, 0x35, 0x8b, 0x61, 0xe6, 0x7e, 0xdf, 0x68,
	0x55, 0x37, 0x98, 0xfd, 0x21, 0x10, 0x7f, 0x47, 0x61, 0x8a, 0xdd, 0xb8, 0xf2, 0x29, 0xdc, 0xc5,
	0xda, 0x9a, 0x6e, 0x73, 0x85, 0x9d, 0x5b, 0x3c, 0xe7, 0x77, 0xdc, 0xe0, 0x1b, 0x76, 0xec, 0x15,
	0x3c, 0x14, 0xad, 0xdd, 0x29, 0xb3, 0xd1, 0x6e, 0xbf, 0x33, 0x9d, 0x39, 0xd3, 0x7d, 0x2f, 0xf8,
	0xff, 0x1d, 0xbc, 0x06, 0xc5, 0x16, 0x27, 0x5e, 0xea, 0xbd, 0x5e, 0xd8, 0x7b, 0xcf, 0x21, 0xca,
	0xf1, 0xa7, 0x32, 0x98, 0x04, 0x29, 0x59, 0x52, 0x3e, 0x74, 0xec, 0x11, 0x84, 0xa5, 0xac, 0xa4,
	0x4d, 0xc2, 0x94, 0x2c, 0x63, 0xee, 0x9b, 0xec, 0x2b, 0xdc, 0x1b, 0x33, 0xfb, 0xc7, 0x60, 0xef,
	0x20, 0x32, 0xd8, 0xb4, 0xa5, 0x4d, 0x48, 0x4a, 0x4f, 0x40, 0x31, 0xb8, 0xb3, 0xbf, 0x04, 0xd8,
	0xa1, 0xcc, 0x5e, 0xc0, 0x1c, 0xeb, 0x6b, 0x2c, 0x95, 0xc6, 0x1b, 0x18, 0x66, 0xe3, 0xac, 0x4f,
	0x3c, 0xc1, 0x74, 0x76, 0x0a, 0x26, 0xfa, 0x1f, 0x98, 0x82, 0xe3, 0x98, 0x5e, 0x42, 0x6c, 0xfa,
	0x88, 0xd7, 0xb8, 0xdd, 0x58, 0x59, 0xa1, 0xc3, 0x42, 0xf9, 0x7c, 0x1c, 0xfe, 0x90, 0x15, 0x66,
	0x9f, 0x20, 0xf8, 0x22, 0x2c, 0x32, 0x06, 0x41, 0x87, 0xc2, 0xb8, 0xf0, 0x21, 0x77, 0x75, 0xcf,
	0xb3, 0x52, 0xb5, 0xdd, 0xb9, 0xc4, 0x21, 0xf7, 0x0d, 0x7b, 0x00, 0x74, 0x2b, 0x7c, 0xc0, 0x90,
	0xf7, 0xe5, 0xc5, 0x6f, 0x02, 0xb7, 0x3f, 0x7b, 0x82, 0xec, 0x3d, 0xd0, 0x75, 0x6b, 0xd9, 0xf9,
	0x14, 0xe9, 0x78, 0x82, 0x8b, 0xc7, 0x07, 0xf3, 0xe1, 0xb4, 0x6e, 0xb1, 0x8f, 0x10, 0xf9, 0xf7,
	0xc4, 0x9e, 0xdc, 0x34, 0x4d, 0xee, 0x6d, 0xb1, 0x38, 0x26, 0x8d, 0x2b, 0xf2, 0xc8, 0x7d, 0x01,
	0x6f, 0xff, 0x05, 0x00, 0x00, 0xff, 0xff, 0xba, 0x9a, 0x1c, 0xe3, 0x21, 0x03, 0x00, 0x00,
}
