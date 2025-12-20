FROM golang:1.25.1 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /webserver ./cmd/webserver


FROM alpine:3.18
RUN apk add --no-cache ca-certificates

COPY --from=build /webserver /webserver

COPY --from=build /src/internal/storage/migrations /migrations


WORKDIR /
EXPOSE 8080
ENTRYPOINT ["/webserver"]
