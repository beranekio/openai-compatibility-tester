package suites

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/pagination"
)

// FineTuning verifies the Fine-tuning API smoke flow via client.FineTuning.*.
type FineTuning struct{}

func (FineTuning) Name() string { return "fine_tuning" }
func (FineTuning) Description() string {
	return "Fine-tuning API smoke (POST/GET /v1/fine_tuning/jobs, checkpoints; permissions when OPENAI_ADMIN_API_KEY is set)"
}

func (FineTuning) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	var jobID string
	var fileID string
	defer func() {
		cleanupFineTuningArtifacts(client, jobID, fileID)
	}()

	uploaded, err := uploadFineTuneTrainingFile(ctx, client)
	if err != nil {
		return err
	}
	fileID = uploaded.ID

	created, err := client.FineTuning.Jobs.New(ctx, openai.FineTuningJobNewParams{
		Model:        openai.FineTuningJobNewParamsModel(cfg.Model),
		TrainingFile: uploaded.ID,
	})
	if err != nil {
		return fmt.Errorf("fine-tuning job create failed: %w", err)
	}
	if err := validateFineTuningJobEnvelope("fine_tuning", created); err != nil {
		if created != nil && created.ID != "" {
			jobID = created.ID
		}
		return err
	}
	jobID = created.ID
	if created.TrainingFile != uploaded.ID {
		return fail("fine_tuning", fmt.Sprintf("job training_file is %q, want %q", created.TrainingFile, uploaded.ID))
	}
	if !isFineTuningCreateStatusOK(string(created.Status)) {
		return fail("fine_tuning", fmt.Sprintf("job status is %q, want validating_files, queued, or running", created.Status))
	}

	listPage, err := client.FineTuning.Jobs.List(ctx, openai.FineTuningJobListParams{})
	if err != nil {
		return fmt.Errorf("fine-tuning job list failed: %w", err)
	}
	if err := validateFineTuningJobListPage("fine_tuning", listPage); err != nil {
		return err
	}
	found := false
	for _, item := range listPage.Data {
		if item.ID == jobID {
			found = true
			break
		}
	}
	if !found {
		return fail("fine_tuning", "created fine-tuning job missing from list response")
	}

	got, err := client.FineTuning.Jobs.Get(ctx, jobID)
	if err != nil {
		return fmt.Errorf("fine-tuning job get failed: %w", err)
	}
	if err := validateFineTuningJobEnvelope("fine_tuning", got); err != nil {
		return err
	}
	if got.ID != jobID {
		return fail("fine_tuning", fmt.Sprintf("get id is %q, want %q", got.ID, jobID))
	}

	checkpointPage, err := client.FineTuning.Jobs.Checkpoints.List(ctx, jobID, openai.FineTuningJobCheckpointListParams{})
	if err != nil {
		return fmt.Errorf("fine-tuning checkpoint list failed: %w", err)
	}
	if err := validateFineTuningCheckpointListPage("fine_tuning", checkpointPage); err != nil {
		return err
	}
	if len(checkpointPage.Data) > 0 && cfg.AdminAPIKey != "" {
		checkpoint := checkpointPage.Data[0]
		if err := validateFineTuningCheckpoint("fine_tuning", &checkpoint); err != nil {
			return err
		}
		if checkpoint.FineTuningJobID != jobID {
			return fail("fine_tuning", fmt.Sprintf("checkpoint job id is %q, want %q", checkpoint.FineTuningJobID, jobID))
		}

		permPage, err := client.FineTuning.Checkpoints.Permissions.List(
			ctx,
			checkpoint.FineTunedModelCheckpoint,
			openai.FineTuningCheckpointPermissionListParams{},
			option.WithAdminAPIKey(cfg.AdminAPIKey),
		)
		if err != nil {
			return fmt.Errorf("fine-tuning checkpoint permission list failed: %w", err)
		}
		if err := validateFineTuningCheckpointPermissionPage("fine_tuning", permPage); err != nil {
			return err
		}
	}

	cancelled, err := client.FineTuning.Jobs.Cancel(ctx, jobID)
	if err != nil {
		var apiErr *openai.Error
		if errors.As(err, &apiErr) && isFineTuningCancelAlreadyTerminalError(apiErr) {
			return nil
		}
		return fmt.Errorf("fine-tuning job cancel failed: %w", err)
	}
	if err := validateFineTuningJobEnvelope("fine_tuning", cancelled); err != nil {
		return err
	}
	if cancelled.ID != jobID {
		return fail("fine_tuning", fmt.Sprintf("cancel id is %q, want %q", cancelled.ID, jobID))
	}
	if !isFineTuningCancelStatusOK(string(cancelled.Status)) {
		return fail("fine_tuning", fmt.Sprintf("cancel status is %q, want cancelled", cancelled.Status))
	}
	return nil
}

