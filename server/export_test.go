package server

import (
	"net/http"
)

// Export StaticServer.getServer.
func GetServer(s *StaticServer) (*http.Server, error) {
	return s.getServer()
}

var GetHumanByteSize = getHumanByteSize
