package suites

import (
	"fmt"

	"github.com/openai/openai-go/v3"
)

func validateImagesResponse(suite string, resp *openai.ImagesResponse) error {
	if resp == nil {
		return fail(suite, "response is nil")
	}
	if !resp.JSON.Created.Valid() {
		return fail(suite, "response missing created")
	}
	if len(resp.Data) == 0 {
		return fail(suite, "response missing data")
	}
	image := resp.Data[0]
	if image.URL == "" && image.B64JSON == "" {
		return fail(suite, "image missing url and b64_json")
	}
	if image.URL != "" && !image.JSON.URL.Valid() {
		return fail(suite, fmt.Sprintf("image url field invalid: %q", image.URL))
	}
	if image.B64JSON != "" && !image.JSON.B64JSON.Valid() {
		return fail(suite, "image b64_json field invalid")
	}
	return nil
}
