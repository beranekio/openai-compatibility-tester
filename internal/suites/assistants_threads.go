package suites

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/packages/pagination"
	"github.com/openai/openai-go/v3/shared"
)

const (
	assistantThreadUserMessage        = "Reply with exactly the word: pong."
	assistantThreadCreateName         = "compatibility-test-assistant-thread"
	assistantThreadCreateInstructions = "Reply with exactly the word: pong."
	threadRunPollInterval             = 1 * time.Second
)

// AssistantsThreads verifies deprecated thread/message/run lifecycle via client.Beta.Threads.*.
type AssistantsThreads struct{}

func (AssistantsThreads) Name() string { return "assistants_threads" }
func (AssistantsThreads) Description() string {
	return "Deprecated Assistants thread/message/run lifecycle (POST /v1/threads, /v1/threads/{id}/messages, /v1/threads/{id}/runs)"
}
func (AssistantsThreads) Deprecated() bool { return true }

func (AssistantsThreads) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	assistantDeleted := false
	threadDeleted := false
	var assistantID string
	var threadID string
	defer func() {
		if threadID != "" && !threadDeleted {
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_, _ = client.Beta.Threads.Delete(cleanupCtx, threadID)
		}
		if assistantID != "" && !assistantDeleted {
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_, _ = client.Beta.Assistants.Delete(cleanupCtx, assistantID)
		}
	}()

	assistant, err := client.Beta.Assistants.New(ctx, openai.BetaAssistantNewParams{
		Model:        shared.ChatModel(cfg.Model),
		Name:         openai.String(assistantThreadCreateName),
		Instructions: openai.String(assistantThreadCreateInstructions),
	})
	if err != nil {
		return fmt.Errorf("assistant create failed: %w", err)
	}
	if err := validateAssistantObject("assistants_threads", assistant); err != nil {
		return err
	}
	assistantID = assistant.ID

	thread, err := client.Beta.Threads.New(ctx, openai.BetaThreadNewParams{
		Metadata: shared.Metadata{
			"suite": "assistants_threads",
		},
	})
	if err != nil {
		return fmt.Errorf("thread create failed: %w", err)
	}
	if err := validateThreadObject("assistants_threads", thread); err != nil {
		return err
	}
	threadID = thread.ID

	gotThread, err := client.Beta.Threads.Get(ctx, threadID)
	if err != nil {
		return fmt.Errorf("thread get failed: %w", err)
	}
	if err := validateThreadObject("assistants_threads", gotThread); err != nil {
		return err
	}
	if gotThread.ID != threadID {
		return fail("assistants_threads", fmt.Sprintf("get id is %q, want %q", gotThread.ID, threadID))
	}

	userMessage, err := client.Beta.Threads.Messages.New(ctx, threadID, openai.BetaThreadMessageNewParams{
		Role: openai.BetaThreadMessageNewParamsRoleUser,
		Content: openai.BetaThreadMessageNewParamsContentUnion{
			OfString: openai.String(assistantThreadUserMessage),
		},
	})
	if err != nil {
		return fmt.Errorf("thread message create failed: %w", err)
	}
	if err := validateThreadMessageObject("assistants_threads", userMessage); err != nil {
		return err
	}
	if userMessage.Role != openai.MessageRoleUser {
		return fail("assistants_threads", fmt.Sprintf("message role is %q, want user", userMessage.Role))
	}
	if !threadMessageContainsText(userMessage, assistantThreadUserMessage) {
		return fail("assistants_threads", "message create response missing submitted text")
	}

	run, err := client.Beta.Threads.Runs.New(ctx, threadID, openai.BetaThreadRunNewParams{
		AssistantID: assistantID,
	})
	if err != nil {
		return fmt.Errorf("thread run create failed: %w", err)
	}
	if err := validateThreadRunObject("assistants_threads", run); err != nil {
		return err
	}
	if run.AssistantID != assistantID {
		return fail("assistants_threads", fmt.Sprintf("run assistant_id is %q, want %q", run.AssistantID, assistantID))
	}
	if run.ThreadID != threadID {
		return fail("assistants_threads", fmt.Sprintf("run thread_id is %q, want %q", run.ThreadID, threadID))
	}

	gotRun, err := waitForThreadRunCompleted(ctx, client, "assistants_threads", threadID, run.ID)
	if err != nil {
		return err
	}
	if gotRun.AssistantID != assistantID {
		return fail("assistants_threads", fmt.Sprintf("run assistant_id is %q, want %q", gotRun.AssistantID, assistantID))
	}
	if gotRun.ThreadID != threadID {
		return fail("assistants_threads", fmt.Sprintf("run thread_id is %q, want %q", gotRun.ThreadID, threadID))
	}

	messagePage, err := client.Beta.Threads.Messages.List(ctx, threadID, openai.BetaThreadMessageListParams{
		Limit: openai.Int(20),
		Order: openai.BetaThreadMessageListParamsOrderAsc,
	})
	if err != nil {
		return fmt.Errorf("thread message list failed: %w", err)
	}
	if err := validateThreadMessagePage("assistants_threads", messagePage); err != nil {
		return err
	}
	if !threadMessagesContainText(messagePage.Data, assistantThreadUserMessage) {
		return fail("assistants_threads", "list response missing user message text")
	}
	if !threadMessagesHaveAssistantOutput(messagePage.Data) {
		return fail("assistants_threads", "list response missing assistant message text")
	}

	gotMessage, err := client.Beta.Threads.Messages.Get(ctx, threadID, userMessage.ID)
	if err != nil {
		return fmt.Errorf("thread message get failed: %w", err)
	}
	if err := validateThreadMessageObject("assistants_threads", gotMessage); err != nil {
		return err
	}
	if gotMessage.ID != userMessage.ID {
		return fail("assistants_threads", fmt.Sprintf("get message id is %q, want %q", gotMessage.ID, userMessage.ID))
	}

	deletedThread, err := client.Beta.Threads.Delete(ctx, threadID)
	if err != nil {
		return fmt.Errorf("thread delete failed: %w", err)
	}
	if err := validateThreadDeleted("assistants_threads", deletedThread); err != nil {
		return err
	}
	threadDeleted = true

	deletedAssistant, err := client.Beta.Assistants.Delete(ctx, assistantID)
	if err != nil {
		return fmt.Errorf("assistant delete failed: %w", err)
	}
	if err := validateAssistantDeleted("assistants_threads", deletedAssistant); err != nil {
		return err
	}
	assistantDeleted = true

	_, threadGetErr := client.Beta.Threads.Get(ctx, threadID)
	if threadGetErr == nil {
		return fail("assistants_threads", "thread get after delete succeeded; thread still exists")
	}
	var threadAPIError *openai.Error
	if !errors.As(threadGetErr, &threadAPIError) {
		return fmt.Errorf("thread get after delete failed: %w", threadGetErr)
	}
	if threadAPIError.StatusCode != http.StatusNotFound {
		return fail("assistants_threads", fmt.Sprintf("thread get after delete returned status %d, want 404", threadAPIError.StatusCode))
	}
	return nil
}

