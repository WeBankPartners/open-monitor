// Code generated by protoc-gen-go. DO NOT EDIT.
// source: xattr.proto

package hadoop_hdfs

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type XAttrSetFlagProto int32

const (
	XAttrSetFlagProto_XATTR_CREATE  XAttrSetFlagProto = 1
	XAttrSetFlagProto_XATTR_REPLACE XAttrSetFlagProto = 2
)

var XAttrSetFlagProto_name = map[int32]string{
	1: "XATTR_CREATE",
	2: "XATTR_REPLACE",
}
var XAttrSetFlagProto_value = map[string]int32{
	"XATTR_CREATE":  1,
	"XATTR_REPLACE": 2,
}

func (x XAttrSetFlagProto) Enum() *XAttrSetFlagProto {
	p := new(XAttrSetFlagProto)
	*p = x
	return p
}
func (x XAttrSetFlagProto) String() string {
	return proto.EnumName(XAttrSetFlagProto_name, int32(x))
}
func (x *XAttrSetFlagProto) UnmarshalJSON(data []byte) error {
	value, err := proto.UnmarshalJSONEnum(XAttrSetFlagProto_value, data, "XAttrSetFlagProto")
	if err != nil {
		return err
	}
	*x = XAttrSetFlagProto(value)
	return nil
}
func (XAttrSetFlagProto) EnumDescriptor() ([]byte, []int) { return fileDescriptor1, []int{0} }

type XAttrProto_XAttrNamespaceProto int32

const (
	XAttrProto_USER     XAttrProto_XAttrNamespaceProto = 0
	XAttrProto_TRUSTED  XAttrProto_XAttrNamespaceProto = 1
	XAttrProto_SECURITY XAttrProto_XAttrNamespaceProto = 2
	XAttrProto_SYSTEM   XAttrProto_XAttrNamespaceProto = 3
	XAttrProto_RAW      XAttrProto_XAttrNamespaceProto = 4
)

var XAttrProto_XAttrNamespaceProto_name = map[int32]string{
	0: "USER",
	1: "TRUSTED",
	2: "SECURITY",
	3: "SYSTEM",
	4: "RAW",
}
var XAttrProto_XAttrNamespaceProto_value = map[string]int32{
	"USER":     0,
	"TRUSTED":  1,
	"SECURITY": 2,
	"SYSTEM":   3,
	"RAW":      4,
}

func (x XAttrProto_XAttrNamespaceProto) Enum() *XAttrProto_XAttrNamespaceProto {
	p := new(XAttrProto_XAttrNamespaceProto)
	*p = x
	return p
}
func (x XAttrProto_XAttrNamespaceProto) String() string {
	return proto.EnumName(XAttrProto_XAttrNamespaceProto_name, int32(x))
}
func (x *XAttrProto_XAttrNamespaceProto) UnmarshalJSON(data []byte) error {
	value, err := proto.UnmarshalJSONEnum(XAttrProto_XAttrNamespaceProto_value, data, "XAttrProto_XAttrNamespaceProto")
	if err != nil {
		return err
	}
	*x = XAttrProto_XAttrNamespaceProto(value)
	return nil
}
func (XAttrProto_XAttrNamespaceProto) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor1, []int{0, 0}
}

type XAttrProto struct {
	Namespace        *XAttrProto_XAttrNamespaceProto `protobuf:"varint,1,req,name=namespace,enum=hadoop.hdfs.XAttrProto_XAttrNamespaceProto" json:"namespace,omitempty"`
	Name             *string                         `protobuf:"bytes,2,req,name=name" json:"name,omitempty"`
	Value            []byte                          `protobuf:"bytes,3,opt,name=value" json:"value,omitempty"`
	XXX_unrecognized []byte                          `json:"-"`
}

func (m *XAttrProto) Reset()                    { *m = XAttrProto{} }
func (m *XAttrProto) String() string            { return proto.CompactTextString(m) }
func (*XAttrProto) ProtoMessage()               {}
func (*XAttrProto) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{0} }

func (m *XAttrProto) GetNamespace() XAttrProto_XAttrNamespaceProto {
	if m != nil && m.Namespace != nil {
		return *m.Namespace
	}
	return XAttrProto_USER
}

func (m *XAttrProto) GetName() string {
	if m != nil && m.Name != nil {
		return *m.Name
	}
	return ""
}

func (m *XAttrProto) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

