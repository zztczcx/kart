# syntax=docker/dockerfile:1
FROM golang:1.23-alpine AS build
WORKDIR /app
RUN apk add --no-cache git build-base
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/server ./cmd/server

FROM gcr.io/distroless/base-debian12
WORKDIR /
COPY --from=build /out/server /server
COPY api/openapi.yaml /api/openapi.yaml
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/server"]


