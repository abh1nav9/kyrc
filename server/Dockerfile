# Build the kyrc leaderboard server.
#
# IMPORTANT: build context must be the REPO ROOT, not server/. The server
# module has `replace github.com/abh1nav9/kyrc => ../`, so it needs the parent
# module present at build time. render.yaml / any host must set the Docker
# build context to the repo root and point at this file.
#
#   docker build -f server/Dockerfile -t kyrc-server .
FROM golang:1.24-alpine AS build
WORKDIR /src

# Copy the parent (CLI) module — the server depends on its internal packages.
COPY go.mod go.sum ./
COPY internal ./internal

# Copy the server module and build it.
COPY server ./server
WORKDIR /src/server
RUN go mod download
RUN CGO_ENABLED=0 go build -o /kyrc-server .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /kyrc-server /usr/local/bin/kyrc-server
# DATABASE_URL (and optional PORT, CORS_ORIGIN) come from the environment.
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/kyrc-server"]
