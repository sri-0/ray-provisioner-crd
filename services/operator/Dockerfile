FROM golang:1.22 AS builder
WORKDIR /workspace
COPY services/operator/go.mod services/operator/go.sum ./
RUN go mod download
COPY services/operator ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /operator ./cmd/operator
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /api-server ./cmd/api-server

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /operator /operator
COPY --from=builder /api-server /api-server
USER 65532:65532
ENTRYPOINT ["/operator"]
