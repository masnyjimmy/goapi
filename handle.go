package goapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

var ErrInvalidValueFormat = errors.New("invalid value format")

func parseValue(value string, t JsonType) (reflect.Value, error) {
	switch t {
	case JsonBoolean:
		switch value {
		case "true":
			return reflect.ValueOf(true), nil
		case "false":
			return reflect.ValueOf(false), nil
		default:
			return reflect.Value{}, fmt.Errorf("invalid boolean value: %s", value)
		}
	case JsonInteger:
		val, err := strconv.ParseInt(value, 10, 0)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("invalid integer value: %s", value)
		}
		return reflect.ValueOf(val), nil
	case JsonNumber:
		val, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(val), nil
	case JsonString:
		return reflect.ValueOf(value), nil
	}

	panic("invalid json type")
}

type HandleParam struct {
	In       ParamIn
	JsonType JsonType
	Required bool
	Name     string
	Special  bool
}

type HandleData struct {
	Endpoint Endpoint
	Params   []HandleParam
}

func DebugLol()
{

}

func makeRouterHandle(api *API, data HandleData) httprouter.Handle {

	errorHandler := api.errorHandler

	return func(
		w http.ResponseWriter,
		req *http.Request,
		params httprouter.Params,
	) {
		DebugLol()
		endpointType := reflect.TypeOf(data.Endpoint)
		out := make([]reflect.Value, len(data.Params))
		bodyParams := make([]int, 0)

		response := newResponse(&w, false)

		for index, el := range data.Params {
			paramType := endpointType.In(index)

			// handle special types first

			switch el.In {
			case ParamUndefined:
				switch paramType {
				case GetType[Response]():
					out[index] = reflect.ValueOf(response)
				}
			case ParamPath, ParamQuery: // parameter
				value := params.ByName(el.Name)

				if value == "" {
					value = req.URL.Query().Get(el.Name)
				}

				if value == "" {
					if el.Required {
						panic("Required but not provided")
						//TODO: handle http error, (invalid parameters)
					} else {
						out[index] = reflect.Zero(paramType)
					}
				} else {
					parsedValue, err := parseValue(value, el.JsonType)

					if err != nil {
						panic("Unable to parse value")
						//TODO: handle http error, invalid request or som
					}

					out[index] = parsedValue.Convert(paramType)
				}
			case ParamHeader: // header
				value := req.Header.Get(el.Name)

				if value == "" {
					if el.Required {
						panic("Unhandled: el.Required but not provided")
					} else {
						out[index] = reflect.Zero(paramType)
					}
				} else {
					parsedValue, err := parseValue(value, el.JsonType)

					if err != nil {
						panic("unable to parse value")
					}

					out[index] = parsedValue.Convert(paramType)
				}
			case ParamCookie:
				cookie, err := req.Cookie(el.Name)

				if err != nil {
					if el.Required {
						panic("Required but not provided")
					} else {
						out[index] = reflect.Zero(paramType)
					}
				} else {
					parsedValue, err := parseValue(cookie.Value, el.JsonType)

					if err != nil {
						panic("unable to parse value")
					}

					out[index] = parsedValue.Convert(paramType)
				}
			case ParamBody: // parse json
				bodyParams = append(bodyParams, index)
			}
		}

		if bc := len(bodyParams); bc == 1 {
			index := bodyParams[0]
			value := reflect.New(endpointType.In(index)).Interface()
			err := json.NewDecoder(req.Body).Decode(&value)

			if err != nil {
				panic("cannot decode")
				//TODO: handle
			}

			out[index] = reflect.ValueOf(value).Elem()
		} else if bc > 1 {
			schemes := make(map[string]any)

			for _, bodyIndex := range bodyParams {
				prefix := schemePrefix(data.Params[bodyIndex].Name)
				value := reflect.New(endpointType.In(bodyIndex)).Elem().Interface()
				schemes[prefix] = value
			}
			bodyBytes, err := io.ReadAll(req.Body)

			if err != nil {
				panic("Unable to read body")
			}

			var rawSchemes map[string]json.RawMessage
			if err := json.Unmarshal(bodyBytes, &rawSchemes); err != nil {
				panic("cannot unmarshal")
			}

			for _, bodyIndex := range bodyParams {
				prefix := schemePrefix(data.Params[bodyIndex].Name)

				rawJSON, ok := rawSchemes[prefix]
				if !ok {
					continue
				}

				value := reflect.New(endpointType.In(bodyIndex))
				if err := json.Unmarshal(rawJSON, value.Interface()); err != nil {
					panic("cannot unmarshal to target type")
				}

				out[bodyIndex] = value.Elem()
			}
		}
		ret := reflect.ValueOf(data.Endpoint).Call(out)

		// api error handling
		handleError := func(value reflect.Value) bool {
			if value.IsNil() {
				return false
			}

			errorResponse := newResponse(&w, true)
			result := errorHandler(errorResponse, req, value.Interface())

			bytes, err := json.Marshal(result)

			if err != nil {
				panic(err)
			}

			// apply headers from response
			for key, values := range errorResponse.Headers {
				for _, value := range values {
					w.Header().Add(key, value)
				}
			}

			w.WriteHeader(errorResponse.Status)

			if _, err := w.Write(bytes); err != nil {
				panic(err)
			}

			return true
		}

		// handle error, if no error and any value then send value
		errorHandled := false

		switch len(ret) {
		case 1:
			errorHandled = handleError(ret[0])
		case 2:
			errorHandled = handleError(ret[1])
		}

		if !errorHandled {
			// apply headers from response
			for key, values := range response.Headers {
				for _, value := range values {
					w.Header().Add(key, value)
				}
			}
			w.WriteHeader(response.Status)

			bytes, err := json.Marshal(ret[0].Interface())
			if err != nil {
				panic(err)
			}
			if _, err := w.Write(bytes); err != nil {
				panic(err)
			}
		}

	}
}
