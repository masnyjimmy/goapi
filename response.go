package goapi

import (
	"fmt"
	"net/http"
)

type _Response struct {
	w *http.ResponseWriter
	// req     *http.Request
	Headers http.Header
	Status  int
}

func (r *_Response) SetCookie(cookie http.Cookie) error {

	if r.w == nil {
		return fmt.Errorf("response must be a parameter to handle cookies")
	}

	http.SetCookie(*r.w, &cookie)
	return nil
}

func (r *_Response) DeleteCookie(key string) error {

	if r.w == nil {
		return fmt.Errorf("response must be a parameter to handle cookies")
	}

	http.SetCookie(*r.w, &http.Cookie{
		Name:     key,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	return nil
}

type Response = *_Response

func newResponse(w *http.ResponseWriter, err bool) Response {
	out := &_Response{
		w:       w,
		Headers: make(http.Header),
		Status:  200,
	}

	if err {
		out.Status = 500
	}

	return out
}