type SetXAttrRequestProto struct {
	Src              *string     `protobuf:"bytes,1,req,name=src" json:"src,omitempty"`
	XAttr            *XAttrProto `protobuf:"bytes,2,opt,name=xAttr" json:"xAttr,omitempty"`
	Flag             *uint32     `protobuf:"varint,3,opt,name=flag" json:"flag,omitempty"`
	XXX_unrecognized []byte      `json:"-"`
}

func (m *SetXAttrRequestProto) Reset()                    { *m = SetXAttrRequestProto{} }
func (m *SetXAttrRequestProto) String() string            { return proto.CompactTextString(m) }
func (*SetXAttrRequestProto) ProtoMessage()               {}
func (*SetXAttrRequestProto) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{1} }

func (m *SetXAttrRequestProto) GetSrc() string {
	if m != nil && m.Src != nil {
		return *m.Src
	}
	return ""
}

func (m *SetXAttrRequestProto) GetXAttr() *XAttrProto {
	if m != nil {
		return m.XAttr
	}
	return nil
}

func (m *SetXAttrRequestProto) GetFlag() uint32 {
	if m != nil && m.Flag != nil {
		return *m.Flag
	}
	return 0
}

type SetXAttrResponseProto struct {
	XXX_unrecognized []byte `json:"-"`
}

func (m *SetXAttrResponseProto) Reset()                    { *m = SetXAttrResponseProto{} }
func (m *SetXAttrResponseProto) String() string            { return proto.CompactTextString(m) }
func (*SetXAttrResponseProto) ProtoMessage()               {}
func (*SetXAttrResponseProto) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{2} }

type GetXAttrsRequestProto struct {
	Src              *string       `protobuf:"bytes,1,req,name=src" json:"src,omitempty"`
	XAttrs           []*XAttrProto `protobuf:"bytes,2,rep,name=xAttrs" json:"xAttrs,omitempty"`
	XXX_unrecognized []byte        `json:"-"`
}

func (m *GetXAttrsRequestProto) Reset()                    { *m = GetXAttrsRequestProto{} }
func (m *GetXAttrsRequestProto) String() string            { return proto.CompactTextString(m) }
func (*GetXAttrsRequestProto) ProtoMessage()               {}
func (*GetXAttrsRequestProto) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{3} }

func (m *GetXAttrsRequestProto) GetSrc() string {
	if m != nil && m.Src != nil {
		return *m.Src
	}
	return ""
}

func (m *GetXAttrsRequestProto) GetXAttrs() []*XAttrProto {
	if m != nil {
		return m.XAttrs
	}
	return nil
}

type GetXAttrsResponseProto struct {
	XAttrs           []*XAttrProto `protobuf:"bytes,1,rep,name=xAttrs" json:"xAttrs,omitempty"`
	XXX_unrecognized []byte        `json:"-"`
}

func (m *GetXAttrsResponseProto) Reset()                    { *m = GetXAttrsResponseProto{} }
func (m *GetXAttrsResponseProto) String() string            { return proto.CompactTextString(m) }
func (*GetXAttrsResponseProto) ProtoMessage()               {}
func (*GetXAttrsResponseProto) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{4} }

func (m *GetXAttrsResponseProto) GetXAttrs() []*XAttrProto {
	if m != nil {
		return m.XAttrs
	}
	return nil
}

