package suites

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/packages/pagination"
)

const skillVersionUpdatedContent = "# compatibility test skill v2\n"

// SkillVersions verifies skill version lifecycle and content retrieval via
// client.Skills.Versions.* and client.Skills.Content.*.
type SkillVersions struct{}

func (SkillVersions) Name() string { return "skill_versions" }
func (SkillVersions) Description() string {
	return "Skill versions lifecycle and content retrieval (POST/GET/DELETE /v1/skills/{id}/versions, GET /v1/skills/{id}/content)"
}

func (SkillVersions) Run(ctx context.Context, client openai.Client, _ *config.Config) error {
	deleted := false
	var skillID string
	defer func() {
		if skillID != "" && !deleted {
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_, _ = client.Skills.Delete(cleanupCtx, skillID)
		}
	}()

	created, err := createSkillForSuite(ctx, client, "skill_versions")
	if err != nil {
		return err
	}
	skillID = created.ID
	initialVersion := created.DefaultVersion

	versionCreated, err := client.Skills.Versions.New(ctx, skillID, openai.SkillVersionNewParams{
		Files: openai.SkillVersionNewParamsFilesUnion{
			OfFileArray: []io.Reader{skillVersionFileReader()},
		},
	})
	if err != nil {
		return fmt.Errorf("skill version create failed: %w", err)
	}
	if err := validateSkillVersionObject("skill_versions", versionCreated); err != nil {
		return err
	}
	if versionCreated.SkillID != skillID {
		return fail("skill_versions", fmt.Sprintf("create skill_id is %q, want %q", versionCreated.SkillID, skillID))
	}
	if versionCreated.Version == initialVersion {
		return fail("skill_versions", fmt.Sprintf("create version is %q, want a new version after %q", versionCreated.Version, initialVersion))
	}

	got, err := client.Skills.Versions.Get(ctx, skillID, versionCreated.Version)
	if err != nil {
		return fmt.Errorf("skill version get failed: %w", err)
	}
	if err := validateSkillVersionObject("skill_versions", got); err != nil {
		return err
	}
	if got.ID != versionCreated.ID {
		return fail("skill_versions", fmt.Sprintf("get id is %q, want %q", got.ID, versionCreated.ID))
	}
	if got.Version != versionCreated.Version {
		return fail("skill_versions", fmt.Sprintf("get version is %q, want %q", got.Version, versionCreated.Version))
	}

	listPage, err := client.Skills.Versions.List(ctx, skillID, openai.SkillVersionListParams{
		Limit: openai.Int(100),
		Order: openai.SkillVersionListParamsOrderDesc,
	})
	if err != nil {
		return fmt.Errorf("skill version list failed: %w", err)
	}
	found, err := skillVersionListContains(listPage, versionCreated.Version)
	if err != nil {
		return err
	}
	if !found {
		return fail("skill_versions", "created skill version missing from list response")
	}

	skillContentResp, err := client.Skills.Content.Get(ctx, skillID)
	if err != nil {
		return fmt.Errorf("skill content get failed: %w", err)
	}
	if err := validateSkillContentResponse("skill_versions", skillContentResp, smallSkillFileBytes()); err != nil {
		return err
	}

	versionContentResp, err := client.Skills.Versions.Content.Get(ctx, skillID, versionCreated.Version)
	if err != nil {
		return fmt.Errorf("skill version content get failed: %w", err)
	}
	if err := validateSkillContentResponse("skill_versions", versionContentResp, []byte(skillVersionUpdatedContent)); err != nil {
		return err
	}

	deletedVersion, err := client.Skills.Versions.Delete(ctx, skillID, versionCreated.Version)
	if err != nil {
		return fmt.Errorf("skill version delete failed: %w", err)
	}
	if err := validateSkillVersionDeleteResponse("skill_versions", deletedVersion, versionCreated.ID, versionCreated.Version); err != nil {
		return err
	}

	_, getErr := client.Skills.Versions.Get(ctx, skillID, versionCreated.Version)
	if getErr == nil {
		return fail("skill_versions", "get after version delete succeeded; version still exists")
	}
	var apiErr *openai.Error
	if !errors.As(getErr, &apiErr) {
		return fmt.Errorf("get after version delete failed: %w", getErr)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		return fail("skill_versions", fmt.Sprintf("get after version delete returned status %d, want 404", apiErr.StatusCode))
	}

	if _, err := client.Skills.Delete(ctx, skillID); err != nil {
		return fmt.Errorf("skill delete failed: %w", err)
	}
	deleted = true
	return nil
}

