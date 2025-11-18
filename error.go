package goapi

import (
	"fmt"
	"net/http"
	"reflect"
)

type ErrorHandler[E error, T any] = func(r Response, req *http.Request, err E) T

type genericErrorHandler = func(r Response, req *http.Request, err any) any

type _APIError struct {
	StatusCode int
	Detail     string
	Headers    http.Header
}
type APIError = *_APIError

func NewAPIError(statusCode int, detail string, headers http.Header) APIError {
	return &_APIError{statusCode, detail, headers}
}

func (err *_APIError) Error() string {
	return fmt.Sprintf("[%v]: %v", err.StatusCode, err.Detail)
}

type DefaultErrorType struct {
	Detail string `json:"detail"`
}

func DefaultErrorHandler() ErrorHandler[APIError, DefaultErrorType] {
	return func(r Response, req *http.Request, err APIError) DefaultErrorType {

		for key, values := range err.Headers {
			for _, val := range values {
				r.Headers.Add(key, val)
			}
		}
		r.Status = err.StatusCode

		return DefaultErrorType{
			Detail: err.Detail,
		}
	}
}

func GetType[T any]() reflect.Type {
	return reflect.TypeOf((*T)(nil)).Elem()
}
