package goapi

import (
	"fmt"
	"net/http"
	"reflect"
	"slices"
	"unicode"

	"github.com/julienschmidt/httprouter"
)

type AppMeta struct {
	Title       string
	Version     string
	Description string
}

type schemeGroup struct {
	Name    string
	Schemes []string
}

type schemeGroups []schemeGroup

func (sg *schemeGroups) addScheme(name string, scheme string) error {
	for i := range *sg {
		group := &(*sg)[i]
		if group.Name == name {
			if slices.Contains(group.Schemes, scheme) {
				return fmt.Errorf("Duplicate scheme")
			}
			group.Schemes = append(group.Schemes, scheme)
			return nil
		}
	}

	*sg = append(*sg, schemeGroup{
		Name:    name,
		Schemes: []string{scheme},
	})

	return nil
}

type API struct {
	errorHandler genericErrorHandler
	errorIn      reflect.Type
	errorOut     reflect.Type
	errorScheme  string

	router       *httprouter.Router
	Meta         AppMeta
	Servers      Servers
	Tags         Tags
	Schemas      Schemas
	SchemeGroups schemeGroups
	Endpoints    Endpoints
}

func NewAPI[E error, T any](
	router *httprouter.Router,
	errorHandler ErrorHandler[E, T],
	meta AppMeta,
) API {
	api := API{
		errorHandler: func(r Response, req *http.Request, err any) any {
			return errorHandler(r, req, err.(E))
		},
		Meta:   meta,
		router: router,
	}

	schema, err := api.Schemas.RegisterSchema(GetType[T]())

	if err != nil {
		panic(err)
	}

	api.errorIn = GetType[E]()
	api.errorOut = GetType[T]()
	api.errorScheme = schema.Name

	return api
}

func (api *API) Router() Router {
	return Router{
		api:    api,
		prefix: "",
	}
}

func (api *API) Setup() error {
	return generate(api)
}

func schemePrefix(scheme string) string {
	runes := []rune(scheme)

	runes[0] = unicode.ToLower(runes[0])

	return string(runes)
}
