package suites

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
)

// smallPNG is an 8x8 RGBA PNG with transparent pixels for DALL-E 2 edit compatibility.
var smallPNG = func() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			if x < 4 {
				img.SetRGBA(x, y, color.RGBA{R: 255, A: 0})
			} else {
				img.SetRGBA(x, y, color.RGBA{R: 255, A: 255})
			}
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		panic(err)
	}
	return buf.Bytes()
}()

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

func smallPNGReader() io.Reader {
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

func smallWAVReader() io.Reader {
	return &namedWAVReader{r: bytes.NewReader(smallWAVBytes())}
}

const smallTextFileContent = "compatibility test file\n"

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

func smallTextFileReader() io.Reader {
	return &namedTextReader{
		r:        bytes.NewReader([]byte(smallTextFileContent)),
		filename: "test.txt",
	}
}

func smallTextFileBytes() []byte {
	return []byte(smallTextFileContent)
}

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

func smallSkillFileReader() io.Reader {
	return skillFileReader(smallSkillFileContent)
}

func skillVersionFileReader() io.Reader {
	return skillFileReader(skillVersionUpdatedContent)
}

func skillFileReader(content string) io.Reader {
	return &namedSkillFileReader{
		r:        bytes.NewReader([]byte(content)),
		filename: skillBundleFolder + "/SKILL.md",
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

// smallBatchJSONLReader returns a minimal JSONL input file for chat completion batch jobs.
func smallBatchJSONLReader(model string) io.Reader {
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

// smallWAVBytes returns a minimal mono 8-bit WAV file for multipart audio upload tests.
func smallWAVBytes() []byte {
	const (
		sampleRate    = uint32(8000)
		numSamples    = uint32(4000) // 0.5s at 8 kHz
		bitsPerSample = uint16(8)
		numChannels   = uint16(1)
	)
	dataSize := numSamples
	fileSize := uint32(36 + dataSize)

	var b bytes.Buffer
	b.WriteString("RIFF")
	_ = binary.Write(&b, binary.LittleEndian, fileSize)
	b.WriteString("WAVE")
	b.WriteString("fmt ")
	_ = binary.Write(&b, binary.LittleEndian, uint32(16))
	_ = binary.Write(&b, binary.LittleEndian, uint16(1))
	_ = binary.Write(&b, binary.LittleEndian, numChannels)
	_ = binary.Write(&b, binary.LittleEndian, sampleRate)
	_ = binary.Write(&b, binary.LittleEndian, sampleRate*uint32(numChannels)*uint32(bitsPerSample)/8)
	_ = binary.Write(&b, binary.LittleEndian, uint16(numChannels*bitsPerSample/8))
	_ = binary.Write(&b, binary.LittleEndian, bitsPerSample)
	b.WriteString("data")
	_ = binary.Write(&b, binary.LittleEndian, dataSize)
	_, _ = b.Write(make([]byte, dataSize))
	return b.Bytes()
}