FROM golang:1.27.2-bookworm AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /bellwright-mcp . && CGO_ENABLED=0 go build -trimpath -o /live-map ./cmd/live-map
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /bellwright-mcp /bellwright-mcp
COPY --from=builder /live-map /live-map
ENTRYPOINT ["/bellwright-mcp"]
