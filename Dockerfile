FROM golang:latest AS build-image

ADD . /src
RUN cd /src && \
    go build -ldflags "-linkmode external -extldflags -static" -o /target/h2static ./cmd/h2static && \
    strip -s /target/h2static


FROM scratch

COPY --from=build-image /target /

EXPOSE 8080/tcp
VOLUME /www
ENTRYPOINT ["/h2static", "-log", "-dir", "/www"]
