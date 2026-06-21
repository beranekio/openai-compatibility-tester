# openai-compatibility-tester

Docker container that checks whether an arbitrary HTTP endpoint is compatible with the [OpenAI API](https://platform.openai.com/docs/api-reference) by exercising it through the [official OpenAI Go SDK](https://github.com/openai/openai-go).

Each test suite calls a real SDK method (models list, chat completions, Responses API, embeddings, and more). If the endpoint returns payloads the SDK cannot parse, or responses that fail basic validation, the process exits with a non-zero status code — making the image suitable for CI gates and compatibility smoke tests.

## Quick start

```bash
docker run --rm \
  -e OPENAI_BASE_URL=https://your-endpoint.example/v1 \
  -e OPENAI_API_KEY=your-api-key \
  -e OPENAI_MODEL=your-chat-model \
  ghcr.io/beranekio/openai-compatibility-tester:latest
```

This runs the default suites (`models`, `models_get`, `chat_completions`, `chat_completions_stream`, `responses`, `responses_stream`). To see every available suite:

```bash
docker run --rm ghcr.io/beranekio/openai-compatibility-tester:latest --list-suites
```

## Configuration

All settings can be passed as environment variables or CLI flags.

| Variable | Flag | Required | Default | Description |
|----------|------|----------|---------|-------------|
| `OPENAI_BASE_URL` | `--base-url` | yes | — | API base URL including `/v1` (e.g. `https://api.openai.com/v1`). Query parameters are not supported. |
| `OPENAI_API_KEY` | `--api-key` | yes | — | Bearer token sent to the endpoint |
| `OPENAI_MODEL` | `--model` | no | `gpt-4o-mini` | Model for chat completion suites and the model ID fetched by `models_get` |
| `TEST_SUITES` | `--suites` | no | `all` | Comma-separated suite names, or preset: `all`/`default`, `extended`, `full` |
| `REQUEST_TIMEOUT` | `--timeout` | no | `2m` | Per-suite request timeout (batch suites may need a longer value against real APIs while jobs finish) |
| `ALLOW_INSECURE_HTTP` | `--allow-insecure-http` | no | `false` | Allow plaintext `http://` to non-loopback hosts (loopback HTTP is always permitted) |
| `OPENAI_ORG_ID` | `--org-id` | no | — | OpenAI organization ID sent as `OpenAI-Organization` when set |
| `OPENAI_PROJECT_ID` | `--project-id` | no | — | OpenAI project ID sent as `OpenAI-Project` when set |

Some suites require additional model variables (vision, image, audio, video, etc.). See the [suite-specific configuration](docs/suites.md#suite-specific-model-configuration) for the full list.

## Selecting suites

Use `TEST_SUITES` with a preset or an explicit comma-separated list.

| Preset | Scope |
|--------|-------|
| `all` / `default` | Core chat, models, and Responses suites |
| `extended` | default plus tools, JSON, streaming variants, embeddings, vision, reasoning, moderations, images, and audio |
| `full` | every registered suite, including deprecated and specialized APIs |

```bash
# A subset
docker run --rm \
  -e OPENAI_BASE_URL=https://your-endpoint.example/v1 \
  -e OPENAI_API_KEY=your-api-key \
  -e TEST_SUITES=models,chat_completions \
  ghcr.io/beranekio/openai-compatibility-tester:latest
```

For the complete suite catalog, presets, and per-suite examples, see **[docs/suites.md](docs/suites.md)**.

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

GitHub Actions runs unit tests and builds the Docker image on every push and pull request to `main`. When tests pass on a push to `main`, a multi-architecture image (`linux/amd64`, `linux/arm64`) is published to GHCR:

`ghcr.io/beranekio/openai-compatibility-tester:latest`
