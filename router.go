package goapi

import (
	"strings"

	"github.com/julienschmidt/httprouter"
)

type Router struct {
	api    *API
	prefix string
}

type Method string

const (
	MethodGet     Method = "get"
	MethodPost    Method = "post"
	MethodPut     Method = "put"
	MethodOptions Method = "options"
)

type RouterHandler = func(*Router)

type RouteSpec struct {
	Tags        []string
	Summary     string
	Description string
	OperationId string
}

func (r *Router) AddRoute(prefix string, handler RouterHandler) {
	router := Router{
		api:    r.api,
		prefix: joinPrefix(r.prefix, prefix),
	}

	handler(&router)
}

func (r *Router) Route(method Method, prefix string, fn Endpoint, spec RouteSpec) httprouter.Handle {
	fullPath := joinPrefix(r.prefix, prefix)

	if len(spec.Tags) == 0 {
		if tag := defaultTag(r.prefix); tag != "" {
			spec.Tags = append(spec.Tags, tag)
		}
	}

	ep := NewEndpointMethod(r.api, method, fullPath, fn, spec)

	return ep.Handler
}

func (r *Router) Get(prefix string, fn Endpoint, spec RouteSpec) httprouter.Handle {
	return r.Route(MethodGet, prefix, fn, spec)
}

func (r *Router) Post(prefix string, fn Endpoint, spec RouteSpec) httprouter.Handle {
	return r.Route(MethodPost, prefix, fn, spec)

}

func (r *Router) Options(prefix string, fn Endpoint, spec RouteSpec) httprouter.Handle {
	return r.Route(MethodOptions, prefix, fn, spec)
}

func joinPrefix(base, segment string) string {
	base = strings.TrimSuffix(base, "/")

	if segment == "" {
		return base
	}

	if !strings.HasPrefix(segment, "/") {
		segment = "/" + segment
	}

	return base + segment
}

func defaultTag(prefix string) string {
	trimmed := strings.Trim(prefix, "/")
	if trimmed == "" {
		return ""
	}

	parts := strings.Split(trimmed, "/")
	return parts[len(parts)-1]
}
