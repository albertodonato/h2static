# Tiny static web server with TLS and HTTP/2 support

[![Build status](https://img.shields.io/travis/albertodonato/staticserve.svg)](https://travis-ci.com/albertodonato/staticserve)
[![Go Report Card](https://goreportcard.com/badge/github.com/albertodonato/staticserve)](https://goreportcard.com/report/github.com/albertodonato/staticserve)
[![GoDoc](https://godoc.org/github.com/albertodonato/staticserve?status.svg)](https://godoc.org/github.com/albertodonato/staticserve)

A minimal HTTP server using the builtin Go `http` library. It supports TLS and HTTP/2.

It can be run simply as

```bash
go run staticserve.go
```

## Build

Run

```bash
go build staticserve.go
```

which produces a `staticserve` binary.


## HTTPS support

To run the server on HTTPS, a key/certificate pair is required. The service can be run with

```bash
staticserve -tls-cert cert.pem -tls-key key.pem
```

## Usage

Full usage options are as follows:

```
Usage of staticserve:
  -addr string
        address and port to listen on (default ":8080")
  -dir string
        directory to serve (default ".")
  -disable-h2
        disable HTTP/2 support
  -log
        log requests
  -tls-cert string
        certificate file for TLS connections
  -tls-key string
        key file for TLS connections
```
