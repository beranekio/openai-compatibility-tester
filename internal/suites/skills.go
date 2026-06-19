package suites

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/beranekio/openai-compatibility-tester/internal/config"
	"github.com/beranekio/openai-compatibility-tester/internal/testutil"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/packages/pagination"
)

const skillCreateName = "compatibility-test-skill"

// Skills verifies the Skills API lifecycle via client.Skills.*.
type Skills struct{}

func (Skills) Name() string { return "skills" }
func (Skills) Description() string {
	return "Skills API lifecycle (POST /v1/skills, GET /v1/skills, GET/POST/DELETE /v1/skills/{id}, POST /v1/skills/{id}/versions)"
}

func (Skills) Run(ctx context.Context, client openai.Client, _ *config.Config) error {
	deleted := false
	var skillID string
	defer func() {
		if skillID != "" && !deleted {
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_, _ = client.Skills.Delete(cleanupCtx, skillID)
		}
	}()

	created, err := client.Skills.New(ctx, openai.SkillNewParams{
		Files: openai.SkillNewParamsFilesUnion{
			OfFileArray: []io.Reader{testutil.SmallSkillFileReader()},
		},
	})
	if err != nil {
		return fmt.Errorf("skill create failed: %w", err)
	}
	if err := validateSkillObject("skills", created); err != nil {
		return err
	}
	skillID = created.ID
	if created.Name != skillCreateName {
		return fail("skills", fmt.Sprintf("create name is %q, want %q", created.Name, skillCreateName))
	}
	if created.DefaultVersion == "" {
		return fail("skills", "create missing default_version")
	}
	if created.LatestVersion == "" {
		return fail("skills", "create missing latest_version")
	}

	got, err := client.Skills.Get(ctx, skillID)
	if err != nil {
		return fmt.Errorf("skill get failed: %w", err)
	}
	if err := validateSkillObject("skills", got); err != nil {
		return err
	}
	if got.ID != skillID {
		return fail("skills", fmt.Sprintf("get id is %q, want %q", got.ID, skillID))
	}
	if got.Name != skillCreateName {
		return fail("skills", fmt.Sprintf("get name is %q, want %q", got.Name, skillCreateName))
	}

	listPage, err := client.Skills.List(ctx, openai.SkillListParams{
		Limit: openai.Int(100),
		Order: openai.SkillListParamsOrderDesc,
	})
	if err != nil {
		return fmt.Errorf("skill list failed: %w", err)
	}
	found, err := skillListContains(listPage, skillID)
	if err != nil {
		return err
	}
	if !found {
		return fail("skills", "created skill missing from list response")
	}

	newVersion, err := client.Skills.Versions.New(ctx, skillID, openai.SkillVersionNewParams{
		Files: openai.SkillVersionNewParamsFilesUnion{
			OfFileArray: []io.Reader{testutil.SkillVersionFileReader()},
		},
	})
	if err != nil {
		return fmt.Errorf("skill version create failed: %w", err)
	}
	if err := validateSkillVersionObject("skills", newVersion); err != nil {
		return err
	}
	if newVersion.Version == created.DefaultVersion {
		return fail("skills", fmt.Sprintf("new version is %q, want a version after %q", newVersion.Version, created.DefaultVersion))
	}

	updated, err := client.Skills.Update(ctx, skillID, openai.SkillUpdateParams{
		DefaultVersion: newVersion.Version,
	})
	if err != nil {
		return fmt.Errorf("skill update failed: %w", err)
	}
	if err := validateSkillObject("skills", updated); err != nil {
		return err
	}
	if updated.ID != skillID {
		return fail("skills", fmt.Sprintf("update id is %q, want %q", updated.ID, skillID))
	}
	if updated.DefaultVersion != newVersion.Version {
		return fail("skills", fmt.Sprintf("update default_version is %q, want %q", updated.DefaultVersion, newVersion.Version))
	}

	deletedResp, err := client.Skills.Delete(ctx, skillID)
	if err != nil {
		return fmt.Errorf("skill delete failed: %w", err)
	}
	if err := validateSkillDeleteResponse("skills", deletedResp, skillID); err != nil {
		return err
	}

	_, getErr := client.Skills.Get(ctx, skillID)
	if getErr == nil {
		return fail("skills", "get after delete succeeded; skill still exists")
	}
	var apiErr *openai.Error
	if !errors.As(getErr, &apiErr) {
		return fmt.Errorf("get after delete failed: %w", getErr)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		return fail("skills", fmt.Sprintf("get after delete returned status %d, want 404", apiErr.StatusCode))
	}
	deleted = true
	return nil
}

