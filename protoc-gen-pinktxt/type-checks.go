package main

import (
	"text/template"

	desc "github.com/nilium/pinktxt/internal/plugin/google/protobuf"
)

func mergeTypeChecks(funcs template.FuncMap) template.FuncMap {
	if funcs == nil {
		funcs = make(template.FuncMap)
	}

	funcs["is_repeated"] = isRepeated
	funcs["is_optional"] = isOptional
	funcs["is_required"] = isRequired
	funcs["is_double"] = isDouble
	funcs["is_float"] = isFloat
	funcs["is_int64"] = isInt64
	funcs["is_uint64"] = isUint64
	funcs["is_int32"] = isInt32
	funcs["is_fixed64"] = isFixed64
	funcs["is_fixed32"] = isFixed32
	funcs["is_bool"] = isBool
	funcs["is_string"] = isString
	funcs["is_group"] = isGroup
	funcs["is_message"] = isMessage
	funcs["is_bytes"] = isBytes
	funcs["is_uint32"] = isUint32
	funcs["is_enum"] = isEnum
	funcs["is_sfixed32"] = isSfixed32
	funcs["is_sfixed64"] = isSfixed64
	funcs["is_sint32"] = isSint32
	funcs["is_sint64"] = isSint64
	return funcs
}

func isRepeated(d interface{}) bool {
	if p, ok := d.(*desc.FieldDescriptorProto); ok && p != nil && p.Type != nil {
		d = p.GetLabel()
	}

	if i, ok := d.(desc.FieldDescriptorProto_Label); ok {
		return i == desc.FieldDescriptorProto_LABEL_REPEATED
	}

	return false
}

func isOptional(d interface{}) bool {
	if p, ok := d.(*desc.FieldDescriptorProto); ok && p != nil && p.Type != nil {
		d = p.GetLabel()
	}

	if i, ok := d.(desc.FieldDescriptorProto_Label); ok {
		return i == desc.FieldDescriptorProto_LABEL_OPTIONAL
	}

	return false
}

func isRequired(d interface{}) bool {
	if p, ok := d.(*desc.FieldDescriptorProto); ok && p != nil && p.Type != nil {
		d = p.GetLabel()
	}

	if i, ok := d.(desc.FieldDescriptorProto_Label); ok {
		return i == desc.FieldDescriptorProto_LABEL_REQUIRED
	}

	return false
}

func isDouble(d interface{}) bool {
	if p, ok := d.(*desc.FieldDescriptorProto); ok && p != nil && p.Type != nil {
		d = p.GetType()
	}

	if i, ok := d.(desc.FieldDescriptorProto_Type); ok {
		return i == desc.FieldDescriptorProto_TYPE_DOUBLE
	}

	return false
}

func isFloat(d interface{}) bool {
	if p, ok := d.(*desc.FieldDescriptorProto); ok && p != nil && p.Type != nil {
		d = p.GetType()
	}

	if i, ok := d.(desc.FieldDescriptorProto_Type); ok {
		return i == desc.FieldDescriptorProto_TYPE_FLOAT
	}

	return false
}

func isInt64(d interface{}) bool {
	if p, ok := d.(*desc.FieldDescriptorProto); ok && p != nil && p.Type != nil {
		d = p.GetType()
	}

	if i, ok := d.(desc.FieldDescriptorProto_Type); ok {
		return i == desc.FieldDescriptorProto_TYPE_INT64
	}

	return false
}

func isUint64(d interface{}) bool {
	if p, ok := d.(*desc.FieldDescriptorProto); ok && p != nil && p.Type != nil {
		d = p.GetType()
	}

	if i, ok := d.(desc.FieldDescriptorProto_Type); ok {
		return i == desc.FieldDescriptorProto_TYPE_UINT64
	}

	return false
}

func isInt32(d interface{}) bool {
	if p, ok := d.(*desc.FieldDescriptorProto); ok && p != nil && p.Type != nil {
		d = p.GetType()
	}

	if i, ok := d.(desc.FieldDescriptorProto_Type); ok {
		return i == desc.FieldDescriptorProto_TYPE_INT32
	}

	return false
}

func isFixed64(d interface{}) bool {
	if p, ok := d.(*desc.FieldDescriptorProto); ok && p != nil && p.Type != nil {
		d = p.GetType()
	}

	if i, ok := d.(desc.FieldDescriptorProto_Type); ok {
		return i == desc.FieldDescriptorProto_TYPE_FIXED64
	}

	return false
}

