package suites

import (
	"fmt"

	"github.com/openai/openai-go/v3/packages/ssestream"
)

// imageStreamEventInfo carries the fields consumeImageStream validates, extracted
// from a typed stream event union by the per-suite info func.
type imageStreamEventInfo struct {
	eventType              string
	b64JSON                string
	createdAtValid         bool
	outputFormatValid      bool
	sizeValid              bool
	partialImageIndexValid bool
	partialImageIndex      int64
}

// consumeImageStream drains an image generation/edit SSE stream, validating the
// terminal "<prefix>.completed" event is reached and that no events follow it.
// Partial-image events are accepted but not required (providers may omit them).
// Each event must carry the render-critical metadata (b64_json, created_at,
// output_format, size) that the OpenAI API marks required on these events.
func consumeImageStream[T any](suite string, stream *ssestream.Stream[T], prefix string, info func(T) imageStreamEventInfo) error {
	var terminalReached bool
	var nextPartialIndex int64
	for stream.Next() {
		ev := info(stream.Current())
		if terminalReached {
			return fail(suite, fmt.Sprintf("stream event %q after terminal %s.completed event", ev.eventType, prefix))
		}
		switch ev.eventType {
		case prefix + ".partial_image":
			if err := validateImageStreamEvent(suite, "partial_image", ev); err != nil {
				return err
			}
			// partial_image_index is 0-based and increments across partials;
			// clients use it to order partial images.
			if ev.partialImageIndex != nextPartialIndex {
				return fail(suite, fmt.Sprintf("partial_image index is %d, want %d", ev.partialImageIndex, nextPartialIndex))
			}
			nextPartialIndex++
		case prefix + ".completed":
			if err := validateImageStreamEvent(suite, "completed", ev); err != nil {
				return err
			}
			terminalReached = true
		default:
			return fail(suite, fmt.Sprintf("unexpected stream event type %q", ev.eventType))
		}
	}
	if err := stream.Err(); err != nil {
		return fmt.Errorf("%s stream failed: %w", suite, err)
	}
	if !terminalReached {
		return fail(suite, fmt.Sprintf("stream ended without %s.completed event", prefix))
	}
	return nil
}

// validateImageStreamEvent checks the render-critical metadata on an image
// stream event. quality/background are intentionally not enforced (informational
// fields where providers may reasonably differ).
func validateImageStreamEvent(suite, kind string, ev imageStreamEventInfo) error {
	if ev.b64JSON == "" {
		return fail(suite, kind+" event missing b64_json")
	}
	if !ev.createdAtValid {
		return fail(suite, kind+" event missing created_at")
	}
	if !ev.outputFormatValid {
		return fail(suite, kind+" event missing output_format")
	}
	if !ev.sizeValid {
		return fail(suite, kind+" event missing size")
	}
	if kind == "partial_image" && !ev.partialImageIndexValid {
		return fail(suite, "partial_image event missing partial_image_index")
	}
	return nil
}
