package goapi

import "reflect"

type JsonType string

const (
	JsonNull    JsonType = "null"
	JsonBoolean JsonType = "boolean"
	JsonNumber  JsonType = "number"
	JsonInteger JsonType = "integer"
	JsonString  JsonType = "string"
	JsonArray   JsonType = "array"
	JsonObject  JsonType = "object"
)

type Meta struct {
	Type JsonType
	Rest map[string]string
}

func BuildTypeMeta(jsonType JsonType, t reflect.Type) Meta {
	meta := Meta{
		Type: jsonType,
		Rest: make(map[string]string),
	}

	if format := extractFormat(t); format != "" {
		meta.Rest["format"] = format
	}

	return meta
}

func BuildFieldMeta(jsonType JsonType, f reflect.StructField) Meta {
	meta := BuildTypeMeta(jsonType, f.Type)

	if tag, has := f.Tag.Lookup("format"); has {
		meta.Rest["format"] = tag
	}

	return meta
}
