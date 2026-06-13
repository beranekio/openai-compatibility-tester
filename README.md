# openai-compatibility-tester

Docker container that tests whether an arbitrary HTTP endpoint is compatible with the [OpenAI API](https://platform.openai.com/docs/api-reference) by exercising it through the [official OpenAI Go SDK](https://github.com/openai/openai-go).

Each test suite calls a real SDK method (models list, chat completions, Responses API, embeddings, and more). If the endpoint returns payloads the SDK cannot parse, or responses that fail basic validation, the process exits with a non-zero status code.

## Quick start

```bash
docker run --rm \
  -e OPENAI_BASE_URL=https://your-endpoint.example/v1 \
  -e OPENAI_API_KEY=your-api-key \
  -e OPENAI_MODEL=your-chat-model \
  ghcr.io/beranekio/openai-compatibility-tester:latest
```

## Configuration

| Variable | Flag | Required | Default | Description |
|----------|------|----------|---------|-------------|
| `OPENAI_BASE_URL` | `--base-url` | yes | — | API base URL including `/v1` (e.g. `https://api.openai.com/v1`). Query parameters are not supported. |
| `OPENAI_API_KEY` | `--api-key` | yes | — | Bearer token sent to the endpoint |
| `OPENAI_MODEL` | `--model` | no | `gpt-4o-mini` | Model used for chat and responses suites |
| `OPENAI_COMPLETION_MODEL` | `--completion-model` | no | `gpt-3.5-turbo-instruct` when `completions` is selected, otherwise same as `OPENAI_MODEL` | Model used for the legacy completions suite |
| `OPENAI_EMBEDDING_MODEL` | `--embedding-model` | when `embeddings` is selected | — | Model used for the embeddings suite |
| `TEST_SUITES` | `--suites` | no | `all` | Comma-separated suite names, or `all` for the default set |
| `REQUEST_TIMEOUT` | `--timeout` | no | `2m` | Per-suite request timeout |

List available suites:

```bash
docker run --rm ghcr.io/beranekio/openai-compatibility-tester:latest --list-suites
```

### Test suites

| Suite | SDK surface | Endpoint |
|-------|-------------|----------|
| `models` | `client.Models.List` | `GET /v1/models` |
| `chat_completions` | `client.Chat.Completions.New` | `POST /v1/chat/completions` |
| `chat_completions_stream` | `client.Chat.Completions.NewStreaming` | `POST /v1/chat/completions` (stream) |
| `completions` | `client.Completions.New` | `POST /v1/completions` |
| `embeddings` | `client.Embeddings.New` | `POST /v1/embeddings` |
| `responses` | `client.Responses.New` | `POST /v1/responses` |
| `responses_stream` | `client.Responses.NewStreaming` | `POST /v1/responses` (stream) |

Default suites: `models`, `chat_completions`, `chat_completions_stream`, `responses`, `responses_stream`.

Add `embeddings` only when your endpoint exposes embedding models:

```bash
docker run --rm \
  -e OPENAI_BASE_URL=https://your-endpoint.example/v1 \
  -e OPENAI_API_KEY=your-api-key \
  -e OPENAI_MODEL=your-chat-model \
  -e OPENAI_EMBEDDING_MODEL=your-embedding-model \
  -e TEST_SUITES=models,chat_completions,chat_completions_stream,responses,responses_stream,embeddings \
  ghcr.io/beranekio/openai-compatibility-tester:latest
```

Run a subset:

```bash
docker run --rm \
  -e OPENAI_BASE_URL=https://your-endpoint.example/v1 \
  -e OPENAI_API_KEY=your-api-key \
  -e TEST_SUITES=models,chat_completions \
  ghcr.io/beranekio/openai-compatibility-tester:latest
```

## Exit codes

| Code | Meaning |
|------|---------|
| `0` | All selected suites passed |
| `1` | One or more suites failed compatibility checks |
| `2` | Configuration or runner error |

## Development

```bash
go test ./...
go build -o bin/openai-compatibility-tester ./cmd/openai-compatibility-tester

OPENAI_BASE_URL=http://127.0.0.1:4010/v1 \
OPENAI_API_KEY=test \
./bin/openai-compatibility-tester
```

Build the container locally:

```bash
docker build -t openai-compatibility-tester .
```

## CI and publishing

GitHub Actions runs unit tests and builds the Docker image on every push and pull request to `main`. When tests pass on a push to `main`, the image is published to GHCR:

`ghcr.io/beranekio/openai-compatibility-tester:latest`