func skillVersionListContains(page *pagination.CursorPage[openai.SkillVersion], version string) (bool, error) {
	for page != nil {
		if err := validateSkillVersionListPage("skill_versions", page); err != nil {
			return false, err
		}
		for _, item := range page.Data {
			if item.Version == version {
				return true, nil
			}
		}
		next, err := page.GetNextPage()
		if err != nil {
			return false, fmt.Errorf("skill version list next page failed: %w", err)
		}
		page = next
	}
	return false, nil
}

func validateSkillVersionObject(suite string, version *openai.SkillVersion) error {
	if version == nil {
		return fail(suite, "skill version is nil")
	}
	if version.ID == "" {
		return fail(suite, "skill version missing id")
	}
	if !version.JSON.CreatedAt.Valid() {
		return fail(suite, "skill version missing created_at")
	}
	if !version.JSON.Description.Valid() {
		return fail(suite, "skill version missing description")
	}
	if !version.JSON.Name.Valid() {
		return fail(suite, "skill version missing name")
	}
	if !version.JSON.Object.Valid() {
		return fail(suite, "skill version missing object")
	}
	if string(version.Object) != "skill.version" {
		return fail(suite, fmt.Sprintf("skill version object is %q, want skill.version", version.Object))
	}
	if !version.JSON.SkillID.Valid() {
		return fail(suite, "skill version missing skill_id")
	}
	if !version.JSON.Version.Valid() {
		return fail(suite, "skill version missing version")
	}
	return nil
}

func validateSkillVersionListPage(suite string, page *pagination.CursorPage[openai.SkillVersion]) error {
	if page == nil {
		return fail(suite, "list page is nil")
	}
	if !page.JSON.HasMore.Valid() {
		return fail(suite, "list missing has_more")
	}
	var envelope struct {
		Object string `json:"object"`
	}
	if err := json.Unmarshal([]byte(page.RawJSON()), &envelope); err != nil {
		return fail(suite, "list response is not valid JSON")
	}
	if envelope.Object != "list" {
		return fail(suite, fmt.Sprintf("list object is %q, want list", envelope.Object))
	}
	for i := range page.Data {
		if err := validateSkillVersionObject(suite, &page.Data[i]); err != nil {
			return err
		}
	}
	return nil
}

func validateSkillVersionDeleteResponse(suite string, deleted *openai.DeletedSkillVersion, wantID, wantVersion string) error {
	if deleted == nil {
		return fail(suite, "delete response is nil")
	}
	if deleted.ID != wantID {
		return fail(suite, fmt.Sprintf("delete id is %q, want %q", deleted.ID, wantID))
	}
	if !deleted.Deleted {
		return fail(suite, "delete response deleted is false")
	}
	if !deleted.JSON.Object.Valid() {
		return fail(suite, "delete response missing object")
	}
	if string(deleted.Object) != "skill.version.deleted" {
		return fail(suite, fmt.Sprintf("delete object is %q, want skill.version.deleted", deleted.Object))
	}
	if deleted.Version != wantVersion {
		return fail(suite, fmt.Sprintf("delete version is %q, want %q", deleted.Version, wantVersion))
	}
	return nil
}

func validateSkillContentResponse(suite string, resp *http.Response, want []byte) error {
	if resp == nil {
		return fail(suite, "response is nil")
	}
	if resp.Body == nil {
		return fail(suite, "response body is nil")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("%s: read response body: %w", suite, err)
	}
	if len(body) < len(want) {
		return fail(suite, fmt.Sprintf("response body has %d bytes, want at least %d", len(body), len(want)))
	}
	if !bytes.Equal(body, want) {
		return fail(suite, "content body does not match uploaded skill files")
	}

	contentType := resp.Header.Get("Content-Type")
	if strings.TrimSpace(contentType) == "" {
		return fail(suite, "response missing Content-Type")
	}
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return fail(suite, fmt.Sprintf("Content-Type %q is invalid: %v", contentType, err))
	}
	if !strings.HasPrefix(mediaType, "audio/") &&
		mediaType != "application/octet-stream" &&
		mediaType != "application/zip" &&
		mediaType != "application/binary" {
		return fail(suite, fmt.Sprintf("Content-Type is %q, want audio/*, application/octet-stream, application/zip, or application/binary", mediaType))
	}
	return nil
}

type namedSkillVersionFileReader struct {
	r *bytes.Reader
}

func (r *namedSkillVersionFileReader) Read(p []byte) (int, error) {
	return r.r.Read(p)
}

func (r *namedSkillVersionFileReader) Filename() string {
	return "SKILL.md"
}

func (r *namedSkillVersionFileReader) ContentType() string {
	return "text/markdown"
}

func skillVersionFileReader() io.Reader {
	return &namedSkillVersionFileReader{r: bytes.NewReader([]byte(skillVersionUpdatedContent))}
}