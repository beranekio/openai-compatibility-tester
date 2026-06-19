# AGENTS.md

Guidance for AI coding agents working in this repository.

## Project purpose

`openai-compatibility-tester` is a Go CLI (and Docker image) that checks whether an HTTP endpoint is compatible with the [OpenAI API](https://platform.openai.com/docs/api-reference) by exercising it through the [official OpenAI Go SDK](https://github.com/openai/openai-go) (`github.com/openai/openai-go/v3`).

A suite **passes** when:

1. The SDK can issue the request without client-side errors.
2. The SDK can parse the response (or stream events) into typed structs.
3. Basic response validation rules in the suite are satisfied.

The process exits `0` when all selected suites pass, `1` when any suite fails compatibility checks, and `2` on configuration or runner errors.

## Repository layout

```
cmd/openai-compatibility-tester/   CLI entrypoint
internal/
  config/                          Env/flag parsing, suite selection, validation
  runner/                          SDK client setup, suite orchestration, reporting
  suites/                          One file per suite; shared helpers (stream, output, errors)
  mockserver/                      In-process OpenAI-compatible HTTP server for tests
```

There is no `pkg/` export surface. Keep new code in `internal/`.

## Architecture

```
main → config.Load → runner.RunAll → suites.Suite.Run (per suite)
                              ↓
                    openai.NewClient(option.WithBaseURL, WithAPIKey, WithMaxRetries(0))
```

Each suite implements:

```go
type Suite interface {
    Name() string
    Description() string
    Run(ctx context.Context, client openai.Client, cfg *config.Config) error
}
```

Register new suites in `internal/suites/suite.go` (`All()`, `knownSuites` in config if needed, `RequiredModels()` / `validateModelsForSuites()` when model config is required).

## Adding a new test suite

Follow this checklist for every new suite:

1. **Create** `internal/suites/<name>.go` with a stateless struct (see `models.go`, `chat_completions.go`).
2. **Use the official SDK** — call `client.<Service>.<Method>`; do not hand-craft HTTP requests in suites.
3. **Validate** parsed responses with `fail(suite, message)` from `errors.go`; wrap transport/SDK errors with `fmt.Errorf("...: %w", err)`.
4. **Register** the suite in `suites.All()` and update:
   - `config.DefaultSuites` (only if it should run by default)
   - `config.ExtendedSuites` and `config.FullSuites` (keep `FullSuites` in sync — `internal/suites/suite_test.go` enforces this)
   - `config.knownSuites`
   - `suites.RequiredModels()` and `config.validateModelsForSuites()` (if a model env var is needed)
   - `config.Load()` flags/env vars (if new settings are required)
5. **Extend** `internal/mockserver/server.go` with a handler so CI stays offline.
6. **Test** — add or extend `internal/runner/runner_test.go` to run the new suite against the mock server. If step 4 changed config (flags, env vars, presets, validation), add or update cases in `internal/config/config_test.go` too — `runner_test.go` constructs `config.Config` directly and does not exercise `config.Load()`.
7. **Document** — update `README.md` suite table and env var table.

### Suite design principles

- **Minimal requests** — use the smallest prompt/input that exercises the endpoint (e.g. "Reply with exactly the word: pong").
- **Lenient where providers differ** — accept `content_filter` finish reasons and refusals as valid outcomes (see `output.go`, `isContentFilterFinishReason`).
- **Streaming** — reuse `validateEventStreamContentType` and chunk validators from `stream.go`; always check for a terminal event. Chat: `finish_reason` (including `content_filter`). Responses: `response.completed`, or `response.incomplete` when `isContentFilterIncompleteResponse` applies (see `responses_stream.go`).
- **No retries** — the runner sets `option.WithMaxRetries(0)`; suites should not enable retries.
- **No live OpenAI calls in unit tests** — use `mockserver` only.
- **Per-suite timeout** — suites receive a context from `runner` bounded by `cfg.RequestTimeout`.

### Shared helpers

| File | Use for |
|------|---------|
| `errors.go` | `fail()`, `ValidationError` |
| `output.go` | Content/refusal detection, content-filter incomplete responses |
| `stream.go` | SSE content-type and chat completion chunk validation |

Prefer extending shared helpers over duplicating validation logic across suites.

## Configuration conventions

| Env var | Purpose |
|---------|---------|
| `OPENAI_BASE_URL` | Required. Conventionally ends with `/v1` (see README); SDK appends paths relative to this base. No query params. |
| `OPENAI_API_KEY` | Required when running suites. Not required for `--list-suites`. Bearer token. |
| `OPENAI_MODEL` | Chat completion suites (default `gpt-4o-mini`) |
| `OPENAI_RESPONSES_MODEL` | Responses suites (defaults to `OPENAI_MODEL`) |
| `OPENAI_COMPLETION_MODEL` | Legacy completions (defaults to `gpt-3.5-turbo-instruct` when selected) |
| `OPENAI_EMBEDDING_MODEL` | Required when `embeddings` is selected |
| `OPENAI_REALTIME_MODEL` | Realtime API suites (defaults to `gpt-realtime`) |
| `OPENAI_VIDEO_MODEL` | Required when `videos` is selected |
| `OPENAI_CHATKIT_WORKFLOW_ID` | ChatKit sessions workflow (default `wf_mock_compat_test` when `chatkit_sessions` selected) |
| `OPENAI_CHATKIT_TEST_THREAD_ID` | Optional disposable thread for `chatkit_threads` delete test |
| `TEST_SUITES` | Comma-separated names or `all` |
| `REQUEST_TIMEOUT` | Per-suite timeout (default `2m`) |
| `ALLOW_INSECURE_HTTP` | Allow non-loopback `http://` |

Reuse existing model settings when the suite belongs to an established family (`OPENAI_MODEL` for chat, `OPENAI_RESPONSES_MODEL` for Responses, etc.). Add a dedicated env var and `validateModelsForSuites` entry only when the suite needs a genuinely different model category (e.g. vision, image generation, TTS). Planned presets (`extended`, `full`) are tracked in [#45](https://github.com/beranekio/openai-compatibility-tester/issues/45).

## Testing

```bash
go test ./...
go build -o bin/openai-compatibility-tester ./cmd/openai-compatibility-tester
```

`internal/config/config_test.go` covers flag/env parsing. `internal/runner/runner_test.go` runs suites against `mockserver.New()` and `mockserver.BrokenServer()`.

**Every new suite must have a mock handler.** CI runs `go test ./...`, builds the binary, and builds the Docker image (see `.github/workflows/ci.yml`); it does not hit real APIs.

Local smoke test against the mock server:

```bash
go build -o bin/openai-compatibility-tester ./cmd/openai-compatibility-tester
# In another terminal: go test ./internal/runner -run TestRunAllPasses -v  # uses embedded mock
```

## CI and Docker

- GitHub Actions (`.github/workflows/ci.yml`): `go test ./...`, binary build, Docker build on every PR/push to `main`.
- Pushes to `main` publish `ghcr.io/beranekio/openai-compatibility-tester:latest`.
- Dockerfile: multi-stage, distroless nonroot image, entrypoint is the binary.

Do not break the Docker entrypoint contract (no shell wrapper; flags/env only).

## Code style

- Go 1.24+ (`go.mod`). Match existing package naming and file layout.
- Stateless suite structs with value receivers for `Name`/`Description`/`Run`.
- Wrap errors with context; use `fail()` for compatibility validation failures.
- Keep suites focused — one SDK method family per suite file.
- Avoid drive-by refactors, new dependencies, or unrelated files in a PR.
- Comments only where non-obvious; no docstrings on trivial helpers.

## Roadmap and issue tracking

Expansion work is tracked in GitHub issues #7–#47, organized into milestones:

| Milestone | Focus |
|-----------|-------|
| Sprint 1 | Config, tool calling, JSON mode, models get, mock parity |
| Sprint 2 | Completions stream, embeddings batch, Responses tools, errors |
| Sprint 3 | Vision, Responses lifecycle, moderations |
| Extended tier | Images, audio, chat advanced |
| Full tier | Files, batches, vector stores, specialized APIs |
| Infrastructure | Multipart helpers, pagination, auth headers |

Each issue has a **Dependencies** section (`Blocked by` / `Blocks`). Check [#48](https://github.com/beranekio/openai-compatibility-tester/issues/48) for the overview graph before starting work.

Labels: `phase-1` … `phase-8`, `suite`, `infrastructure`, `enhancement`.

## Out of scope

Do not add suites for:

- **Admin API** (`client.Admin.*`) — organization management, not inference proxy compatibility.
- **Webhook verification** — server-side concern.
- **SDK retry behavior** — intentionally disabled in the runner.

## Common pitfalls

- **Base URL** — conventionally ends with `/v1` (not enforced by `validateBaseURL`); SDK appends paths like `chat/completions`. Query strings are rejected.
- **Default suites** — `completions` and `embeddings` exist but are not in `DefaultSuites`; changing defaults affects all Docker users.
- **Content filter** — empty text with `finish_reason: content_filter` is a pass, not a fail.
- **Responses stream** — terminal events must not be followed by more events; see `responses_stream.go`.
- **Mock parity** — forgetting to update `mockserver` breaks CI even if suite code is correct.
- **SDK version** — bump `github.com/openai/openai-go/v3` in `go.mod` only when needed; run `go test ./...` after.

## PR checklist

- [ ] `go test ./...` passes
- [ ] New suite registered in `suite.go` (+ config/README if needed)
- [ ] Mock server handler added
- [ ] `runner_test.go` includes new suite in `TestRunAllPassesAgainstMockServer`
- [ ] `config_test.go` updated if config parsing, validation, or presets changed
- [ ] README updated for user-facing changes
- [ ] Focused diff — no unrelated changes