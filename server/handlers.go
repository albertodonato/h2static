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

// ServeHTTP logs server startup and serves via the configured handler.
func (h LoggingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s %s", r.Proto, r.Method, r.URL)
	h.Handler.ServeHTTP(w, r)
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
	http.Error(w, "401 unauthorized", http.StatusUnauthorized)
}

func (h BasicAuthHandler) hashPassword(password string) string {
	hash := sha512.New()
	hash.Write([]byte(password))
	return hex.EncodeToString(hash.Sum(nil))
}
