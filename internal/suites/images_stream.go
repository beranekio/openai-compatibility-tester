package suites

import (
	"fmt"

	"github.com/openai/openai-go/v3/packages/ssestream"
)

// consumeImageStream drains an image generation/edit SSE stream, validating the
// terminal "<prefix>.completed" event is reached and that no events follow it.
// Partial-image events are accepted but not required (providers may omit them).
// The info func extracts the event type and b64_json payload from a union event.
func consumeImageStream[T any](suite string, stream *ssestream.Stream[T], prefix string, info func(T) (eventType, b64JSON string)) error {
	var terminalReached bool
	for stream.Next() {
		if terminalReached {
			return fail(suite, fmt.Sprintf("stream event after terminal %s.completed event", prefix))
		}
		eventType, b64 := info(stream.Current())
		switch eventType {
		case prefix + ".partial_image":
			if b64 == "" {
				return fail(suite, "partial_image event missing b64_json")
			}
		case prefix + ".completed":
			if b64 == "" {
				return fail(suite, "completed event missing b64_json")
			}
			terminalReached = true
		default:
			return fail(suite, fmt.Sprintf("unexpected stream event type %q", eventType))
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