type ListXAttrsRequestProto struct {
	Src              *string `protobuf:"bytes,1,req,name=src" json:"src,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *ListXAttrsRequestProto) Reset()                    { *m = ListXAttrsRequestProto{} }
func (m *ListXAttrsRequestProto) String() string            { return proto.CompactTextString(m) }
func (*ListXAttrsRequestProto) ProtoMessage()               {}
func (*ListXAttrsRequestProto) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{5} }

func (m *ListXAttrsRequestProto) GetSrc() string {
	if m != nil && m.Src != nil {
		return *m.Src
	}
	return ""
}

type ListXAttrsResponseProto struct {
	XAttrs           []*XAttrProto `protobuf:"bytes,1,rep,name=xAttrs" json:"xAttrs,omitempty"`
	XXX_unrecognized []byte        `json:"-"`
}

func (m *ListXAttrsResponseProto) Reset()                    { *m = ListXAttrsResponseProto{} }
func (m *ListXAttrsResponseProto) String() string            { return proto.CompactTextString(m) }
func (*ListXAttrsResponseProto) ProtoMessage()               {}
func (*ListXAttrsResponseProto) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{6} }

func (m *ListXAttrsResponseProto) GetXAttrs() []*XAttrProto {
	if m != nil {
		return m.XAttrs
	}
	return nil
}

type RemoveXAttrRequestProto struct {
	Src              *string     `protobuf:"bytes,1,req,name=src" json:"src,omitempty"`
	XAttr            *XAttrProto `protobuf:"bytes,2,opt,name=xAttr" json:"xAttr,omitempty"`
	XXX_unrecognized []byte      `json:"-"`
}

func (m *RemoveXAttrRequestProto) Reset()                    { *m = RemoveXAttrRequestProto{} }
func (m *RemoveXAttrRequestProto) String() string            { return proto.CompactTextString(m) }
func (*RemoveXAttrRequestProto) ProtoMessage()               {}
func (*RemoveXAttrRequestProto) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{7} }

func (m *RemoveXAttrRequestProto) GetSrc() string {
	if m != nil && m.Src != nil {
		return *m.Src
	}
	return ""
}

func (m *RemoveXAttrRequestProto) GetXAttr() *XAttrProto {
	if m != nil {
		return m.XAttr
	}
	return nil
}

type RemoveXAttrResponseProto struct {
	XXX_unrecognized []byte `json:"-"`
}

func (m *RemoveXAttrResponseProto) Reset()                    { *m = RemoveXAttrResponseProto{} }
func (m *RemoveXAttrResponseProto) String() string            { return proto.CompactTextString(m) }
func (*RemoveXAttrResponseProto) ProtoMessage()               {}
func (*RemoveXAttrResponseProto) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{8} }

func init() {
	proto.RegisterType((*XAttrProto)(nil), "hadoop.hdfs.XAttrProto")
	proto.RegisterType((*SetXAttrRequestProto)(nil), "hadoop.hdfs.SetXAttrRequestProto")
	proto.RegisterType((*SetXAttrResponseProto)(nil), "hadoop.hdfs.SetXAttrResponseProto")
	proto.RegisterType((*GetXAttrsRequestProto)(nil), "hadoop.hdfs.GetXAttrsRequestProto")
	proto.RegisterType((*GetXAttrsResponseProto)(nil), "hadoop.hdfs.GetXAttrsResponseProto")
	proto.RegisterType((*ListXAttrsRequestProto)(nil), "hadoop.hdfs.ListXAttrsRequestProto")
	proto.RegisterType((*ListXAttrsResponseProto)(nil), "hadoop.hdfs.ListXAttrsResponseProto")
	proto.RegisterType((*RemoveXAttrRequestProto)(nil), "hadoop.hdfs.RemoveXAttrRequestProto")
	proto.RegisterType((*RemoveXAttrResponseProto)(nil), "hadoop.hdfs.RemoveXAttrResponseProto")
	proto.RegisterEnum("hadoop.hdfs.XAttrSetFlagProto", XAttrSetFlagProto_name, XAttrSetFlagProto_value)
	proto.RegisterEnum("hadoop.hdfs.XAttrProto_XAttrNamespaceProto", XAttrProto_XAttrNamespaceProto_name, XAttrProto_XAttrNamespaceProto_value)
}

func init() { proto.RegisterFile("xattr.proto", fileDescriptor1) }

var fileDescriptor1 = []byte{
	// 408 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x93, 0xdf, 0x6e, 0xd3, 0x30,
	0x18, 0xc5, 0x71, 0xd2, 0xfd, 0xe9, 0x97, 0x0e, 0x79, 0x66, 0x5b, 0x22, 0xae, 0x22, 0x4b, 0x48,
	0xd1, 0x10, 0x41, 0xda, 0x0d, 0xdc, 0x66, 0xc5, 0xa0, 0xa2, 0x01, 0x93, 0x93, 0x8a, 0x6d, 0x37,
	0xc8, 0xca, 0xbc, 0x16, 0x91, 0xd5, 0x21, 0x76, 0xab, 0x3e, 0x0e, 0xcf, 0xc4, 0x13, 0xa1, 0x38,
	0xad, 0x92, 0x4a, 0xfc, 0xa9, 0xc4, 0xee, 0x8e, 0x3f, 0x9f, 0x9c, 0xf3, 0xfb, 0x14, 0x19, 0xbc,
	0xa5, 0x30, 0xa6, 0x8a, 0xcb, 0x4a, 0x19, 0x45, 0xbc, 0xa9, 0xb8, 0x55, 0xaa, 0x8c, 0xa7, 0xb7,
	0x77, 0x9a, 0xfe, 0x44, 0x00, 0x57, 0x89, 0x31, 0xd5, 0xa5, 0xbd, 0x1b, 0x41, 0x7f, 0x26, 0xee,
	0xa5, 0x2e, 0x45, 0x2e, 0x03, 0x14, 0x3a, 0xd1, 0xe3, 0xb3, 0xe7, 0x71, 0xc7, 0x1f, 0xb7, 0xde,
	0x46, 0x7e, 0x5c, 0xbb, 0xed, 0x8c, 0xb7, 0x5f, 0x13, 0x02, 0xbd, 0xfa, 0x10, 0x38, 0xa1, 0x13,
	0xf5, 0xb9, 0xd5, 0xe4, 0x08, 0x76, 0x16, 0xa2, 0x98, 0xcb, 0xc0, 0x0d, 0x51, 0x34, 0xe0, 0xcd,
	0x81, 0x7e, 0x82, 0x27, 0xbf, 0xc9, 0x22, 0xfb, 0xd0, 0x1b, 0xa7, 0x8c, 0xe3, 0x47, 0xc4, 0x83,
	0xbd, 0x8c, 0x8f, 0xd3, 0x8c, 0xbd, 0xc1, 0x88, 0x0c, 0x60, 0x3f, 0x65, 0xc3, 0x31, 0x1f, 0x65,
	0xd7, 0xd8, 0x21, 0x00, 0xbb, 0xe9, 0x75, 0x9a, 0xb1, 0x0f, 0xd8, 0x25, 0x7b, 0xe0, 0xf2, 0xe4,
	0x33, 0xee, 0xd1, 0x6f, 0x70, 0x94, 0x4a, 0x63, 0x33, 0xb9, 0xfc, 0x3e, 0x97, 0xda, 0x34, 0x89,
	0x18, 0x5c, 0x5d, 0xe5, 0x76, 0xaf, 0x3e, 0xaf, 0x25, 0x79, 0x01, 0x3b, 0xcb, 0xda, 0x16, 0x38,
	0x21, 0x8a, 0xbc, 0x33, 0xff, 0x0f, 0xbb, 0xf2, 0xc6, 0x55, 0xef, 0x74, 0x57, 0x88, 0x89, 0xc5,
	0x3f, 0xe0, 0x56, 0x53, 0x1f, 0x8e, 0xdb, 0x32, 0x5d, 0xaa, 0x99, 0x6e, 0xf8, 0xe9, 0x0d, 0x1c,
	0xbf, 0x5b, 0x5d, 0xe8, 0x7f, 0x60, 0xbc, 0x84, 0x5d, 0x5b, 0xa0, 0x03, 0x27, 0x74, 0xff, 0xc6,
	0xb1, 0xb2, 0xd1, 0x11, 0x9c, 0x74, 0xb2, 0x3b, 0xad, 0x9d, 0x28, 0xb4, 0x5d, 0xd4, 0x29, 0x9c,
	0x5c, 0x7c, 0xd5, 0x5b, 0x71, 0xd2, 0xf7, 0xe0, 0x77, 0xbd, 0xff, 0xd5, 0x7b, 0x03, 0x3e, 0x97,
	0xf7, 0x6a, 0x21, 0x1f, 0xfe, 0x3f, 0xd1, 0xa7, 0x10, 0x6c, 0x64, 0x77, 0x40, 0x4f, 0x5f, 0xc3,
	0xa1, 0x9d, 0xa6, 0xd2, 0xbc, 0x2d, 0xc4, 0x64, 0xdd, 0x38, 0xb8, 0x4a, 0xb2, 0x8c, 0x7f, 0x19,
	0x72, 0x96, 0x64, 0x0c, 0x23, 0x72, 0x08, 0x07, 0xcd, 0x84, 0xb3, 0xcb, 0x8b, 0x64, 0xc8, 0xb0,
	0x73, 0xfe, 0x0a, 0x9e, 0xa9, 0x6a, 0x12, 0x8b, 0x52, 0xe4, 0x53, 0xb9, 0x41, 0x60, 0x1f, 0x56,
	0xae, 0x8a, 0x46, 0x9c, 0x7b, 0x2d, 0x91, 0xfe, 0x81, 0xd0, 0xaf, 0x00, 0x00, 0x00, 0xff, 0xff,
	0x56, 0x91, 0x13, 0x1d, 0x7f, 0x03, 0x00, 0x00,
}