func skillListContains(page *pagination.CursorPage[openai.Skill], skillID string) (bool, error) {
	for page != nil {
		if err := validateSkillListPage("skills", page); err != nil {
			return false, err
		}
		for _, item := range page.Data {
			if item.ID == skillID {
				return true, nil
			}
		}
		next, err := page.GetNextPage()
		if err != nil {
			return false, fmt.Errorf("skill list next page failed: %w", err)
		}
		page = next
	}
	return false, nil
}

func validateSkillObject(suite string, skill *openai.Skill) error {
	if skill == nil {
		return fail(suite, "skill is nil")
	}
	if skill.ID == "" {
		return fail(suite, "skill missing id")
	}
	if !skill.JSON.CreatedAt.Valid() {
		return fail(suite, "skill missing created_at")
	}
	if !skill.JSON.DefaultVersion.Valid() {
		return fail(suite, "skill missing default_version")
	}
	if !skill.JSON.Description.Valid() {
		return fail(suite, "skill missing description")
	}
	if !skill.JSON.LatestVersion.Valid() {
		return fail(suite, "skill missing latest_version")
	}
	if !skill.JSON.Name.Valid() {
		return fail(suite, "skill missing name")
	}
	if !skill.JSON.Object.Valid() {
		return fail(suite, "skill missing object")
	}
	if string(skill.Object) != "skill" {
		return fail(suite, fmt.Sprintf("skill object is %q, want skill", skill.Object))
	}
	return nil
}

func validateSkillListPage(suite string, page *pagination.CursorPage[openai.Skill]) error {
	if page == nil {
		return fail(suite, "list page is nil")
	}
	if !page.JSON.HasMore.Valid() {
		return fail(suite, "list missing has_more")
	}
	var envelope struct {
		Object  string `json:"object"`
		FirstID string `json:"first_id"`
		LastID  string `json:"last_id"`
	}
	if err := json.Unmarshal([]byte(page.RawJSON()), &envelope); err != nil {
		return fail(suite, "list response is not valid JSON")
	}
	if envelope.Object != "list" {
		return fail(suite, fmt.Sprintf("list object is %q, want list", envelope.Object))
	}
	if len(page.Data) == 0 {
		return nil
	}
	if envelope.FirstID == "" {
		return fail(suite, "list missing first_id")
	}
	if envelope.LastID == "" {
		return fail(suite, "list missing last_id")
	}
	if envelope.FirstID != page.Data[0].ID {
		return fail(suite, fmt.Sprintf("list first_id is %q, want %q", envelope.FirstID, page.Data[0].ID))
	}
	if envelope.LastID != page.Data[len(page.Data)-1].ID {
		return fail(suite, fmt.Sprintf("list last_id is %q, want %q", envelope.LastID, page.Data[len(page.Data)-1].ID))
	}
	for i := range page.Data {
		if err := validateSkillObject(suite, &page.Data[i]); err != nil {
			return err
		}
	}
	return nil
}

func validateSkillDeleteResponse(suite string, deleted *openai.DeletedSkill, wantID string) error {
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
	if string(deleted.Object) != "skill.deleted" {
		return fail(suite, fmt.Sprintf("delete object is %q, want skill.deleted", deleted.Object))
	}
	return nil
}

func createSkillForSuite(ctx context.Context, client openai.Client, suite string) (*openai.Skill, error) {
	created, err := client.Skills.New(ctx, openai.SkillNewParams{
		Files: openai.SkillNewParamsFilesUnion{
			OfFileArray: []io.Reader{testutil.SmallSkillFileReader()},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("%s: skill create failed: %w", suite, err)
	}
	if err := validateSkillObject(suite, created); err != nil {
		return nil, err
	}
	return created, nil
}
