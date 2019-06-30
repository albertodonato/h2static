package server

import (
	"net/http"
)

// Export StaticServer.getServer.
func GetServer(s StaticServer) *http.Server {
	return s.getServer()
}
