package main

import (
	"flag"
	"log"
	"net/http"
	"os"
)

type flags struct {
	Addr    string
	Dir     string
	Log     bool
	TLSCert string
	TLSKey  string
}

type loggingHandler struct {
	Handler http.Handler
}

func (lh *loggingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s %s", r.Proto, r.Method, r.URL)
	lh.Handler.ServeHTTP(w, r)
}

func parseFlags() (f flags) {
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	flag.StringVar(&f.Addr, "addr", ":8080", "address and port to listen on")
	flag.StringVar(&f.Dir, "dir", pwd, "directory to serve")
	flag.BoolVar(&f.Log, "log", false, "log requests")
	flag.StringVar(&f.TLSCert, "tls-cert", "", "certificate file for TLS connections")
	flag.StringVar(&f.TLSKey, "tls-key", "", "key file for TLS connections")
	flag.Parse()
	return
}

func main() {
	flags := parseFlags()
	serveTLS := flags.TLSCert != "" && flags.TLSKey != ""

	if flags.Log {
		kind := "HTTP"
		if serveTLS {
			kind = "HTTPS"
		}
		log.Printf("Starting %s server on %s, serving path %s", kind, flags.Addr, flags.Dir)
	}

	handler := http.FileServer(http.Dir(flags.Dir))
	if flags.Log {
		handler = &loggingHandler{Handler: handler}
	}

	var err error
	if serveTLS {
		err = http.ListenAndServeTLS(flags.Addr, flags.TLSCert, flags.TLSKey, handler)
	} else {
		err = http.ListenAndServe(flags.Addr, handler)
	}
	log.Fatal(err)
}
