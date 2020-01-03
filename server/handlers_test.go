package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/albertodonato/h2static/server"
)

func TestBasicAuthHanbler(t *testing.T) {
	suite.Run(t, new(BasicAuthHandlerTestSuite))
}

type BasicAuthHandlerTestSuite struct {
	suite.Suite

	handler server.BasicAuthHandler
}

func (s *BasicAuthHandlerTestSuite) SetupTest() {
	s.handler = server.BasicAuthHandler{
		Handler: http.NotFoundHandler(), // a valid request returns 404
		Credentials: map[string]string{
			// password is "bar"
			"foo": "d82c4eb5261cb9c8aa9855edd67d1bd10482f41529858d925094d173fa662aa91ff39bc5b188615273484021dfb16fd8284cf684ccf0fc795be3aa2fc1e6c181"},
		Realm: "tests",
	}
}

func (s *BasicAuthHandlerTestSuite) TestNoCredentials() {
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	s.handler.ServeHTTP(w, r)
	response := w.Result()
	s.Equal(http.StatusUnauthorized, response.StatusCode)
	s.Equal("401 Unauthorized", response.Status)
	s.Equal(
		`Basic realm="tests", charset="UTF-8"`,
		response.Header.Get("WWW-Authenticate"))
}

func (s *BasicAuthHandlerTestSuite) TestInvalidCredentials() {
	r := httptest.NewRequest("GET", "/", nil)
	r.SetBasicAuth("foo", "wrong")
	w := httptest.NewRecorder()
	s.handler.ServeHTTP(w, r)
	response := w.Result()
	s.Equal(http.StatusUnauthorized, response.StatusCode)
	s.Equal("401 Unauthorized", response.Status)
	s.Equal(
		`Basic realm="tests", charset="UTF-8"`,
		response.Header.Get("WWW-Authenticate"))
}

func (s *BasicAuthHandlerTestSuite) TestValidCredentials() {
	r := httptest.NewRequest("GET", "/", nil)
	r.SetBasicAuth("foo", "bar")
	w := httptest.NewRecorder()
	s.handler.ServeHTTP(w, r)
	response := w.Result()
	s.Equal(http.StatusNotFound, response.StatusCode)
}
