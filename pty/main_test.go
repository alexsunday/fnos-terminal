package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMuxNoRedirect(t *testing.T) {
	mux := http.NewServeMux()
	cleanPrefix := "/app/terminal"

	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK:" + r.URL.Path))
	})

	mux.Handle(cleanPrefix, dummyHandler)
	mux.Handle(cleanPrefix+"/", dummyHandler)
	mux.Handle("/", dummyHandler)

	// Test 1: GET /app/terminal
	rec1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodGet, "/app/terminal", nil)
	mux.ServeHTTP(rec1, req1)
	require.Equal(t, http.StatusOK, rec1.Code)
	require.Equal(t, "OK:/app/terminal", rec1.Body.String())

	// Test 2: GET /app/terminal/
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/app/terminal/", nil)
	mux.ServeHTTP(rec2, req2)
	require.Equal(t, http.StatusOK, rec2.Code)
	require.Equal(t, "OK:/app/terminal/", rec2.Body.String())
}
