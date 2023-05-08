FROM golang:1.20.3-alpine3.17 AS builder

RUN mkdir -p /build
WORKDIR /build

COPY go.mod .
COPY go.sum .
RUN go mod vendor
COPY . .
RUN CGO_ENABLED=0 env GOOS=linux GOARCH=amd64 go build -o k8s_installer main.go

RUN mkdir -p /dist
RUN cp -r /build/extra_script /dist/extra_script
RUN cp /build/k8s_setup.sh /dist/k8s_setup.sh
RUN cp /build/config.yaml /dist/config.yaml
RUN cp /build/k8s_installer /dist/k8s_installer

FROM golang:alpine3.17

RUN apk add --update --no-cache openssh sshpass
RUN mkdir -p /app
WORKDIR /app

COPY --chown=0:0 --from=builder /dist /app/

ENTRYPOINT ["/app/k8s_installer"]
CMD ["server"]