package suites

import (
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/shared"
)

const weatherToolName = "get_weather"

func getWeatherToolDefinition() shared.FunctionDefinitionParam {
	return shared.FunctionDefinitionParam{
		Name:        weatherToolName,
		Description: openai.String("Get the current weather for a location."),
		Parameters: shared.FunctionParameters{
			"type": "object",
			"properties": map[string]any{
				"location": map[string]any{
					"type":        "string",
					"description": "City and state, e.g. San Francisco, CA",
				},
			},
			"required": []string{"location"},
		},
	}
}

func weatherTools() []openai.ChatCompletionToolUnionParam {
	return []openai.ChatCompletionToolUnionParam{
		openai.ChatCompletionFunctionTool(getWeatherToolDefinition()),
	}
}

func requiredToolChoice() openai.ChatCompletionToolChoiceOptionUnionParam {
	return openai.ChatCompletionToolChoiceOptionUnionParam{
		OfAuto: openai.String("required"),
	}
}