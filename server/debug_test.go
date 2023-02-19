package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/albertodonato/h2static/server"
	"github.com/albertodonato/h2static/testhelpers"
)

func TestServeDebug(t *testing.T) {
	suite.Run(t, new(ServeDebugTestSuite))
}

type ServeDebugTestSuite struct {
	testhelpers.TempDirTestSuite
}

func (s *ServeDebugTestSuite) TestDebugVarURL() {
	server := http.Server{
		Addr:    ":0",
		Handler: server.NewDebugMux(),
	}
	r := httptest.NewRequest("GET", "/debug/vars", nil)
	w := httptest.NewRecorder()
	server.Handler.ServeHTTP(w, r)
	response := w.Result()
	s.Equal(http.StatusOK, response.StatusCode)
	var payload map[string]any
	json.Unmarshal(w.Body.Bytes(), &payload)
	_, ok := payload["cmdline"]
	s.True(ok)
}

func (s *ServeDebugTestSuite) TestDebugPprofURLS() {
	server := http.Server{
		Addr:    ":0",
		Handler: server.NewDebugMux(),
	}
	r := httptest.NewRequest("GET", "/debug/pprof/", nil)
	w := httptest.NewRecorder()
	server.Handler.ServeHTTP(w, r)
	response := w.Result()
	s.Equal(http.StatusOK, response.StatusCode)
	s.Equal("text/html; charset=utf-8", response.Header.Get("Content-Type"))
}
