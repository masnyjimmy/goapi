package goapi_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/masnyjimmy/goapi"
	"github.com/stretchr/testify/assert"
)

type Calculation struct {
	Left  int `json:"left"`
	Right int `json:"right"`
}

type Result struct {
	Result int `json:"result"`
}

func Calculate(calc Calculation) (Result, goapi.APIError) {
	if calc.Left < 0 || calc.Right < 0 {
		return Result{}, goapi.NewAPIError(
			http.StatusBadRequest,
			"left and right must be positive",
			nil,
		)
	}

	return Result{
		Result: calc.Left + calc.Right,
	}, nil
}

func EndpointSuccess(t *testing.T) {
	router := httprouter.New()
	api := goapi.NewAPI(router, goapi.DefaultErrorHandler())
	appRouter := api.Router()
	handle := appRouter.Route(goapi.MethodPost, "/calculate", Calculate, goapi.RouteSpec{})

	calculation := Calculation{
		Left:  5,
		Right: 16,
	}

	bodyBytes, err := json.Marshal(calculation)

	assert.NoError(t, err, "Unable to parse Calculation?")

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("post", "/calculate", bytes.NewReader(bodyBytes))

	handle(recorder, req, httprouter.Params{})

	bodyBytes, err = io.ReadAll(recorder.Body)

	assert.NoError(t, err, "Unable to read body bytes?")

	var result Result

	err = json.Unmarshal(bodyBytes, &result)

	assert.NoError(t, err, "Unable to unmarshal results")

	assert.Exactly(t, 21, result.Result, "Invalid calculate result")
}

func EndpointFail(t *testing.T) {
	router := httprouter.New()
	api := goapi.NewAPI(router, goapi.DefaultErrorHandler())
	appRouter := api.Router()

	handle := appRouter.Route(goapi.MethodPost, "/calculate", Calculate, goapi.RouteSpec{})

	calculation := Calculation{
		Left:  -1,
		Right: 16,
	}

	bodyBytes, err := json.Marshal(calculation)

	assert.NoError(t, err, "Unable to parse Calculation?")

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("post", "/calculate", bytes.NewReader(bodyBytes))
	handle(recorder, req, httprouter.Params{})

	assert.Exactly(t, http.StatusBadRequest, recorder.Code, "Wrong status code")

	bodyBytes, err = io.ReadAll(recorder.Body)

	assert.NoError(t, err, "Unable to read body bytes")

	var result goapi.DefaultErrorType

	err = json.Unmarshal(bodyBytes, &result)

	assert.NoError(t, err, "Unable to unmarshall error")

	assert.Exactly(t, 21, result.Detail, "Invalid calculate result")
}
