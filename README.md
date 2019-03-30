# Tiny static web server with TLS and HTTP/2 support

![h2static logo](./logo.svg)

[![Build status](https://img.shields.io/travis/albertodonato/h2static.svg)](https://travis-ci.com/albertodonato/h2static)
[![Go Report Card](https://goreportcard.com/badge/github.com/albertodonato/h2static)](https://goreportcard.com/report/github.com/albertodonato/h2static)
[![GoDoc](https://godoc.org/github.com/albertodonato/h2static?status.svg)](https://godoc.org/github.com/albertodonato/h2static)
[![Snap Status](https://build.snapcraft.io/badge/albertodonato/h2static.svg)](https://build.snapcraft.io/user/albertodonato/h2static)


A minimal HTTP server using the builtin Go `http` library. It supports TLS and HTTP/2.

It can be run simply as

```bash
go run h2static.go
```

## Build

Run

```bash
go build h2static.go
```

which produces a `h2static` binary.


## HTTPS support

To run the server on HTTPS, a key/certificate pair is required. The service can be run with

```bash
h2static -tls-cert cert.pem -tls-key key.pem
```

## Usage

Full usage options are as follows:

```
Usage of h2static:
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

## Install from Snap

It's also possible to install the tool from the [Snap Store](https://snapcraft.io), on systems where Snaps are supported, via

```bash
sudo snap install h2static
```

The `h2static` binary should be available in path.

It's also possible to configure the service in the snap so that it run
automatically.  See `snap info h2static` for details about the available snap
settings.

[![Get it from the Snap Store](https://snapcraft.io/static/images/badges/en/snap-store-black.svg)](https://snapcraft.io/h2static)
