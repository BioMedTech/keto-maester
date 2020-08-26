# Compile stage
FROM golang:latest AS build-env

# Build Delve
RUN go get github.com/go-delve/delve/cmd/dlv

ADD . /dockerdev
WORKDIR /dockerdev

RUN CGO_ENABLED=0 GO111MODULE=on GOOS=linux GOARCH=amd64 go build -gcflags="all=-N -l" -a -o /manager main.go

# Final stage
FROM debian:buster

EXPOSE 40000
WORKDIR /

COPY --from=build-env /go/bin/dlv /
COPY --from=build-env /manager /

CMD ["/dlv", "--listen=0.0.0.0:40000", "--headless=true", "--api-version=2", "--accept-multiclient", "exec", "/manager"]
