package goapi

import (
	"fmt"
	"reflect"
	"strings"
)

type Property struct {
	Name string
	Meta Meta
}

type Schema struct {
	sourceType reflect.Type
	Name       string
	Properties []Property
}

type Schemas []Schema

func (s *Schemas) RegisterSchema(Type reflect.Type) (Schema, error) {

	for _, schema := range *s {
		if schema.sourceType == Type {
			return schema, nil
		}
	}

	schema := Schema{
		sourceType: Type,
		Name:       Type.Name(),
		Properties: []Property{},
	}

	for i := 0; i < Type.NumField(); i++ {
		field := Type.Field(i)

		if !field.IsExported() {
			continue
		}
		// handle json tag part
		if tag := field.Tag.Get("json"); tag == "-" {
			continue
		}

		name := field.Name
		if tag := field.Tag.Get("json"); tag != "" {
			parts := strings.Split(tag, ",")
			if parts[0] != "" {
				name = parts[0]
			}
		}
		// build meta (openapi schema) part
		fieldType := field.Type

		jt, err := resolveJsonType(fieldType)

		if err != nil {
			return Schema{}, err
		}

		meta := BuildFieldMeta(jt.jsonType, field)
		// use default if not defined
		if _, has := meta.Rest["format"]; !has && jt.format != "" {
			meta.Rest["format"] = jt.format
		}

		schema.Properties = append(schema.Properties, Property{
			Name: name,
			Meta: meta,
		})
		// //Add nested schemas (TODO: handle it first (add as ref))
		// if fieldType.Kind() == reflect.Struct {
		// 	s.RegisterSchema(fieldType)
		// }
	}

	*s = append(*s, schema)

	return schema, nil
}

type jsonTypeDescriptor struct {
	jsonType JsonType
	format   string
}

func resolveJsonType(Type reflect.Type) (jsonTypeDescriptor, error) {

	Type = derefType(Type)

	switch Type.Kind() {
	case reflect.Bool:
		return jsonTypeDescriptor{jsonType: JsonBoolean}, nil
	case reflect.Int:
		return jsonTypeDescriptor{jsonType: JsonInteger}, nil
	case reflect.Int8:
		return jsonTypeDescriptor{jsonType: JsonInteger, format: "int8"}, nil
	case reflect.Int16:
		return jsonTypeDescriptor{jsonType: JsonInteger, format: "int16"}, nil
	case reflect.Int32:
		return jsonTypeDescriptor{jsonType: JsonInteger, format: "int32"}, nil
	case reflect.Int64:
		return jsonTypeDescriptor{jsonType: JsonInteger, format: "int64"}, nil
	case reflect.Uint:
		return jsonTypeDescriptor{jsonType: JsonInteger, format: "uint"}, nil
	case reflect.Uint8:
		return jsonTypeDescriptor{jsonType: JsonInteger, format: "uint8"}, nil
	case reflect.Uint16:
		return jsonTypeDescriptor{jsonType: JsonInteger, format: "uint16"}, nil
	case reflect.Uint32:
		return jsonTypeDescriptor{jsonType: JsonInteger, format: "uint32"}, nil
	case reflect.Uint64:
		return jsonTypeDescriptor{jsonType: JsonInteger, format: "uint64"}, nil
	case reflect.Float32:
		return jsonTypeDescriptor{jsonType: JsonNumber, format: "float"}, nil
	case reflect.Float64:
		return jsonTypeDescriptor{jsonType: JsonNumber, format: "double"}, nil
	case reflect.String:
		return jsonTypeDescriptor{jsonType: JsonString}, nil
	default:
		return jsonTypeDescriptor{}, fmt.Errorf("invalid type (%s)", reflect.TypeOf(Type).Name())
	}
}
