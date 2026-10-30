# syntax=docker/dockerfile:1
FROM golang:1.26-alpine AS build
WORKDIR /src

# With go.sum committed (CI enforces this) the build uses exactly the locked versions
# and the module download layer is cached until go.mod/go.sum change.
COPY go.mod go.sum* ./
RUN if [ -f go.sum ]; then go mod download; fi

COPY . .
# Transitional fallback: without a committed go.sum, resolve versions at build time.
# Remove this line once go.sum is on main.
RUN if [ ! -f go.sum ]; then go mod tidy; fi \
 && CGO_ENABLED=0 go build -mod=readonly -trimpath -ldflags="-s -w" -o /out/nemosyne-api ./cmd/nemosyne-api

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/nemosyne-api /nemosyne-api
EXPOSE 8081
ENTRYPOINT ["/nemosyne-api"]