func validateThreadObject(suite string, thread *openai.Thread) error {
	if thread == nil {
		return fail(suite, "thread is nil")
	}
	if thread.ID == "" {
		return fail(suite, "thread missing id")
	}
	if !thread.JSON.CreatedAt.Valid() {
		return fail(suite, "thread missing created_at")
	}
	if !thread.JSON.Metadata.Valid() {
		return fail(suite, "thread missing metadata")
	}
	if !thread.JSON.Object.Valid() {
		return fail(suite, "thread missing object")
	}
	if string(thread.Object) != "thread" {
		return fail(suite, fmt.Sprintf("thread object is %q, want thread", thread.Object))
	}
	return nil
}

func validateThreadDeleted(suite string, deleted *openai.ThreadDeleted) error {
	if deleted == nil {
		return fail(suite, "delete response is nil")
	}
	if deleted.ID == "" {
		return fail(suite, "delete response missing id")
	}
	if !deleted.JSON.Deleted.Valid() {
		return fail(suite, "delete response missing deleted")
	}
	if !deleted.Deleted {
		return fail(suite, "delete response deleted is false")
	}
	if !deleted.JSON.Object.Valid() {
		return fail(suite, "delete response missing object")
	}
	if string(deleted.Object) != "thread.deleted" {
		return fail(suite, fmt.Sprintf("delete object is %q, want thread.deleted", deleted.Object))
	}
	return nil
}

