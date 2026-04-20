# Builds the todo-api server binary.
#
# Build context expected: the gerpo repository root (the directory that
# contains the top-level go.mod). docker-compose.yml sets `context: ../..`
# so this runs automatically; if you invoke `docker build` by hand, call it
# from the gerpo repo root with `-f examples/todo-api/Dockerfile .`.
FROM golang:1.24-alpine AS build
WORKDIR /src

COPY . /src
WORKDIR /src/examples/todo-api

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/server ./cmd/server

FROM alpine:3.20
RUN adduser -D -u 10001 app
USER app
COPY --from=build /out/server /usr/local/bin/server
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/server"]
