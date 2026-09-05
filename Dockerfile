# Build
FROM golang:1.27-alpine AS build

RUN apk add --no-cache git

RUN go install github.com/a-h/templ/cmd/templ@v0.3.1020

ENV CGO_ENABLED=0

WORKDIR /go/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN templ generate && go build -ldflags="-s -w" -o bin/ezauth-example ./cmd/server

# Deploy
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /opt/ezauth-example
COPY --from=build /go/src/app/bin/ezauth-example ./ezauth-example
COPY --from=build /go/src/app/static ./static

EXPOSE 8080

CMD ["./ezauth-example"]
