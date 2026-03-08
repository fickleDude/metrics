package handler

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateHandler(t *testing.T) {
	type request struct {
		url    string
		method string
	}
	tests := []struct {
		name    string
		code    int
		request request
	}{
		{
			name:    "StatusOk",
			code:    200,
			request: request{url: "http://localhost:8080/update/counter/test/9", method: "POST"},
		},
		{
			name:    "StatusMethodNotAllowed",
			code:    405,
			request: request{url: "http://localhost:8080/update/counter/test/9", method: "GET"},
		},
		{
			name:    "StatusNotFound",
			code:    404,
			request: request{url: "http://localhost:8080/update/counter/9", method: "POST"},
		},
		{
			name:    "StatusBadRequest invalid type",
			code:    400,
			request: request{url: "http://localhost:8080/update/invalid/test/9", method: "POST"},
		},
		{
			name:    "StatusBadRequest invalid format",
			code:    400,
			request: request{url: "http://localhost:8080/update/counter/test/invalid", method: "POST"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := MemStorageHandler{} //create handler
			//handle request
			request := httptest.NewRequest(test.request.method, test.request.url, nil)
			w := httptest.NewRecorder()
			handler.UpdateHandler(w, request)
			//examine response
			result := w.Result()
			defer result.Body.Close()
			assert.Equal(t, test.code, result.StatusCode)
		})
	}
}
