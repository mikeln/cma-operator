# Initial preparations
FROM golang:1.10.1-alpine
RUN apk update && apk add git

# Let's add the code in
ADD . /go/src/github.com/samsung-cnct/cma-operator

# Time to intall and rund ep
RUN go get -u github.com/golang/dep/cmd/dep
WORKDIR /go/src/github.com/samsung-cnct/cma-operator
RUN dep ensure

# Build package in container
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o ./cma-operator cmd/cma-operator/main.go

# Now that we have build a static binary, we can stuff it in a bare-bones alpine image
FROM alpine
COPY --from=0 /go/src/github.com/samsung-cnct/cma-operator/cma-operator /cma-operator

ENTRYPOINT ["/cma-operator"]
