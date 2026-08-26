# syntax=docker/dockerfile:1

FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 go build -o /out/restaurant-service .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /out/restaurant-service /usr/local/bin/restaurant-service
ENTRYPOINT ["/usr/local/bin/restaurant-service"]
