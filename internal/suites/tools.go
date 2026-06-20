package suites

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
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

func noToolChoice() openai.ChatCompletionToolChoiceOptionUnionParam {
	return openai.ChatCompletionToolChoiceOptionUnionParam{
		OfAuto: openai.String(string(openai.ChatCompletionToolChoiceOptionAutoNone)),
	}
}

func weatherResponseTools() []responses.ToolUnionParam {
	return []responses.ToolUnionParam{{
		OfFunction: &responses.FunctionToolParam{
			Name:        weatherToolName,
			Description: openai.String("Get the current weather for a location."),
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"location": map[string]any{
						"type":        "string",
						"description": "City and state, e.g. San Francisco, CA",
					},
				},
				"required": []string{"location"},
			},
		},
	}}
}

func requiredResponseToolChoice() responses.ResponseNewParamsToolChoiceUnion {
	return responses.ResponseNewParamsToolChoiceUnion{
		OfToolChoiceMode: openai.Opt(responses.ToolChoiceOptionsRequired),
	}
}

func validateWeatherToolArguments(suite string, arguments string) error {
	if arguments == "" {
		return fail(suite, "function_call missing arguments")
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(arguments), &parsed); err != nil {
		return fail(suite, fmt.Sprintf("function_call arguments are not valid JSON: %v", err))
	}
	location, ok := parsed["location"]
	if !ok {
		return fail(suite, `function_call arguments missing required "location" field`)
	}
	locationStr, ok := location.(string)
	if !ok {
		return fail(suite, `"location" field is not a string`)
	}
	if strings.TrimSpace(locationStr) == "" {
		return fail(suite, `"location" field is empty`)
	}
	return nil
}
