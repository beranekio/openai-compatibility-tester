FROM golang:1.24-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/openai-compatibility-tester ./cmd/openai-compatibility-tester

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /out/openai-compatibility-tester /usr/local/bin/openai-compatibility-tester

ENTRYPOINT ["/usr/local/bin/openai-compatibility-tester"]