func uploadFineTuneTrainingFile(ctx context.Context, client openai.Client) (*openai.FileObject, error) {
	uploaded, err := client.Files.New(ctx, openai.FileNewParams{
		File:    smallFineTuneJSONLReader(),
		Purpose: openai.FilePurposeFineTune,
	})
	if err != nil {
		return nil, fmt.Errorf("fine-tune training file upload failed: %w", err)
	}
	if err := validateFileObject("fine_tuning", uploaded); err != nil {
		if uploaded != nil && uploaded.ID != "" {
			deleteFineTuneTrainingFile(client, uploaded.ID)
		}
		return nil, err
	}
	if string(uploaded.Purpose) != string(openai.FilePurposeFineTune) {
		if uploaded.ID != "" {
			deleteFineTuneTrainingFile(client, uploaded.ID)
		}
		return nil, fail("fine_tuning", fmt.Sprintf("upload purpose is %q, want fine-tune", uploaded.Purpose))
	}
	return uploaded, nil
}

func deleteFineTuneTrainingFile(client openai.Client, fileID string) {
	cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, _ = client.Files.Delete(cleanupCtx, fileID)
}

func cleanupFineTuningArtifacts(client openai.Client, jobID, fileID string) {
	if jobID != "" {
		cancelCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_, _ = client.FineTuning.Jobs.Cancel(cancelCtx, jobID)
	}
	if fileID != "" {
		deleteFineTuneTrainingFile(client, fileID)
	}
}

func isFineTuningCreateStatusOK(status string) bool {
	return status == "validating_files" || status == "queued" || status == "running"
}

func isFineTuningCancelStatusOK(status string) bool {
	return status == "cancelled"
}

func isFineTuningCancelAlreadyTerminalError(apiErr *openai.Error) bool {
	if apiErr == nil {
		return false
	}
	switch apiErr.StatusCode {
	case http.StatusConflict, http.StatusBadRequest:
		if apiErr.Code == "invalid_job_status" {
			return true
		}
		detail := strings.ToLower(strings.Join([]string{apiErr.Code, apiErr.Message, apiErr.Type}, " "))
		statusIsSucceeded := detailContainsWord(detail, "succeed", "succeeded", "finished", "complete", "completed")
		statusIsCancelled := detailContainsWord(detail, "cancelled", "canceled")
		terminalSignal := detailContainsWord(detail, "already", "terminal") ||
			strings.Contains(detail, "cannot") || strings.Contains(detail, "can't") || strings.Contains(detail, "can not")
		return terminalSignal && (statusIsSucceeded || statusIsCancelled)
	default:
		return false
	}
}

func validateFineTuningJobEnvelope(suite string, job *openai.FineTuningJob) error {
	if job == nil {
		return fail(suite, "fine-tuning job is nil")
	}
	if job.ID == "" {
		return fail(suite, "fine-tuning job missing id")
	}
	if !job.JSON.CreatedAt.Valid() {
		return fail(suite, "fine-tuning job missing created_at")
	}
	if job.Model == "" {
		return fail(suite, "fine-tuning job missing model")
	}
	if !job.JSON.Object.Valid() {
		return fail(suite, "fine-tuning job missing object")
	}
	if string(job.Object) != "fine_tuning.job" {
		return fail(suite, fmt.Sprintf("fine-tuning job object is %q, want fine_tuning.job", job.Object))
	}
	if !job.JSON.Status.Valid() {
		return fail(suite, "fine-tuning job missing status")
	}
	if !job.JSON.TrainingFile.Valid() {
		return fail(suite, "fine-tuning job missing training_file")
	}
	if job.TrainingFile == "" {
		return fail(suite, "fine-tuning job training_file is empty")
	}
	return nil
}

func validateFineTuningJobListPage(suite string, page *pagination.CursorPage[openai.FineTuningJob]) error {
	if page == nil {
		return fail(suite, "fine-tuning job list page is nil")
	}
	if !page.JSON.HasMore.Valid() {
		return fail(suite, "fine-tuning job list missing has_more")
	}
	var envelope struct {
		Object string `json:"object"`
	}
	if err := json.Unmarshal([]byte(page.RawJSON()), &envelope); err != nil {
		return fail(suite, "fine-tuning job list response is not valid JSON")
	}
	if envelope.Object != "list" {
		return fail(suite, fmt.Sprintf("fine-tuning job list object is %q, want list", envelope.Object))
	}
	for i := range page.Data {
		if err := validateFineTuningJobEnvelope(suite, &page.Data[i]); err != nil {
			return err
		}
	}
	return nil
}

