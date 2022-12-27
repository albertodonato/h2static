package server

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// load BasicAuth credentials from file
func loadCredentials(path string) (map[string]string, error) {
	credentials := make(map[string]string)
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		tokens := strings.Split(scanner.Text(), ":")
		if len(tokens) != 2 {
			absPath, err := filepath.Abs(path)
			if err != nil {
				return nil, err
			}
			log.Printf("Skipping invalid credentials at %s:%d", absPath, lineNum)
			continue
		}
		credentials[tokens[0]] = tokens[1]
	}
	return credentials, nil
}
