package testutil

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
)

//go:embed testdata/small.png
var smallPNG []byte

//go:embed testdata/small.wav
var smallWAV []byte

// SkillBundleFolder is the top-level directory name in skill upload zip bundles.
const SkillBundleFolder = "compatibility-test-skill"

const smallTextFileContent = "compatibility test file\n"

const smallSkillFileContent = `---
name: compatibility-test-skill
description: compatibility test skill
---

Compatibility test skill instructions.
`

const skillVersionUpdatedContent = `---
name: compatibility-test-skill
description: compatibility test skill v2
---

Compatibility test skill instructions v2.
`

type namedPNGReader struct {
	r *bytes.Reader
}

func (r *namedPNGReader) Read(p []byte) (int, error) {
	return r.r.Read(p)
}

func (r *namedPNGReader) Filename() string {
	return "test.png"
}

func (r *namedPNGReader) ContentType() string {
	return "image/png"
}

// SmallPNGBytes returns a copy of the embedded 8x8 RGBA PNG used for multipart image uploads.
func SmallPNGBytes() []byte {
	buf := make([]byte, len(smallPNG))
	copy(buf, smallPNG)
	return buf
}

// SmallPNGReader returns a multipart-ready reader for the embedded PNG fixture.
func SmallPNGReader() io.Reader {
	return &namedPNGReader{r: bytes.NewReader(smallPNG)}
}

type namedWAVReader struct {
	r *bytes.Reader
}

func (r *namedWAVReader) Read(p []byte) (int, error) {
	return r.r.Read(p)
}

func (r *namedWAVReader) Filename() string {
	return "test.wav"
}

func (r *namedWAVReader) ContentType() string {
	return "audio/wav"
}

// SmallWAVBytes returns a copy of the embedded minimal mono 8-bit WAV fixture.
func SmallWAVBytes() []byte {
	buf := make([]byte, len(smallWAV))
	copy(buf, smallWAV)
	return buf
}

// SmallWAVReader returns a multipart-ready reader for the embedded WAV fixture.
func SmallWAVReader() io.Reader {
	return &namedWAVReader{r: bytes.NewReader(smallWAV)}
}

type namedTextReader struct {
	r        *bytes.Reader
	filename string
}

func (r *namedTextReader) Read(p []byte) (int, error) {
	return r.r.Read(p)
}

func (r *namedTextReader) Filename() string {
	return r.filename
}

func (r *namedTextReader) ContentType() string {
	return "text/plain"
}

// SmallTextFileReader returns a multipart-ready reader for a small text file fixture.
func SmallTextFileReader() io.Reader {
	return &namedTextReader{
		r:        bytes.NewReader([]byte(smallTextFileContent)),
		filename: "test.txt",
	}
}

// SmallTextFileBytes returns the bytes of the small text file fixture.
func SmallTextFileBytes() []byte {
	return []byte(smallTextFileContent)
}

type namedSkillFileReader struct {
	r        *bytes.Reader
	filename string
}

func (r *namedSkillFileReader) Read(p []byte) (int, error) {
	return r.r.Read(p)
}

func (r *namedSkillFileReader) Filename() string {
	return r.filename
}

func (r *namedSkillFileReader) ContentType() string {
	return "text/markdown"
}

// SmallSkillFileReader returns a multipart-ready reader for a minimal skill bundle file.
func SmallSkillFileReader() io.Reader {
	return SkillFileReader(smallSkillFileContent)
}

// SkillVersionFileReader returns a multipart-ready reader for an updated skill bundle file.
func SkillVersionFileReader() io.Reader {
	return SkillFileReader(skillVersionUpdatedContent)
}

// SkillFileReader returns a multipart-ready reader for skill bundle content.
func SkillFileReader(content string) io.Reader {
	return &namedSkillFileReader{
		r:        bytes.NewReader([]byte(content)),
		filename: SkillBundleFolder + "/SKILL.md",
	}
}

type namedJSONLReader struct {
	r        *bytes.Reader
	filename string
}

func (r *namedJSONLReader) Read(p []byte) (int, error) {
	return r.r.Read(p)
}

func (r *namedJSONLReader) Filename() string {
	return r.filename
}

func (r *namedJSONLReader) ContentType() string {
	return "application/jsonl"
}

// SmallFineTuneJSONLReader returns a minimal chat-format JSONL file for fine-tuning jobs.
// OpenAI requires at least 10 training examples.
func SmallFineTuneJSONLReader() io.Reader {
	line, err := json.Marshal(map[string]any{
		"messages": []map[string]string{
			{"role": "system", "content": "You are a helpful assistant."},
			{"role": "user", "content": "Reply with exactly the word: pong"},
			{"role": "assistant", "content": "pong"},
		},
	})
	if err != nil {
		panic(fmt.Sprintf("marshal fine-tune jsonl: %v", err))
	}
	var buf bytes.Buffer
	for range 10 {
		buf.Write(line)
		buf.WriteByte('\n')
	}
	return &namedJSONLReader{
		r:        bytes.NewReader(buf.Bytes()),
		filename: "fine-tune.jsonl",
	}
}

// SmallBatchJSONLReader returns a minimal JSONL input file for chat completion batch jobs.
func SmallBatchJSONLReader(model string) io.Reader {
	line, err := json.Marshal(map[string]any{
		"custom_id": "batch-request-1",
		"method":    "POST",
		"url":       "/v1/chat/completions",
		"body": map[string]any{
			"model": model,
			"messages": []map[string]string{
				{"role": "user", "content": "Reply with exactly the word: pong"},
			},
		},
	})
	if err != nil {
		panic(fmt.Sprintf("marshal batch jsonl: %v", err))
	}
	return &namedJSONLReader{
		r:        bytes.NewReader(append(line, '\n')),
		filename: "batch-input.jsonl",
	}
}
