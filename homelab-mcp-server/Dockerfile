FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go-sdk/ ./go-sdk/
COPY homelab-mcp-server/ ./homelab-mcp-server/
WORKDIR /src/homelab-mcp-server
RUN go build -o /mcp-server .

FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=build /mcp-server /mcp-server
EXPOSE 3001
ENTRYPOINT ["/mcp-server"]
