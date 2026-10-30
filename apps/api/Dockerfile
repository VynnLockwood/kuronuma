# syntax=docker/dockerfile:1
FROM golang:1.24-alpine AS build
WORKDIR /src
COPY . .
# `go mod tidy` fills in go.sum on first build. Once go.sum is committed, replace
# it with `go mod download` for reproducible, cacheable builds.
RUN go mod tidy \
 && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/nemosyne-api ./cmd/nemosyne-api

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/nemosyne-api /nemosyne-api
EXPOSE 8081
ENTRYPOINT ["/nemosyne-api"]
