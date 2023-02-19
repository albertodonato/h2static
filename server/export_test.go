package server

import (
	"net/http"
)

// Export StaticServer.getServer.
func GetServer(s *StaticServer) (*http.Server, error) {
	return s.getServer()
}

// Export getHumanByteSize.
var GetHumanByteSize = getHumanByteSize

// Export NewDebugMux.
func NewDebugMux() *http.ServeMux {
	return newDebugMux()
}
