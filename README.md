# Tiny static web server with TLS and HTTP/2 support

![h2static logo](./logo.svg)

[![Build status](https://img.shields.io/travis/albertodonato/h2static.svg)](https://travis-ci.com/albertodonato/h2static)
[![Go Report Card](https://goreportcard.com/badge/github.com/albertodonato/h2static)](https://goreportcard.com/report/github.com/albertodonato/h2static)
[![Snap Status](https://build.snapcraft.io/badge/albertodonato/h2static.svg)](https://build.snapcraft.io/user/albertodonato/h2static)


A minimal HTTP server for serving static files, using the builtin Go `http`
library.

It provides a few handy features for serving files and static websites:

* support for HTTP/2
* support for TLS (HTTPS)
* support for HTTP Basic Authentication
* directory listing in HTML and JSON format
* serve `index.html`/`index.htm` files for the contaning directory
* serve the corresponding `.html`/`.htm` file for a path without the suffix
  (when such path doesn't exist)

It can be run simply as

```bash
go run ./cmd/h2static
```

and built with

```bash
go build ./cmd/h2static
```

which produces a `h2static` binary.


## HTTPS support

To run the server on HTTPS, a key/certificate pair in PEM format is
required. The service can be run with

```bash
h2static -tls-cert cert.pem -tls-key key.pem
```

## JSON directory listing

When requesting a path that matches a directory, it's possible to get the
listing in JSON format by setting the `Accept` header to `application/json` in
the request:

```
$ curl -s -H "Accept: application/json" http://localhost:8080/ | jq
{
  "Name": "/",
  "IsRoot": true,
  "Entries": [
    {
      "Name": "bar.txt",
      "IsDir": false,
      "Size": 11
    },
    {
      "Name": "foo.txt",
      "IsDir": false,
      "Size": 6
    },
    {
      "Name": "subdir/",
      "IsDir": true,
      "Size": 0
    }
  ]
}
```


## Usage

Full usage options are as follows:

```
Usage of h2static:

  -addr string
        address and port to listen on (default ":8080")
  -basic-auth string
        password file for Basic Auth (each line should be in the form "user:SHA512-hash")
  -dir string
        directory to serve (default ".")
  -disable-h2
        disable HTTP/2 support
  -disable-lookup-with-suffix
        disable matching files with .htm(l) suffix for paths without suffix
  -log
        log requests
  -show-dotfiles
        show files whose name starts with a dot
  -tls-cert string
        certificate file for TLS connections
  -tls-key string
        key file for TLS connections
```


## Install from Snap

The tool can be installed from the [Snap Store](https://snapcraft.io), on
systems where Snaps are supported, via

```bash
sudo snap install h2static
```

The `h2static` binary should be available in path.

It's also possible to configure the service in the snap so that it run
automatically.  See `snap info h2static` for details about the available snap
settings.

[![Get it from the Snap Store](https://snapcraft.io/static/images/badges/en/snap-store-black.svg)](https://snapcraft.io/h2static)
