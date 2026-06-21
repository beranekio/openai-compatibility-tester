# Test suites

Every suite calls a real method on the [official OpenAI Go SDK](https://github.com/openai/openai-go). A suite **passes** when the SDK can issue the request, parse the response (or stream), and satisfy basic validation rules. See the [README](../README.md) for exit codes and general usage.

## Presets

| Preset | `TEST_SUITES` value | Scope |
|--------|----------------------|-------|
| Default | `all` or `default` | `models`, `models_get`, `chat_completions`, `chat_completions_stream`, `responses`, `responses_stream` |
| Extended | `extended` | default plus streaming usage, logprobs, JSON mode, tool calling, multi-turn, completions, embeddings, vision, reasoning, moderations, images, audio, and `error_responses` |
| Full | `full` | every registered suite, including deprecated and specialized APIs |

Deprecated Assistants API suites (`assistants`, `assistants_threads`) are **opt-in** — included in `full`, but not in `default` or `extended`. They use `OPENAI_MODEL` and are labeled `(deprecated)` in `--list-suites`.

## Suite reference

| Suite | SDK surface | Endpoint |
|-------|-------------|----------|
| `models` | `client.Models.List` | `GET /v1/models` |
| `models_get` | `client.Models.Get` | `GET /v1/models/{id}` |
| `chat_completions` | `client.Chat.Completions.New` | `POST /v1/chat/completions` |
| `chat_completions_stream` | `client.Chat.Completions.NewStreaming` | `POST /v1/chat/completions` (stream) |
| `chat_completions_stream_usage` | `client.Chat.Completions.NewStreaming` (`stream_options.include_usage`) | `POST /v1/chat/completions` (stream) |
| `chat_completions_logprobs` | `client.Chat.Completions.New` (`logprobs`) | `POST /v1/chat/completions` |
| `chat_completions_json` | `client.Chat.Completions.New` (`response_format` json_schema) | `POST /v1/chat/completions` |
| `chat_completions_vision` | `client.Chat.Completions.New` (with image input) | `POST /v1/chat/completions` |
| `chat_completions_reasoning` | `client.Chat.Completions.New` (reasoning model) | `POST /v1/chat/completions` |
| `chat_completions_audio` | `client.Chat.Completions.New` (with audio output) | `POST /v1/chat/completions` |
| `chat_completions_tools` | `client.Chat.Completions.New` (with `tools`) | `POST /v1/chat/completions` |
| `chat_completions_tools_stream` | `client.Chat.Completions.NewStreaming` (with `tools`) | `POST /v1/chat/completions` (stream) |
| `chat_completions_multi_turn` | `client.Chat.Completions.New` (multi-turn history with developer and tool messages) | `POST /v1/chat/completions` |
| `chat_completions_get` | `client.Chat.Completions.Get` | `GET /v1/chat/completions/{id}` |
| `chat_completions_list` | `client.Chat.Completions.List` | `GET /v1/chat/completions` |
| `chat_completions_delete` | `client.Chat.Completions.Delete` | `DELETE /v1/chat/completions/{id}` |
| `chat_completions_messages` | `client.Chat.Completions.Messages.List` | `GET /v1/chat/completions/{id}/messages` |
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
| `responses_cancel` | `client.Responses.Cancel` | `POST /v1/responses/{id}/cancel` (passes when background create is already `completed`) |
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
| `files` | `client.Files.New`, `List`, `Get`, `Content`, `Delete` | `POST/GET/DELETE /v1/files`, `GET /v1/files/{id}/content` |
| `uploads` | `client.Uploads.New`, `Parts.New`, `Complete` | `POST /v1/uploads`, `POST /v1/uploads/{id}/parts`, `POST /v1/uploads/{id}/complete` |
| `batches_create` | `client.Batches.New` | `POST /v1/batches` |
| `batches_get` | `client.Batches.Get` | `GET /v1/batches/{id}` |
| `batches_cancel` | `client.Batches.Cancel` | `POST /v1/batches/{id}/cancel` |
| `conversations` | `client.Conversations.New`, `Get`, `Update`, `Delete`; `client.Conversations.Items.New`, `List`, `Get`, `Delete` | `POST/GET/DELETE /v1/conversations`, `POST/GET/DELETE /v1/conversations/{id}/items` |
| `vector_stores` | `client.VectorStores.New`, `Get`, `Update`, `List`, `Search`, `Delete` | `POST/GET/DELETE /v1/vector_stores`, `POST /v1/vector_stores/{id}/search` |
| `vector_store_files` | `client.VectorStores.Files.New`, `List`, `Get`, `Delete` | `POST/GET/DELETE /v1/vector_stores/{id}/files` |
| `vector_store_file_batches` | `client.VectorStores.FileBatches.New`, `Get`, `ListFiles`, `Cancel` | `POST/GET /v1/vector_stores/{id}/file_batches`, `POST /v1/vector_stores/{id}/file_batches/{batch_id}/cancel` |
| `realtime_client_secrets` | `client.Realtime.ClientSecrets.New` | `POST /v1/realtime/client_secrets` (WebSocket sessions not exercised) |
| `containers` | `client.Containers.New`, `Get`, `List`, `Delete` | `POST /v1/containers`, `GET /v1/containers`, `GET/DELETE /v1/containers/{id}` |
| `container_files` | `client.Containers.Files.New`, `List`, `Get`, `Delete`; `client.Containers.Files.Content.Get` | `POST/GET /v1/containers/{id}/files`, `GET/DELETE /v1/containers/{id}/files/{file_id}`, `GET /v1/containers/{id}/files/{file_id}/content` |
| `videos` | `client.Videos.New`, `PollStatus`, `Get`, `List`, `DownloadContent`, `Delete` | `POST/GET/DELETE /v1/videos`, `GET /v1/videos/{id}/content` |
| `skills` | `client.Skills.New`, `Get`, `Update`, `List`, `Delete`; `client.Skills.Versions.New` | `POST /v1/skills`, `GET /v1/skills`, `GET/POST/DELETE /v1/skills/{id}`, `POST /v1/skills/{id}/versions` |
| `skill_versions` | `client.Skills.Versions.New`, `Get`, `List`, `Delete`; `client.Skills.Content.Get`; `client.Skills.Versions.Content.Get` | `POST/GET /v1/skills/{id}/versions`, `GET/DELETE /v1/skills/{id}/versions/{version}`, `GET /v1/skills/{id}/content`, `GET /v1/skills/{id}/versions/{version}/content` |
| `fine_tuning` | `client.FineTuning.Jobs.New`, `List`, `Get`, `Cancel`; `client.FineTuning.Jobs.Checkpoints.List`; `client.FineTuning.Checkpoints.Permissions.List` | `POST/GET /v1/fine_tuning/jobs`, `POST /v1/fine_tuning/jobs/{id}/cancel`, `GET /v1/fine_tuning/jobs/{id}/checkpoints`, `GET /v1/fine_tuning/checkpoints/{fine_tuned_model_checkpoint}/permissions` |
| `chatkit_sessions` | `client.Beta.ChatKit.Sessions.New`, `Cancel` | `POST /v1/chatkit/sessions`, `POST /v1/chatkit/sessions/{id}/cancel` |
| `chatkit_threads` | `client.Beta.ChatKit.Threads.List`, `Get`, `ListItems`[, `Delete`] | `GET /v1/chatkit/threads`, `GET /v1/chatkit/threads/{id}`, `GET /v1/chatkit/threads/{id}/items`[, `DELETE /v1/chatkit/threads/{id}`] |
| `(deprecated) assistants` | `client.Beta.Assistants.New`, `Get`, `Update`, `List`, `Delete` | `GET/POST /v1/assistants`, `GET/POST/DELETE /v1/assistants/{id}` |
| `(deprecated) assistants_threads` | `client.Beta.Threads.New`, `Get`, `Update`, `Delete`; `client.Beta.Threads.Messages.New`, `List`, `Get`; `client.Beta.Threads.Runs.New`, `Get` | `POST /v1/threads`, `GET/POST/DELETE /v1/threads/{id}`, `POST/GET /v1/threads/{id}/messages`, `GET /v1/threads/{id}/messages/{message_id}`, `POST /v1/threads/{id}/runs`, `GET /v1/threads/{id}/runs/{run_id}` |
| `error_responses` | `client.Chat.Completions.New` (invalid model) | `POST /v1/chat/completions` |

## Suite-specific model configuration

Most suites reuse `OPENAI_MODEL` / `OPENAI_RESPONSES_MODEL`. The following variables are only required for the suites listed below.

| Variable | Required by | Default | Notes |
|----------|-------------|---------|-------|
| `OPENAI_RESPONSES_MODEL` | Responses suites | same as `OPENAI_MODEL` | Model used for Responses API suites |
| `OPENAI_COMPLETION_MODEL` | `completions`, `completions_stream` | `gpt-3.5-turbo-instruct` when selected, otherwise same as `OPENAI_MODEL` | Legacy completions |
| `OPENAI_EMBEDDING_MODEL` | `embeddings`, `embeddings_batch` | — | Embedding model |
| `OPENAI_VISION_MODEL` | `chat_completions_vision` | same as `OPENAI_MODEL` | Vision-capable model |
| `OPENAI_REASONING_MODEL` | `chat_completions_reasoning` | — | e.g. `o3-mini`, `o4-mini` |
| `OPENAI_IMAGE_MODEL` | `images_generations`, `images_edits` | — | GPT Image models or `dall-e-2`; `dall-e-3` not supported for edits |
| `OPENAI_VIDEO_MODEL` | `videos` | — | e.g. `sora-2` |
| `OPENAI_TTS_MODEL` | `audio_speech` | — | Text-to-speech model |
| `OPENAI_WHISPER_MODEL` | `audio_transcriptions`, `audio_translations` | — | e.g. `whisper-1` |
| `OPENAI_TRANSCRIPTION_MODEL` | `audio_transcriptions_stream` | — | e.g. `gpt-4o-mini-transcribe` |
| `OPENAI_REALTIME_MODEL` | `realtime_client_secrets` | `gpt-realtime` | Realtime API |
| `OPENAI_ADMIN_API_KEY` | `fine_tuning` (permissions only) | — | Skipped when unset |
| `OPENAI_CHATKIT_WORKFLOW_ID` | `chatkit_sessions` | `wf_mock_compat_test` (when selected) | Set explicitly for real endpoints |
| `OPENAI_CHATKIT_TEST_THREAD_ID` | `chatkit_threads` (delete only) | — | Disposable thread ID; omit for read-only checks |

## Examples

The base image is `ghcr.io/beranekio/openai-compatibility-tester:latest`. Every example below assumes `OPENAI_BASE_URL` and `OPENAI_API_KEY` are set; the variables relevant to each suite are highlighted.

### Default suites

```bash
docker run --rm \
  -e OPENAI_BASE_URL=https://your-endpoint.example/v1 \
  -e OPENAI_API_KEY=your-api-key \
  -e OPENAI_MODEL=your-chat-model \
  ghcr.io/beranekio/openai-compatibility-tester:latest
```

### Run a subset

```bash
docker run --rm \
  -e OPENAI_BASE_URL=https://your-endpoint.example/v1 \
  -e OPENAI_API_KEY=your-api-key \
  -e TEST_SUITES=models,chat_completions \
  ghcr.io/beranekio/openai-compatibility-tester:latest
```

### Tool calling and structured output

Tool-calling suites (`chat_completions_tools`, `chat_completions_tools_stream`, `responses_tools`, `responses_tools_stream`) and structured JSON output (`chat_completions_json`, `responses_json`) are included in `extended` and `full`, but not in the default set.

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

### Legacy completions

Add `completions` and `completions_stream` only when your endpoint exposes `/v1/completions`:

```bash
docker run --rm \
  -e OPENAI_BASE_URL=https://your-endpoint.example/v1 \
  -e OPENAI_API_KEY=your-api-key \
  -e OPENAI_COMPLETION_MODEL=your-instruct-model \
  -e TEST_SUITES=models,completions,completions_stream \
  ghcr.io/beranekio/openai-compatibility-tester:latest
```

### Embeddings

```bash
docker run --rm \
  -e OPENAI_BASE_URL=https://your-endpoint.example/v1 \
  -e OPENAI_API_KEY=your-api-key \
  -e OPENAI_MODEL=your-chat-model \
  -e OPENAI_EMBEDDING_MODEL=your-embedding-model \
  -e TEST_SUITES=models,chat_completions,chat_completions_stream,responses,responses_stream,embeddings,embeddings_batch \
  ghcr.io/beranekio/openai-compatibility-tester:latest
```

### Vision

Requires a vision-capable model (`OPENAI_VISION_MODEL` defaults to `OPENAI_MODEL`):

```bash
docker run --rm \
  -e OPENAI_BASE_URL=https://your-endpoint.example/v1 \
  -e OPENAI_API_KEY=your-api-key \
  -e OPENAI_MODEL=your-vision-model \
  -e TEST_SUITES=chat_completions_vision \
  ghcr.io/beranekio/openai-compatibility-tester:latest
```

### Reasoning models

Requires a reasoning-capable model (`OPENAI_REASONING_MODEL`). Validation is lenient: passes when the response has assistant content or refusal, reports `reasoning_tokens` in usage, exposes non-empty `reasoning_content`, or returns a `content_filter` finish reason — so proxies that strip reasoning fields can still pass if they return normal chat output:

```bash
docker run --rm \
  -e OPENAI_BASE_URL=https://your-endpoint.example/v1 \
  -e OPENAI_API_KEY=your-api-key \
  -e OPENAI_REASONING_MODEL=o3-mini \
  -e TEST_SUITES=chat_completions_reasoning \
  ghcr.io/beranekio/openai-compatibility-tester:latest
```

### Image generation

Requires `OPENAI_IMAGE_MODEL`. `images_edits` uses GPT Image models or `dall-e-2` (`dall-e-3` is not supported for edits):

```bash
docker run --rm \
  -e OPENAI_BASE_URL=https://your-endpoint.example/v1 \
  -e OPENAI_API_KEY=your-api-key \
  -e OPENAI_IMAGE_MODEL=your-image-model \
  -e TEST_SUITES=images_generations \
  ghcr.io/beranekio/openai-compatibility-tester:latest
```

`images_variations` is a legacy DALL-E 2 suite (always requests `dall-e-2`, does not use `OPENAI_IMAGE_MODEL`). Official OpenAI retired DALL-E models in May 2026; included in `full` but not `extended`:

```bash
docker run --rm \
  -e OPENAI_BASE_URL=https://your-endpoint.example/v1 \
  -e OPENAI_API_KEY=your-api-key \
  -e TEST_SUITES=images_variations \
  ghcr.io/beranekio/openai-compatibility-tester:latest
```

### Video generation

Requires `OPENAI_VIDEO_MODEL`. The suite polls until the job completes; increase `REQUEST_TIMEOUT` for slow endpoints:

```bash
docker run --rm \
  -e OPENAI_BASE_URL=https://your-endpoint.example/v1 \
  -e OPENAI_API_KEY=your-api-key \
  -e OPENAI_VIDEO_MODEL=sora-2 \
  -e REQUEST_TIMEOUT=10m \
  -e TEST_SUITES=videos \
  ghcr.io/beranekio/openai-compatibility-tester:latest
```

### Text-to-speech

Requires `OPENAI_TTS_MODEL`:

```bash
docker run --rm \
  -e OPENAI_BASE_URL=https://your-endpoint.example/v1 \
  -e OPENAI_API_KEY=your-api-key \
  -e OPENAI_TTS_MODEL=tts-1 \
  -e TEST_SUITES=audio_speech \
  ghcr.io/beranekio/openai-compatibility-tester:latest
```

### Speech-to-text

Non-streaming transcription and translation use `OPENAI_WHISPER_MODEL` (typically `whisper-1`). Streaming transcription uses `OPENAI_TRANSCRIPTION_MODEL` (e.g. `gpt-4o-mini-transcribe`):

```bash
docker run --rm \
  -e OPENAI_BASE_URL=https://your-endpoint.example/v1 \
  -e OPENAI_API_KEY=your-api-key \
  -e OPENAI_WHISPER_MODEL=whisper-1 \
  -e OPENAI_TRANSCRIPTION_MODEL=gpt-4o-mini-transcribe \
  -e TEST_SUITES=audio_transcriptions,audio_transcriptions_stream,audio_translations \
  ghcr.io/beranekio/openai-compatibility-tester:latest
```

### Realtime API

Opt-in HTTP smoke test for client secret creation (`realtime_client_secrets`). WebSocket sessions are not exercised. Uses `OPENAI_REALTIME_MODEL` (defaults to `gpt-realtime`):

```bash
docker run --rm \
  -e OPENAI_BASE_URL=https://your-endpoint.example/v1 \
  -e OPENAI_API_KEY=your-api-key \
  -e OPENAI_REALTIME_MODEL=gpt-4o-realtime-preview \
  -e TEST_SUITES=realtime_client_secrets \
  ghcr.io/beranekio/openai-compatibility-tester:latest
```

### Fine-tuning

Opt-in only (included in `full`, not `default` or `extended`). Uploads a minimal training JSONL file (10 examples), creates a job, lists/gets it, lists checkpoints once, optionally smoke-tests checkpoint permissions when `OPENAI_ADMIN_API_KEY` is set, and cancels the job. It does **not** poll for training completion; an empty checkpoint list is valid. Job create/cancel can still incur **cost** on real endpoints. Set `OPENAI_ADMIN_API_KEY` only when you also want to exercise the admin-only permissions endpoint:

```bash
docker run --rm \
  -e OPENAI_BASE_URL=https://your-endpoint.example/v1 \
  -e OPENAI_API_KEY=your-api-key \
  -e OPENAI_ADMIN_API_KEY=your-admin-api-key \
  -e OPENAI_MODEL=gpt-4o-mini \
  -e TEST_SUITES=fine_tuning \
  ghcr.io/beranekio/openai-compatibility-tester:latest
```

### Beta ChatKit

Opt-in only (included in `full`, not `default` or `extended`). The SDK sends `OpenAI-Beta: chatkit_beta=v1` on these requests. `chatkit_threads` is read-only by default (list, get, list items). Set `OPENAI_CHATKIT_TEST_THREAD_ID` to a disposable thread ID to also exercise delete:

```bash
docker run --rm \
  -e OPENAI_BASE_URL=https://your-endpoint.example/v1 \
  -e OPENAI_API_KEY=your-api-key \
  -e OPENAI_CHATKIT_WORKFLOW_ID=your-workflow-id \
  -e OPENAI_CHATKIT_TEST_THREAD_ID=your-disposable-thread-id \
  -e TEST_SUITES=chatkit_sessions,chatkit_threads \
  ghcr.io/beranekio/openai-compatibility-tester:latest
```

### Deprecated Assistants API

Opt-in only (included in `full`). Uses `OPENAI_MODEL` for the assistant model and is labeled `(deprecated)` in `--list-suites`:

```bash
docker run --rm \
  -e OPENAI_BASE_URL=https://your-endpoint.example/v1 \
  -e OPENAI_API_KEY=your-api-key \
  -e OPENAI_MODEL=your-chat-model \
  -e TEST_SUITES=assistants,assistants_threads \
  ghcr.io/beranekio/openai-compatibility-tester:latest
```
