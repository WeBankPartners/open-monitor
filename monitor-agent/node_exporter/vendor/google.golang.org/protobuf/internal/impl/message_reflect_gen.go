// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Code generated by generate-types. DO NOT EDIT.

package impl

import (
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoiface"
)

func (m *messageState) Descriptor() protoreflect.MessageDescriptor {
	return m.messageInfo().Desc
}
func (m *messageState) Type() protoreflect.MessageType {
	return m.messageInfo()
}
func (m *messageState) New() protoreflect.Message {
	return m.messageInfo().New()
}
func (m *messageState) Interface() protoreflect.ProtoMessage {
	return m.protoUnwrap().(protoreflect.ProtoMessage)
}
func (m *messageState) protoUnwrap() interface{} {
	return m.pointer().AsIfaceOf(m.messageInfo().GoReflectType.Elem())
}
func (m *messageState) ProtoMethods() *protoiface.Methods {
	m.messageInfo().init()
	return &m.messageInfo().methods
}

// ProtoMessageInfo is a pseudo-internal API for allowing the v1 code
// to be able to retrieve a v2 MessageInfo struct.
//
// WARNING: This method is exempt from the compatibility promise and
// may be removed in the future without warning.
func (m *messageState) ProtoMessageInfo() *MessageInfo {
	return m.messageInfo()
}

func (m *messageState) Range(f func(protoreflect.FieldDescriptor, protoreflect.Value) bool) {
	m.messageInfo().init()
	for _, ri := range m.messageInfo().rangeInfos {
		switch ri := ri.(type) {
		case *fieldInfo:
			if ri.has(m.pointer()) {
				if !f(ri.fieldDesc, ri.get(m.pointer())) {
					return
				}
			}
		case *oneofInfo:
			if n := ri.which(m.pointer()); n > 0 {
				fi := m.messageInfo().fields[n]
				if !f(fi.fieldDesc, fi.get(m.pointer())) {
					return
				}
			}
		}
	}
	m.messageInfo().extensionMap(m.pointer()).Range(f)
}
func (m *messageState) Has(fd protoreflect.FieldDescriptor) bool {
	m.messageInfo().init()
	if fi, xt := m.messageInfo().checkField(fd); fi != nil {
		return fi.has(m.pointer())
	} else {
		return m.messageInfo().extensionMap(m.pointer()).Has(xt)
	}
}
func (m *messageState) Clear(fd protoreflect.FieldDescriptor) {
	m.messageInfo().init()
	if fi, xt := m.messageInfo().checkField(fd); fi != nil {
		fi.clear(m.pointer())
	} else {
		m.messageInfo().extensionMap(m.pointer()).Clear(xt)
	}
}
func (m *messageState) Get(fd protoreflect.FieldDescriptor) protoreflect.Value {
	m.messageInfo().init()
	if fi, xt := m.messageInfo().checkField(fd); fi != nil {
		return fi.get(m.pointer())
	} else {
		return m.messageInfo().extensionMap(m.pointer()).Get(xt)
	}
}
func (m *messageState) Set(fd protoreflect.FieldDescriptor, v protoreflect.Value) {
	m.messageInfo().init()
	if fi, xt := m.messageInfo().checkField(fd); fi != nil {
		fi.set(m.pointer(), v)
	} else {
		m.messageInfo().extensionMap(m.pointer()).Set(xt, v)
	}
}
func (m *messageState) Mutable(fd protoreflect.FieldDescriptor) protoreflect.Value {
	m.messageInfo().init()
	if fi, xt := m.messageInfo().checkField(fd); fi != nil {
		return fi.mutable(m.pointer())
	} else {
		return m.messageInfo().extensionMap(m.pointer()).Mutable(xt)
	}
}
func (m *messageState) NewField(fd protoreflect.FieldDescriptor) protoreflect.Value {
	m.messageInfo().init()
	if fi, xt := m.messageInfo().checkField(fd); fi != nil {
		return fi.newField()
	} else {
		return xt.New()
	}
}
func (m *messageState) WhichOneof(od protoreflect.OneofDescriptor) protoreflect.FieldDescriptor {
	m.messageInfo().init()
	if oi := m.messageInfo().oneofs[od.Name()]; oi != nil && oi.oneofDesc == od {
		return od.Fields().ByNumber(oi.which(m.pointer()))
	}
	panic("invalid oneof descriptor")
}
func (m *messageState) GetUnknown() protoreflect.RawFields {
	m.messageInfo().init()
	return m.messageInfo().getUnknown(m.pointer())
}
func (m *messageState) SetUnknown(b protoreflect.RawFields) {
	m.messageInfo().init()
	m.messageInfo().setUnknown(m.pointer(), b)
}
func (m *messageState) IsValid() bool {
	return !m.pointer().IsNil()
}

