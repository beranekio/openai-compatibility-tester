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
| `OPENAI_REASONING_MODEL` | `--reasoning-model` | when `chat_completions_reasoning` is selected | — | Model used for reasoning chat suites (e.g. `o3-mini`, `o4-mini`) |
| `OPENAI_IMAGE_MODEL` | `--image-model` | when `images_generations` or `images_edits` is selected | — | Model used for image generation and edit suites |
| `OPENAI_TTS_MODEL` | `--tts-model` | when `audio_speech` is selected | — | Model used for text-to-speech suites |
| `OPENAI_WHISPER_MODEL` | `--whisper-model` | when `audio_transcriptions` or `audio_translations` is selected | — | Model for non-streaming transcription and translation (e.g. `whisper-1`) |
| `OPENAI_TRANSCRIPTION_MODEL` | `--transcription-model` | when `audio_transcriptions_stream` is selected | — | Model for streaming transcription (e.g. `gpt-4o-mini-transcribe`) |
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
| `chat_completions_reasoning` | `client.Chat.Completions.New` (reasoning model) | `POST /v1/chat/completions` |
| `chat_completions_tools` | `client.Chat.Completions.New` (with `tools`) | `POST /v1/chat/completions` |
| `chat_completions_tools_stream` | `client.Chat.Completions.NewStreaming` (with `tools`) | `POST /v1/chat/completions` (stream) |
| `chat_completions_multi_turn` | `client.Chat.Completions.New` (multi-turn history with developer and tool messages) | `POST /v1/chat/completions` |
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
| `responses_input_items` | `client.Responses.InputItems.List` | `GET /v1/responses/{id}/input_items` |
| `responses_compact` | `client.Responses.Compact` | `POST /v1/responses/compact` |
| `responses_input_tokens` | `client.Responses.InputTokens.Count` | `POST /v1/responses/input_tokens` |
| `moderations` | `client.Moderations.New` | `POST /v1/moderations` |
| `images_generations` | `client.Images.Generate` | `POST /v1/images/generations` |
| `images_edits` | `client.Images.Edit` | `POST /v1/images/edits` |
| `images_variations` | `client.Images.NewVariation` | `POST /v1/images/variations` |
| `audio_speech` | `client.Audio.Speech.New` | `POST /v1/audio/speech` |
| `audio_transcriptions` | `client.Audio.Transcriptions.New` | `POST /v1/audio/transcriptions` |
| `audio_transcriptions_stream` | `client.Audio.Transcriptions.NewStreaming` | `POST /v1/audio/transcriptions` (stream) |
| `audio_translations` | `client.Audio.Translations.New` | `POST /v1/audio/translations` |

Default suites (`all` or `default`): `models`, `models_get`, `chat_completions`, `chat_completions_stream`, `responses`, `responses_stream`.

Extended preset (`extended`): default suites plus `chat_completions_json`, `chat_completions_tools`, `chat_completions_tools_stream`, `chat_completions_multi_turn`, `responses_tools`, `responses_tools_stream`, `responses_json`, `responses_get`, `responses_delete`, `responses_cancel`, `responses_input_items`, `responses_compact`, `responses_input_tokens`, `completions`, `completions_stream`, `embeddings`, `embeddings_batch`, `chat_completions_vision`, `moderations`, `images_generations`, `images_edits`, `audio_speech`, `audio_transcriptions`, `audio_transcriptions_stream`, and `audio_translations`.

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
  -e OPENAI_IMAGE_MODEL=your-image-model \
  -e OPENAI_TTS_MODEL=tts-1 \
  -e OPENAI_WHISPER_MODEL=whisper-1 \
  -e OPENAI_TRANSCRIPTION_MODEL=gpt-4o-mini-transcribe \
  -e OPENAI_REASONING_MODEL=o3-mini \
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

**Image generation** — requires an image model (`OPENAI_IMAGE_MODEL`):

```bash
docker run --rm \
  -e OPENAI_BASE_URL=https://your-endpoint.example/v1 \
  -e OPENAI_API_KEY=your-api-key \
  -e OPENAI_IMAGE_MODEL=your-image-model \
  -e TEST_SUITES=images_generations \
  ghcr.io/beranekio/openai-compatibility-tester:latest
```

`images_edits` uses `OPENAI_IMAGE_MODEL` (GPT Image models or `dall-e-2`; `dall-e-3` is not supported for edits).

Add `images_variations` only when your endpoint still exposes legacy DALL-E 2 `/v1/images/variations`. The suite always requests `dall-e-2` (the only model that endpoint supports) and does not use `OPENAI_IMAGE_MODEL`. Official OpenAI retired DALL-E models in May 2026; this suite is included in `full` but not in `extended`:

```bash
docker run --rm \
  -e OPENAI_BASE_URL=https://your-endpoint.example/v1 \
  -e OPENAI_API_KEY=your-api-key \
  -e TEST_SUITES=images_variations \
  ghcr.io/beranekio/openai-compatibility-tester:latest
```

**Text-to-speech** — requires a TTS model (`OPENAI_TTS_MODEL`):

```bash
docker run --rm \
  -e OPENAI_BASE_URL=https://your-endpoint.example/v1 \
  -e OPENAI_API_KEY=your-api-key \
  -e OPENAI_TTS_MODEL=tts-1 \
  -e TEST_SUITES=audio_speech \
  ghcr.io/beranekio/openai-compatibility-tester:latest
```

**Speech-to-text** — non-streaming transcription and translation use `OPENAI_WHISPER_MODEL` (typically `whisper-1`). Streaming transcription uses `OPENAI_TRANSCRIPTION_MODEL` (e.g. `gpt-4o-mini-transcribe`):

```bash
docker run --rm \
  -e OPENAI_BASE_URL=https://your-endpoint.example/v1 \
  -e OPENAI_API_KEY=your-api-key \
  -e OPENAI_WHISPER_MODEL=whisper-1 \
  -e OPENAI_TRANSCRIPTION_MODEL=gpt-4o-mini-transcribe \
  -e TEST_SUITES=audio_transcriptions,audio_transcriptions_stream,audio_translations \
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

**Reasoning models** — requires a reasoning-capable model (`OPENAI_REASONING_MODEL`). Included in `extended` and `full`. Validation is lenient: the suite passes when the response has assistant content or refusal, reports `reasoning_tokens` in usage, exposes non-empty `reasoning_content`, or returns a `content_filter` finish reason — so proxies that strip reasoning fields can still pass if they return normal chat output:

```bash
docker run --rm \
  -e OPENAI_BASE_URL=https://your-endpoint.example/v1 \
  -e OPENAI_API_KEY=your-api-key \
  -e OPENAI_REASONING_MODEL=o3-mini \
  -e TEST_SUITES=chat_completions_reasoning \
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