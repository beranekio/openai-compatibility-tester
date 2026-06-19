package suites

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"
)

const (
	skillBundleFolder        = "compatibility-test-skill"
	skillContentMarker       = "Compatibility test skill instructions."
	skillVersionContentMarker = "Compatibility test skill instructions v2."
)

func validateSkillZipContentResponse(suite string, resp *http.Response, wantInSkillMD string) error {
	if resp == nil {
		return fail(suite, "response is nil")
	}
	if resp.Body == nil {
		return fail(suite, "response body is nil")
	}
	if resp.StatusCode != http.StatusOK {
		return fail(suite, fmt.Sprintf("content status is %d, want 200", resp.StatusCode))
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("%s: read response body: %w", suite, err)
	}
	if len(body) == 0 {
		return fail(suite, "response body is empty")
	}

	contentType := resp.Header.Get("Content-Type")
	if strings.TrimSpace(contentType) == "" {
		return fail(suite, "response missing Content-Type")
	}
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return fail(suite, fmt.Sprintf("Content-Type %q is invalid: %v", contentType, err))
	}
	if !isSkillContentMediaType(mediaType) {
		return fail(suite, fmt.Sprintf("Content-Type is %q, want text/*, application/octet-stream, application/zip, application/binary, or application/json", mediaType))
	}

	skillMD, err := skillMDFromZipBundle(body)
	if err != nil {
		return fail(suite, fmt.Sprintf("skill content is not a valid zip bundle: %v", err))
	}
	if wantInSkillMD != "" && !strings.Contains(string(skillMD), wantInSkillMD) {
		return fail(suite, fmt.Sprintf("skill bundle SKILL.md missing %q", wantInSkillMD))
	}
	return nil
}

func isSkillContentMediaType(mediaType string) bool {
	switch {
	case strings.HasPrefix(mediaType, "text/"):
		return true
	case mediaType == "application/octet-stream",
		mediaType == "application/zip",
		mediaType == "application/binary",
		mediaType == "application/json":
		return true
	default:
		return false
	}
}

func skillMDFromZipBundle(data []byte) ([]byte, error) {
	if len(data) < 4 || data[0] != 'P' || data[1] != 'K' {
		return nil, fmt.Errorf("missing zip signature")
	}
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}
	for _, file := range reader.File {
		if !strings.HasSuffix(file.Name, "SKILL.md") {
			continue
		}
		rc, err := file.Open()
		if err != nil {
			return nil, err
		}
		content, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			return nil, err
		}
		if len(content) == 0 {
			return nil, fmt.Errorf("SKILL.md is empty")
		}
		return content, nil
	}
	return nil, fmt.Errorf("SKILL.md not found in zip bundle")
}