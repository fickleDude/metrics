package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fickleDude/metrics.git/internal/repository"
	"github.com/fickleDude/metrics.git/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestUpdateHandler(t *testing.T) {
	//create handler
	repository := repository.NewMemStorage(nil)
	service := service.NewMemStorageService(repository)
	handler := NewMemStorageHandler(service)
	type request struct {
		url    string
		method string
		params map[string]string
	}
	tests := []struct {
		name    string
		code    int
		request request
	}{
		{
			name: "StatusOk",
			code: 200,
			request: request{
				url:    "http://localhost:8080/update",
				method: "POST",
				params: map[string]string{"type": "counter", "name": "test", "value": "9"},
			},
		},
		{
			name: "StatusBadRequest invalid type",
			code: 400,
			request: request{
				url:    "http://localhost:8080/update",
				method: "POST",
				params: map[string]string{"type": "invalid", "name": "test", "value": "9"},
			},
		},
		{
			name: "StatusBadRequest invalid format",
			code: 400,
			request: request{
				url:    "http://localhost:8080/update",
				method: "POST",
				params: map[string]string{"type": "counter", "name": "test", "value": "invalid"},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			//create request
			request := httptest.NewRequest(test.request.method, test.request.url, nil)
			//add url params
			chiCtx := chi.NewRouteContext()
			for name, value := range test.request.params {
				chiCtx.URLParams.Add(name, value)
			}
			request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, chiCtx))
			//handle request
			w := httptest.NewRecorder()
			function := http.HandlerFunc(handler.UpdateMetricHandler)
			function.ServeHTTP(w, request)
			//examine response
			result := w.Result()
			defer result.Body.Close()
			assert.Equal(t, test.code, result.StatusCode)
		})
	}
}