func validateThreadMessageObject(suite string, message *openai.Message) error {
	if message == nil {
		return fail(suite, "message is nil")
	}
	if message.ID == "" {
		return fail(suite, "message missing id")
	}
	if !message.JSON.CreatedAt.Valid() {
		return fail(suite, "message missing created_at")
	}
	if !message.JSON.Content.Valid() {
		return fail(suite, "message missing content")
	}
	if len(message.Content) == 0 {
		return fail(suite, "message content is empty")
	}
	if !message.JSON.Object.Valid() {
		return fail(suite, "message missing object")
	}
	if string(message.Object) != "thread.message" {
		return fail(suite, fmt.Sprintf("message object is %q, want thread.message", message.Object))
	}
	if !message.JSON.Role.Valid() {
		return fail(suite, "message missing role")
	}
	if !message.JSON.Status.Valid() {
		return fail(suite, "message missing status")
	}
	if !message.JSON.ThreadID.Valid() {
		return fail(suite, "message missing thread_id")
	}
	return nil
}

func validateThreadMessagePage(suite string, page *pagination.CursorPage[openai.Message]) error {
	if page == nil {
		return fail(suite, "message page is nil")
	}
	if !page.JSON.HasMore.Valid() {
		return fail(suite, "message page missing has_more")
	}
	var envelope struct {
		Object string `json:"object"`
	}
	if err := json.Unmarshal([]byte(page.RawJSON()), &envelope); err != nil {
		return fail(suite, "message list response is not valid JSON")
	}
	if envelope.Object != "list" {
		return fail(suite, fmt.Sprintf("message list object is %q, want list", envelope.Object))
	}
	for i := range page.Data {
		if err := validateThreadMessageObject(suite, &page.Data[i]); err != nil {
			return err
		}
	}
	return nil
}

func validateThreadRunObject(suite string, run *openai.Run) error {
	if run == nil {
		return fail(suite, "run is nil")
	}
	if run.ID == "" {
		return fail(suite, "run missing id")
	}
	if !run.JSON.AssistantID.Valid() {
		return fail(suite, "run missing assistant_id")
	}
	if !run.JSON.CreatedAt.Valid() {
		return fail(suite, "run missing created_at")
	}
	if !run.JSON.Model.Valid() {
		return fail(suite, "run missing model")
	}
	if !run.JSON.Object.Valid() {
		return fail(suite, "run missing object")
	}
	if string(run.Object) != "thread.run" {
		return fail(suite, fmt.Sprintf("run object is %q, want thread.run", run.Object))
	}
	if !run.JSON.Status.Valid() {
		return fail(suite, "run missing status")
	}
	if !run.JSON.ThreadID.Valid() {
		return fail(suite, "run missing thread_id")
	}
	return nil
}

func waitForThreadRunCompleted(ctx context.Context, client openai.Client, suite, threadID, runID string) (*openai.Run, error) {
	for {
		gotRun, err := client.Beta.Threads.Runs.Get(ctx, threadID, runID)
		if err != nil {
			return nil, fmt.Errorf("thread run get failed: %w", err)
		}
		if err := validateThreadRunObject(suite, gotRun); err != nil {
			return nil, err
		}
		switch gotRun.Status {
		case openai.RunStatusCompleted:
			return gotRun, nil
		case openai.RunStatusFailed, openai.RunStatusCancelled, openai.RunStatusExpired, openai.RunStatusIncomplete:
			return nil, fail(suite, fmt.Sprintf("run failed with terminal status %q", gotRun.Status))
		}
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timed out waiting for run completion: %w", ctx.Err())
		case <-time.After(threadRunPollInterval):
		}
	}
}

func threadMessageContainsText(message *openai.Message, want string) bool {
	for _, content := range message.Content {
		text := content.AsText()
		if text.Text.Value == want {
			return true
		}
	}
	return false
}

func threadMessagesContainText(messages []openai.Message, want string) bool {
	for i := range messages {
		if threadMessageContainsText(&messages[i], want) {
			return true
		}
	}
	return false
}

func threadMessageHasTextOutput(message *openai.Message) bool {
	for _, content := range message.Content {
		text := content.AsText()
		if text.Text.Value != "" {
			return true
		}
	}
	return false
}

func threadMessagesHaveAssistantOutput(messages []openai.Message) bool {
	for i := range messages {
		if messages[i].Role != openai.MessageRoleAssistant {
			continue
		}
		if threadMessageHasTextOutput(&messages[i]) {
			return true
		}
	}
	return false
}