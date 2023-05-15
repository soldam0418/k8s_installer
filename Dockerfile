FROM golang:1.20.3-alpine3.17 AS builder

RUN mkdir -p /build
WORKDIR /build

COPY go.mod .
COPY go.sum .
RUN go mod vendor
COPY . .
RUN CGO_ENABLED=0 env GOOS=linux GOARCH=amd64 go build -o kubenhn main.go

RUN mkdir -p /dist
RUN cp -r /build/config /dist/config
RUN cp /build/k8s_installer /dist/kubenhn

FROM golang:alpine3.17

RUN apk add --update --no-cache openssh sshpass
RUN mkdir -p /app
WORKDIR /app

COPY --chown=0:0 --from=builder /dist /app/

CMD ["kubenhn"]