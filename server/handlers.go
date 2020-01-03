package server

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
)

// LoggingHandler wraps an http.Handler providing logging at startup.
type LoggingHandler struct {
	Handler http.Handler
}

type loggingResponseWriter struct {
	http.ResponseWriter

	statusCode int
	length     int
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

func (w *loggingResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *loggingResponseWriter) Write(b []byte) (n int, err error) {
	n, err = w.ResponseWriter.Write(b)
	w.length += n
	return n, err
}

// ServeHTTP logs server startup and serves via the configured handler.
func (h LoggingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	wr := newLoggingResponseWriter(w)
	h.Handler.ServeHTTP(wr, r)
	log.Printf(
		`%s %s %s %d %d %d "%s"`,
		r.Proto, r.Method, r.URL, r.ContentLength, wr.statusCode, wr.length,
		r.Header.Get("User-Agent"))
}

// BasicAuthHandler provides Basic Authorization.
type BasicAuthHandler struct {
	Handler http.Handler

	// User/password pairs
	Credentials map[string]string
	// The authentication realm
	Realm string
}

// ServeHTTP logs server startup and serves via the configured handler.
func (h BasicAuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user, pass, ok := r.BasicAuth()
	if user == "" {
		h.authRequiredResponse(w)
		return
	}
	hash, userFound := h.Credentials[user]
	if !ok || !userFound || h.hashPassword(pass) != hash {
		h.authRequiredResponse(w)
		return
	}
	h.Handler.ServeHTTP(w, r)
}

func (h BasicAuthHandler) authRequiredResponse(w http.ResponseWriter) {
	w.Header().Set(
		"WWW-Authenticate", fmt.Sprintf(`Basic realm="%s", charset="UTF-8"`, h.Realm))
	code := http.StatusUnauthorized
	http.Error(w, fmt.Sprintf("%d %s", code, http.StatusText(code)), code)
}

func (h BasicAuthHandler) hashPassword(password string) string {
	hash := sha512.New()
	hash.Write([]byte(password))
	return hex.EncodeToString(hash.Sum(nil))
}
