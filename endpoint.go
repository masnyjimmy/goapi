package goapi

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"unicode"

	"github.com/julienschmidt/httprouter"
)

type Endpoint = any

type ParamIn string

const (
	ParamUndefined ParamIn = ""
	ParamPath      ParamIn = "path"
	ParamQuery     ParamIn = "query"
	ParamCookie    ParamIn = "cookie"
	ParamHeader    ParamIn = "header"
	ParamBody      ParamIn = "body"
)

type Parameter struct {
	sourceType  reflect.Type
	Name        string
	In          ParamIn
	Required    bool
	Description string
	Meta        Meta
}

type Parameters []Parameter

type EndpointMethod struct {
	Method       Method
	Tags         []string
	Summary      string
	Description  string
	OperationId  string
	sourceType   reflect.Type
	Parameters   Parameters
	RequestBody  string
	ResponseType string
	Handler      httprouter.Handle
}

type EndpointEntry struct {
	Path    string
	Methods []EndpointMethod
}

func (e *EndpointEntry) Set(value EndpointMethod) {
	for _, el := range e.Methods {
		if el.Method == value.Method {
			panic("duplicate method")
		}
	}

	e.Methods = append(e.Methods, value)
}

type Endpoints []EndpointEntry

func (e *Endpoints) Set(path string, value EndpointMethod) {
	for i := range *e {
		if (*e)[i].Path == path {
			(&(*e)[i]).Set(value)
			return
		}
	}

	*e = append(*e, EndpointEntry{
		Path:    path,
		Methods: []EndpointMethod{value},
	})
}

func (p *Parameters) RegisterParameter(Type reflect.Type, prefix string) (Parameter, error) {

	// read parameter spec
	name := Type.Name()
	description := ""
	var required bool

	if spec := extractSpec(Type); spec != nil {
		name = spec.Name
		required = spec.Required
		description = spec.Description
	}

	// check if query / path param

	in := ParamQuery

	if strings.Contains(prefix, fmt.Sprintf(":%s", name)) {
		// if parameter name in path then set path type
		in = ParamPath
	}

	// override if specified by user
	if inSpec := extractIn(Type); inSpec != ParamUndefined {
		in = inSpec
	}

	// build meta (openapi::schema) part
	jt, err := resolveJsonType(Type)

	if err != nil {
		return Parameter{}, err
	}

	meta := BuildTypeMeta(jt.jsonType, Type)

	if _, has := meta.Rest["format"]; has {
		meta.Rest["format"] = jt.format
	}

	parameter := Parameter{
		sourceType:  Type,
		Name:        name,
		In:          in,
		Required:    required,
		Description: description,
		Meta:        meta,
	}

	*p = append(*p, parameter)

	return parameter, nil
}

func getFunctionName(fn any) string {
	if reflect.TypeOf(fn).Kind() != reflect.Func {
		return ""
	}
	name := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()

	if idx := strings.LastIndex(name, "."); idx != -1 && idx+1 < len(name) {
		name = name[idx+1:]
	}

	r := []rune(name)
	r[0] = unicode.ToUpper(r[0])

	return string(r)
}

func newEndpointMethod(
	api *API,
	method Method,
	prefix string,
	endpoint Endpoint,
	spec RouteSpec,
) EndpointMethod {

	methodName := getFunctionName(endpoint)
	methodType := reflect.TypeOf(endpoint)

	handleData := HandleData{
		Endpoint: endpoint,
		Params:   make([]HandleParam, 0),
	}

	endpointMethod := EndpointMethod{
		Method:      method,
		Tags:        spec.Tags,
		Summary:     spec.Summary,
		Description: spec.Description,
		OperationId: spec.OperationId,
		sourceType:  methodType,
	}

	for p := range methodType.NumIn() {

		ParamType := methodType.In(p)
		handleParam := HandleParam{
			Special: false,
		}

		// process parameters, handle special types, schemas, and parameters
		switch ParamType {
		case GetType[Response]():
			handleParam.Special = true
		default:
			{
				if ParamType.Kind() == reflect.Struct {
					// its struct so its schema -> body -> required
					handleParam.In = ParamBody
					handleParam.Required = true
					handleParam.JsonType = JsonObject

					schema, err := api.Schemas.RegisterSchema(ParamType)

					if err != nil {
						panic(err)
					}

					handleParam.Name = schema.Name

					if endpointMethod.RequestBody == "" {
						endpointMethod.RequestBody = schema.Name
					} else {
						if endpointMethod.RequestBody != methodName {
							if err := api.SchemeGroups.addScheme(methodName, endpointMethod.RequestBody); err != nil {
								panic(err)
							}
							endpointMethod.RequestBody = methodName
						}
						if err := api.SchemeGroups.addScheme(methodName, schema.Name); err != nil {
							panic(err)
						}
					}
				} else {
					parameter, err := endpointMethod.Parameters.RegisterParameter(ParamType, prefix)
					if err != nil {
						panic(err)
					}
					handleParam.In = parameter.In
					handleParam.Required = parameter.Required
					handleParam.JsonType = parameter.Meta.Type
					handleParam.Name = parameter.Name
				}
			}
		}

		handleData.Params = append(handleData.Params, handleParam)

	}

	// handle return type

	switch methodType.NumOut() {
	case 1:
		if methodType.Out(0) != api.errorIn {
			panic(fmt.Errorf("invalid return type, must be (%[1]s) or ([T],%[1]s)", api.errorIn.Name()))
		}
	case 2:
		if methodType.Out(1) != api.errorIn {
			panic(fmt.Errorf("invalid return type, must be (%[1]s) or ([T],%[1]s)", api.errorIn.Name()))
		}
		if schema, err := api.Schemas.RegisterSchema(methodType.Out(0)); err != nil {
			panic(err)
		} else {
			endpointMethod.ResponseType = schema.Name
		}
	default:
		panic(fmt.Errorf("invalid return type, must be (%[1]s) or ([T],%[1]s)", api.errorIn.Name()))
	}

	// add tags to api
	for _, tag := range spec.Tags {
		api.Tags.Set(tag)
	}

	// build handle and set endpoint
	endpointMethod.Handler = makeRouterHandle(api, handleData)
	api.Endpoints.Set(prefix, endpointMethod)

	// add route
	api.router.Handle(strings.ToUpper(string(method)), prefix, endpointMethod.Handler)

	return endpointMethod
}
