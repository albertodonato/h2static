package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/albertodonato/h2static/server"
)

const helpHeader = `
Tiny static web server with TLS and HTTP/2 support.

Usage of %s:
`

// NewStaticServerFromCmdline returns a a StaticServer parsing cmdline args.
func NewStaticServerFromCmdline(fs *flag.FlagSet, args []string) (*server.StaticServer, error) {
	s := &server.StaticServer{}
	fs.StringVar(&s.Addr, "addr", ":8080", "address and port to listen on")
	fs.StringVar(&s.Dir, "dir", ".", "directory to serve")
	fs.BoolVar(
		&s.DisableLookupWithSuffix, "disable-lookup-with-suffix", false,
		"disable matching files with .htm(l) suffix for paths without suffix")
	fs.BoolVar(&s.DisableH2, "disable-h2", false, "disable HTTP/2 support")
	fs.BoolVar(&s.Log, "log", false, "log requests")
	fs.StringVar(&s.TLSCert, "tls-cert", "", "certificate file for TLS connections")
	fs.StringVar(&s.TLSKey, "tls-key", "", "key file for TLS connections")
	fs.Usage = func() {
		output := fs.Output()
		fmt.Fprintf(output, helpHeader, fs.Name())
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	return s, nil
}

func main() {
	server, err := NewStaticServerFromCmdline(flag.CommandLine, os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