func validateFineTuningCheckpointListPage(suite string, page *pagination.CursorPage[openai.FineTuningJobCheckpoint]) error {
	if page == nil {
		return fail(suite, "fine-tuning checkpoint list page is nil")
	}
	if !page.JSON.HasMore.Valid() {
		return fail(suite, "fine-tuning checkpoint list missing has_more")
	}
	if !page.JSON.Data.Valid() {
		return fail(suite, "fine-tuning checkpoint list missing data")
	}
	var envelope struct {
		Object string `json:"object"`
	}
	if err := json.Unmarshal([]byte(page.RawJSON()), &envelope); err != nil {
		return fail(suite, "fine-tuning checkpoint list response is not valid JSON")
	}
	if envelope.Object != "list" {
		return fail(suite, fmt.Sprintf("fine-tuning checkpoint list object is %q, want list", envelope.Object))
	}
	for i := range page.Data {
		if err := validateFineTuningCheckpoint(suite, &page.Data[i]); err != nil {
			return err
		}
	}
	return nil
}

func validateFineTuningCheckpoint(suite string, checkpoint *openai.FineTuningJobCheckpoint) error {
	if checkpoint == nil {
		return fail(suite, "fine-tuning checkpoint is nil")
	}
	if checkpoint.ID == "" {
		return fail(suite, "fine-tuning checkpoint missing id")
	}
	if !checkpoint.JSON.CreatedAt.Valid() {
		return fail(suite, "fine-tuning checkpoint missing created_at")
	}
	if checkpoint.FineTunedModelCheckpoint == "" {
		return fail(suite, "fine-tuning checkpoint missing fine_tuned_model_checkpoint")
	}
	if checkpoint.FineTuningJobID == "" {
		return fail(suite, "fine-tuning checkpoint missing fine_tuning_job_id")
	}
	if !checkpoint.JSON.Object.Valid() {
		return fail(suite, "fine-tuning checkpoint missing object")
	}
	if string(checkpoint.Object) != "fine_tuning.job.checkpoint" {
		return fail(suite, fmt.Sprintf("fine-tuning checkpoint object is %q, want fine_tuning.job.checkpoint", checkpoint.Object))
	}
	if !checkpoint.JSON.StepNumber.Valid() {
		return fail(suite, "fine-tuning checkpoint missing step_number")
	}
	if !checkpoint.JSON.Metrics.Valid() {
		return fail(suite, "fine-tuning checkpoint missing metrics")
	}
	return nil
}

func validateFineTuningCheckpointPermissionPage(suite string, page *pagination.ConversationCursorPage[openai.FineTuningCheckpointPermissionListResponse]) error {
	if page == nil {
		return fail(suite, "fine-tuning checkpoint permission page is nil")
	}
	if !page.JSON.HasMore.Valid() {
		return fail(suite, "fine-tuning checkpoint permission list missing has_more")
	}
	if !page.JSON.Data.Valid() {
		return fail(suite, "fine-tuning checkpoint permission list missing data")
	}
	var envelope struct {
		Object string `json:"object"`
	}
	if err := json.Unmarshal([]byte(page.RawJSON()), &envelope); err != nil {
		return fail(suite, "fine-tuning checkpoint permission list response is not valid JSON")
	}
	if envelope.Object != "list" {
		return fail(suite, fmt.Sprintf("fine-tuning checkpoint permission list object is %q, want list", envelope.Object))
	}
	for i := range page.Data {
		item := page.Data[i]
		if item.ID == "" {
			return fail(suite, "fine-tuning checkpoint permission missing id")
		}
		if !item.JSON.CreatedAt.Valid() {
			return fail(suite, "fine-tuning checkpoint permission missing created_at")
		}
		if !item.JSON.Object.Valid() {
			return fail(suite, "fine-tuning checkpoint permission missing object")
		}
		if string(item.Object) != "checkpoint.permission" {
			return fail(suite, fmt.Sprintf("fine-tuning checkpoint permission object is %q, want checkpoint.permission", item.Object))
		}
		if !item.JSON.ProjectID.Valid() {
			return fail(suite, "fine-tuning checkpoint permission missing project_id")
		}
		if item.ProjectID == "" {
			return fail(suite, "fine-tuning checkpoint permission project_id is empty")
		}
	}
	return nil
}