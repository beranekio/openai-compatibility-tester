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
| `OPENAI_MODEL` | `--model` | no | `gpt-4o-mini` | Model for chat completion suites and the model ID fetched by `models_get` |
| `OPENAI_RESPONSES_MODEL` | `--responses-model` | no | same as `OPENAI_MODEL` | Model used for Responses API suites |
| `OPENAI_COMPLETION_MODEL` | `--completion-model` | no | `gpt-3.5-turbo-instruct` when `completions` or `completions_stream` is selected, otherwise same as `OPENAI_MODEL` | Model used for legacy completions suites |
| `OPENAI_EMBEDDING_MODEL` | `--embedding-model` | when `embeddings` or `embeddings_batch` is selected | — | Model used for embedding suites |
| `OPENAI_VISION_MODEL` | `--vision-model` | when `chat_completions_vision` is selected | same as `OPENAI_MODEL` | Model used for vision chat suites |
| `OPENAI_IMAGE_MODEL` | `--image-model` | when image suites are selected | — | Model used for image generation suites |
| `OPENAI_TTS_MODEL` | `--tts-model` | when `audio_speech` is selected | — | Model used for text-to-speech suites |
| `OPENAI_WHISPER_MODEL` | `--whisper-model` | when audio transcription suites are selected | — | Model used for speech-to-text suites |
| `TEST_SUITES` | `--suites` | no | `all` | Comma-separated suite names, or preset: `all`/`default`, `extended`, `full` |
| `REQUEST_TIMEOUT` | `--timeout` | no | `2m` | Per-suite request timeout |
| `ALLOW_INSECURE_HTTP` | `--allow-insecure-http` | no | `false` | Allow plaintext `http://` to non-loopback hosts (loopback HTTP is always permitted) |

List available suites:

```bash
docker run --rm ghcr.io/beranekio/openai-compatibility-tester:latest --list-suites
```

### Test suites

| Suite | SDK surface | Endpoint |
|-------|-------------|----------|
| `models` | `client.Models.List` | `GET /v1/models` |
| `models_get` | `client.Models.Get` | `GET /v1/models/{id}` |
| `chat_completions` | `client.Chat.Completions.New` | `POST /v1/chat/completions` |
| `chat_completions_stream` | `client.Chat.Completions.NewStreaming` | `POST /v1/chat/completions` (stream) |
| `chat_completions_json` | `client.Chat.Completions.New` (`response_format` json_schema) | `POST /v1/chat/completions` |
| `chat_completions_vision` | `client.Chat.Completions.New` (with image input) | `POST /v1/chat/completions` |
| `chat_completions_tools` | `client.Chat.Completions.New` (with `tools`) | `POST /v1/chat/completions` |
| `chat_completions_tools_stream` | `client.Chat.Completions.NewStreaming` (with `tools`) | `POST /v1/chat/completions` (stream) |
| `completions` | `client.Completions.New` | `POST /v1/completions` |
| `completions_stream` | `client.Completions.NewStreaming` | `POST /v1/completions` (stream) |
| `embeddings` | `client.Embeddings.New` | `POST /v1/embeddings` |
| `embeddings_batch` | `client.Embeddings.New` (array input) | `POST /v1/embeddings` |
| `responses` | `client.Responses.New` | `POST /v1/responses` |
| `responses_stream` | `client.Responses.NewStreaming` | `POST /v1/responses` (stream) |
| `responses_tools` | `client.Responses.New` (with `tools`) | `POST /v1/responses` |
| `responses_tools_stream` | `client.Responses.NewStreaming` (with `tools`) | `POST /v1/responses` (stream) |
| `responses_json` | `client.Responses.New` (`text.format` json_schema) | `POST /v1/responses` |
| `responses_get` | `client.Responses.Get` | `GET /v1/responses/{id}` |
| `responses_delete` | `client.Responses.Delete` | `DELETE /v1/responses/{id}` |
| `responses_cancel` | `client.Responses.Cancel` | `POST /v1/responses/{id}/cancel` |

Default suites (`all` or `default`): `models`, `models_get`, `chat_completions`, `chat_completions_stream`, `responses`, `responses_stream`.

Extended preset (`extended`): default suites plus `chat_completions_json`, `chat_completions_tools`, `chat_completions_tools_stream`, `responses_tools`, `responses_tools_stream`, `responses_json`, `responses_get`, `responses_delete`, `responses_cancel`, `completions`, `completions_stream`, `embeddings`, `embeddings_batch`, and `chat_completions_vision`.

Full preset (`full`): every registered suite (see `--list-suites`).

Structured JSON output (`chat_completions_json`) is **opt-in** — included in `extended` and `full`, but not in the default `all` set:

```bash
docker run --rm \
  -e OPENAI_BASE_URL=https://your-endpoint.example/v1 \
  -e OPENAI_API_KEY=your-api-key \
  -e OPENAI_MODEL=your-chat-model \
  -e TEST_SUITES=chat_completions_json \
  ghcr.io/beranekio/openai-compatibility-tester:latest
```

Tool-calling suites (`chat_completions_tools`, `chat_completions_tools_stream`, `responses_tools`, `responses_tools_stream`) are **opt-in** — included in `extended` and `full`, but not in the default `all` set:

```bash
docker run --rm \
  -e OPENAI_BASE_URL=https://your-endpoint.example/v1 \
  -e OPENAI_API_KEY=your-api-key \
  -e OPENAI_MODEL=your-chat-model \
  -e OPENAI_EMBEDDING_MODEL=your-embedding-model \
  -e TEST_SUITES=extended \
  ghcr.io/beranekio/openai-compatibility-tester:latest
```

Or select tool suites explicitly:

```bash
docker run --rm \
  -e OPENAI_BASE_URL=https://your-endpoint.example/v1 \
  -e OPENAI_API_KEY=your-api-key \
  -e OPENAI_MODEL=your-chat-model \
  -e TEST_SUITES=chat_completions_tools,chat_completions_tools_stream,responses_tools,responses_tools_stream \
  ghcr.io/beranekio/openai-compatibility-tester:latest
```

Add `completions` and `completions_stream` only when your endpoint exposes legacy `/v1/completions`:

```bash
docker run --rm \
  -e OPENAI_BASE_URL=https://your-endpoint.example/v1 \
  -e OPENAI_API_KEY=your-api-key \
  -e OPENAI_COMPLETION_MODEL=your-instruct-model \
  -e TEST_SUITES=models,completions,completions_stream \
  ghcr.io/beranekio/openai-compatibility-tester:latest
```

Add `embeddings` and `embeddings_batch` only when your endpoint exposes embedding models:

```bash
docker run --rm \
  -e OPENAI_BASE_URL=https://your-endpoint.example/v1 \
  -e OPENAI_API_KEY=your-api-key \
  -e OPENAI_MODEL=your-chat-model \
  -e OPENAI_EMBEDDING_MODEL=your-embedding-model \
  -e TEST_SUITES=models,chat_completions,chat_completions_stream,responses,responses_stream,embeddings,embeddings_batch \
  ghcr.io/beranekio/openai-compatibility-tester:latest
```

**Vision** — requires a vision-capable model (`OPENAI_VISION_MODEL` defaults to `OPENAI_MODEL`):

```bash
docker run --rm \
  -e OPENAI_BASE_URL=https://your-endpoint.example/v1 \
  -e OPENAI_API_KEY=your-api-key \
  -e OPENAI_MODEL=your-vision-model \
  -e TEST_SUITES=chat_completions_vision \
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