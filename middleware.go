package goapi

import "net/http"

type Middleware func(req *http.Request, next Middleware) Response