func (m *messageReflectWrapper) Descriptor() protoreflect.MessageDescriptor {
	return m.messageInfo().Desc
}
func (m *messageReflectWrapper) Type() protoreflect.MessageType {
	return m.messageInfo()
}
func (m *messageReflectWrapper) New() protoreflect.Message {
	return m.messageInfo().New()
}
func (m *messageReflectWrapper) Interface() protoreflect.ProtoMessage {
	if m, ok := m.protoUnwrap().(protoreflect.ProtoMessage); ok {
		return m
	}
	return (*messageIfaceWrapper)(m)
}
func (m *messageReflectWrapper) protoUnwrap() interface{} {
	return m.pointer().AsIfaceOf(m.messageInfo().GoReflectType.Elem())
}
func (m *messageReflectWrapper) ProtoMethods() *protoiface.Methods {
	m.messageInfo().init()
	return &m.messageInfo().methods
}

// ProtoMessageInfo is a pseudo-internal API for allowing the v1 code
// to be able to retrieve a v2 MessageInfo struct.
//
// WARNING: This method is exempt from the compatibility promise and
// may be removed in the future without warning.
func (m *messageReflectWrapper) ProtoMessageInfo() *MessageInfo {
	return m.messageInfo()
}

func (m *messageReflectWrapper) Range(f func(protoreflect.FieldDescriptor, protoreflect.Value) bool) {
	m.messageInfo().init()
	for _, ri := range m.messageInfo().rangeInfos {
		switch ri := ri.(type) {
		case *fieldInfo:
			if ri.has(m.pointer()) {
				if !f(ri.fieldDesc, ri.get(m.pointer())) {
					return
				}
			}
		case *oneofInfo:
			if n := ri.which(m.pointer()); n > 0 {
				fi := m.messageInfo().fields[n]
				if !f(fi.fieldDesc, fi.get(m.pointer())) {
					return
				}
			}
		}
	}
	m.messageInfo().extensionMap(m.pointer()).Range(f)
}
func (m *messageReflectWrapper) Has(fd protoreflect.FieldDescriptor) bool {
	m.messageInfo().init()
	if fi, xt := m.messageInfo().checkField(fd); fi != nil {
		return fi.has(m.pointer())
	} else {
		return m.messageInfo().extensionMap(m.pointer()).Has(xt)
	}
}
func (m *messageReflectWrapper) Clear(fd protoreflect.FieldDescriptor) {
	m.messageInfo().init()
	if fi, xt := m.messageInfo().checkField(fd); fi != nil {
		fi.clear(m.pointer())
	} else {
		m.messageInfo().extensionMap(m.pointer()).Clear(xt)
	}
}
func (m *messageReflectWrapper) Get(fd protoreflect.FieldDescriptor) protoreflect.Value {
	m.messageInfo().init()
	if fi, xt := m.messageInfo().checkField(fd); fi != nil {
		return fi.get(m.pointer())
	} else {
		return m.messageInfo().extensionMap(m.pointer()).Get(xt)
	}
}
func (m *messageReflectWrapper) Set(fd protoreflect.FieldDescriptor, v protoreflect.Value) {
	m.messageInfo().init()
	if fi, xt := m.messageInfo().checkField(fd); fi != nil {
		fi.set(m.pointer(), v)
	} else {
		m.messageInfo().extensionMap(m.pointer()).Set(xt, v)
	}
}
func (m *messageReflectWrapper) Mutable(fd protoreflect.FieldDescriptor) protoreflect.Value {
	m.messageInfo().init()
	if fi, xt := m.messageInfo().checkField(fd); fi != nil {
		return fi.mutable(m.pointer())
	} else {
		return m.messageInfo().extensionMap(m.pointer()).Mutable(xt)
	}
}
func (m *messageReflectWrapper) NewField(fd protoreflect.FieldDescriptor) protoreflect.Value {
	m.messageInfo().init()
	if fi, xt := m.messageInfo().checkField(fd); fi != nil {
		return fi.newField()
	} else {
		return xt.New()
	}
}
func (m *messageReflectWrapper) WhichOneof(od protoreflect.OneofDescriptor) protoreflect.FieldDescriptor {
	m.messageInfo().init()
	if oi := m.messageInfo().oneofs[od.Name()]; oi != nil && oi.oneofDesc == od {
		return od.Fields().ByNumber(oi.which(m.pointer()))
	}
	panic("invalid oneof descriptor")
}
func (m *messageReflectWrapper) GetUnknown() protoreflect.RawFields {
	m.messageInfo().init()
	return m.messageInfo().getUnknown(m.pointer())
}
func (m *messageReflectWrapper) SetUnknown(b protoreflect.RawFields) {
	m.messageInfo().init()
	m.messageInfo().setUnknown(m.pointer(), b)
}
func (m *messageReflectWrapper) IsValid() bool {
	return !m.pointer().IsNil()
}
