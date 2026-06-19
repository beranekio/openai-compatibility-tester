package suites

import (
	"net/http"

	"github.com/beranekio/openai-compatibility-tester/internal/testutil"
)

func validateBinaryHTTPResponse(suite string, resp *http.Response, minBytes int) error {
	if err := testutil.ValidateBinaryHTTPResponse(resp, minBytes); err != nil {
		return fail(suite, err.Error())
	}
	return nil
}

func validateBase64Data(suite string, data string, minBytes int) error {
	if err := testutil.ValidateBase64Data(data, minBytes); err != nil {
		return fail(suite, err.Error())
	}
	return nil
}

func validateWAVBytes(suite string, data []byte) error {
	if err := testutil.ValidateWAVBytes(data); err != nil {
		return fail(suite, err.Error())
	}
	return nil
}

func validateBase64WAVData(suite string, data string, minBytes int) error {
	if err := testutil.ValidateBase64WAVData(data, minBytes); err != nil {
		return fail(suite, err.Error())
	}
	return nil
}