func isFixed32(d interface{}) bool {
	if p, ok := d.(*desc.FieldDescriptorProto); ok && p != nil && p.Type != nil {
		d = p.GetType()
	}

	if i, ok := d.(desc.FieldDescriptorProto_Type); ok {
		return i == desc.FieldDescriptorProto_TYPE_FIXED32
	}

	return false
}

func isBool(d interface{}) bool {
	if p, ok := d.(*desc.FieldDescriptorProto); ok && p != nil && p.Type != nil {
		d = p.GetType()
	}

	if i, ok := d.(desc.FieldDescriptorProto_Type); ok {
		return i == desc.FieldDescriptorProto_TYPE_BOOL
	}

	return false
}

func isString(d interface{}) bool {
	if p, ok := d.(*desc.FieldDescriptorProto); ok && p != nil && p.Type != nil {
		d = p.GetType()
	}

	if i, ok := d.(desc.FieldDescriptorProto_Type); ok {
		return i == desc.FieldDescriptorProto_TYPE_STRING
	}

	return false
}

func isGroup(d interface{}) bool {
	if p, ok := d.(*desc.FieldDescriptorProto); ok && p != nil && p.Type != nil {
		d = p.GetType()
	}

	if i, ok := d.(desc.FieldDescriptorProto_Type); ok {
		return i == desc.FieldDescriptorProto_TYPE_GROUP
	}

	return false
}

func isMessage(d interface{}) bool {
	if p, ok := d.(*desc.FieldDescriptorProto); ok && p != nil && p.Type != nil {
		d = p.GetType()
	}

	if i, ok := d.(desc.FieldDescriptorProto_Type); ok {
		return i == desc.FieldDescriptorProto_TYPE_MESSAGE
	}
	_, ok := d.(*desc.DescriptorProto)
	return ok
}

func isBytes(d interface{}) bool {
	if p, ok := d.(*desc.FieldDescriptorProto); ok && p != nil && p.Type != nil {
		d = p.GetType()
	}

	if i, ok := d.(desc.FieldDescriptorProto_Type); ok {
		return i == desc.FieldDescriptorProto_TYPE_BYTES
	}

	return false
}

func isUint32(d interface{}) bool {
	if p, ok := d.(*desc.FieldDescriptorProto); ok && p != nil && p.Type != nil {
		d = p.GetType()
	}

	if i, ok := d.(desc.FieldDescriptorProto_Type); ok {
		return i == desc.FieldDescriptorProto_TYPE_UINT32
	}

	return false
}

func isEnum(d interface{}) bool {
	if p, ok := d.(*desc.FieldDescriptorProto); ok && p != nil && p.Type != nil {
		d = p.GetType()
	}

	if i, ok := d.(desc.FieldDescriptorProto_Type); ok {
		return i == desc.FieldDescriptorProto_TYPE_ENUM
	}

	_, ok := d.(*desc.EnumDescriptorProto)
	return ok
}

func isSfixed32(d interface{}) bool {
	if p, ok := d.(*desc.FieldDescriptorProto); ok && p != nil && p.Type != nil {
		d = p.GetType()
	}

	if i, ok := d.(desc.FieldDescriptorProto_Type); ok {
		return i == desc.FieldDescriptorProto_TYPE_SFIXED32
	}

	return false
}

func isSfixed64(d interface{}) bool {
	if p, ok := d.(*desc.FieldDescriptorProto); ok && p != nil && p.Type != nil {
		d = p.GetType()
	}

	if i, ok := d.(desc.FieldDescriptorProto_Type); ok {
		return i == desc.FieldDescriptorProto_TYPE_SFIXED64
	}

	return false
}

func isSint32(d interface{}) bool {
	if p, ok := d.(*desc.FieldDescriptorProto); ok && p != nil && p.Type != nil {
		d = p.GetType()
	}

	if i, ok := d.(desc.FieldDescriptorProto_Type); ok {
		return i == desc.FieldDescriptorProto_TYPE_SINT32
	}

	return false
}

func isSint64(d interface{}) bool {
	if p, ok := d.(*desc.FieldDescriptorProto); ok && p != nil && p.Type != nil {
		d = p.GetType()
	}

	if i, ok := d.(desc.FieldDescriptorProto_Type); ok {
		return i == desc.FieldDescriptorProto_TYPE_SINT64
	}

	return false
}
