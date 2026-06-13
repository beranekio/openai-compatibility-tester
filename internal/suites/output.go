package suites

import (
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

func responseOutputRefusal(resp *responses.Response) string {
	if resp == nil {
		return ""
	}
	var b strings.Builder
	for _, item := range resp.Output {
		for _, content := range item.Content {
			if content.Type == "refusal" {
				b.WriteString(content.Refusal)
			}
		}
	}
	return b.String()
}

func hasResponseOutput(resp *responses.Response) bool {
	if resp == nil {
		return false
	}
	return resp.OutputText() != "" || responseOutputRefusal(resp) != ""
}

func hasChatMessageOutput(msg openai.ChatCompletionMessage) bool {
	return msg.Content != "" || msg.Refusal != ""
}