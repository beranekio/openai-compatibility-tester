package suites

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/beranekio/openai-compatibility-tester/internal/config"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/packages/pagination"
)

const chatkitThreadUser = "compatibility-test-user"

// ChatKitThreads verifies Beta ChatKit thread lifecycle via client.Beta.ChatKit.Threads.*.
type ChatKitThreads struct{}

func (ChatKitThreads) Name() string { return "chatkit_threads" }
func (ChatKitThreads) Description() string {
	return "Beta ChatKit threads (GET /v1/chatkit/threads; get/items when a thread is listed or OPENAI_CHATKIT_TEST_THREAD_ID is set; DELETE only with OPENAI_CHATKIT_TEST_THREAD_ID)"
}

func (ChatKitThreads) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	listPage, err := client.Beta.ChatKit.Threads.List(ctx, openai.BetaChatKitThreadListParams{
		Limit: openai.Int(10),
		User:  openai.String(chatkitThreadUser),
		Order: openai.BetaChatKitThreadListParamsOrderDesc,
	})
	if err != nil {
		return fmt.Errorf("chatkit thread list failed: %w", err)
	}
	if err := validateChatKitThreadPage("chatkit_threads", listPage); err != nil {
		return err
	}
	if cfg.ChatKitTestThreadID != "" {
		return runChatKitThreadsWithDelete(ctx, client, cfg.ChatKitTestThreadID)
	}
	if len(listPage.Data) == 0 {
		return nil
	}

	return runChatKitThreadsReadOnly(ctx, client, listPage.Data[0].ID)
}

func runChatKitThreadsReadOnly(ctx context.Context, client openai.Client, threadID string) error {
	got, err := client.Beta.ChatKit.Threads.Get(ctx, threadID)
	if err != nil {
		return fmt.Errorf("chatkit thread get failed: %w", err)
	}
	if err := validateChatKitThread("chatkit_threads", got); err != nil {
		return err
	}
	if got.ID != threadID {
		return fail("chatkit_threads", fmt.Sprintf("get id is %q, want %q", got.ID, threadID))
	}

	itemPage, err := client.Beta.ChatKit.Threads.ListItems(ctx, threadID, openai.BetaChatKitThreadListItemsParams{
		Limit: openai.Int(10),
		Order: openai.BetaChatKitThreadListItemsParamsOrderAsc,
	})
	if err != nil {
		return fmt.Errorf("chatkit thread item list failed: %w", err)
	}
	return validateChatKitThreadItemPage("chatkit_threads", threadID, itemPage)
}

func runChatKitThreadsWithDelete(ctx context.Context, client openai.Client, threadID string) error {
	if err := runChatKitThreadsReadOnly(ctx, client, threadID); err != nil {
		return err
	}

	deleted, err := client.Beta.ChatKit.Threads.Delete(ctx, threadID)
	if err != nil {
		return fmt.Errorf("chatkit thread delete failed: %w", err)
	}
	if err := validateChatKitThreadDeleted("chatkit_threads", deleted); err != nil {
		return err
	}
	if deleted.ID != threadID {
		return fail("chatkit_threads", fmt.Sprintf("delete id is %q, want %q", deleted.ID, threadID))
	}

	_, getErr := client.Beta.ChatKit.Threads.Get(ctx, threadID)
	if getErr == nil {
		return fail("chatkit_threads", "thread get after delete succeeded; thread still exists")
	}
	var apiError *openai.Error
	if !errors.As(getErr, &apiError) {
		return fmt.Errorf("thread get after delete failed: %w", getErr)
	}
	if apiError.StatusCode != http.StatusNotFound {
		return fail("chatkit_threads", fmt.Sprintf("thread get after delete returned status %d, want 404", apiError.StatusCode))
	}
	return nil
}

func validateChatKitThread(suite string, thread *openai.ChatKitThread) error {
	if thread == nil {
		return fail(suite, "thread is nil")
	}
	if thread.ID == "" {
		return fail(suite, "thread missing id")
	}
	if !thread.JSON.CreatedAt.Valid() {
		return fail(suite, "thread missing created_at")
	}
	if !thread.JSON.Object.Valid() {
		return fail(suite, "thread missing object")
	}
	if string(thread.Object) != "chatkit.thread" {
		return fail(suite, fmt.Sprintf("thread object is %q, want chatkit.thread", thread.Object))
	}
	if !thread.JSON.Status.Valid() {
		return fail(suite, "thread missing status")
	}
	if !isValidChatKitThreadStatus(thread.Status.Type) {
		return fail(suite, fmt.Sprintf("thread status type is %q, want active, locked, or closed", thread.Status.Type))
	}
	if !thread.JSON.User.Valid() {
		return fail(suite, "thread missing user")
	}
	if thread.User == "" {
		return fail(suite, "thread user is empty")
	}
	return nil
}

func validateChatKitListEnvelope(suite, listContext, pageRaw, firstDataID, lastDataID string) error {
	var envelope struct {
		Object  string `json:"object"`
		FirstID string `json:"first_id"`
		LastID  string `json:"last_id"`
	}
	if err := json.Unmarshal([]byte(pageRaw), &envelope); err != nil {
		return fail(suite, listContext+" response is not valid JSON")
	}
	if envelope.Object != "list" {
		return fail(suite, fmt.Sprintf("%s object is %q, want list", listContext, envelope.Object))
	}
	if firstDataID == "" {
		return nil
	}
	if envelope.FirstID == "" {
		return fail(suite, listContext+" missing first_id")
	}
	if envelope.LastID == "" {
		return fail(suite, listContext+" missing last_id")
	}
	if envelope.FirstID != firstDataID {
		return fail(suite, fmt.Sprintf("%s first_id is %q, want %q", listContext, envelope.FirstID, firstDataID))
	}
	if envelope.LastID != lastDataID {
		return fail(suite, fmt.Sprintf("%s last_id is %q, want %q", listContext, envelope.LastID, lastDataID))
	}
	return nil
}

func validateCreatedAtOrder(suite, listContext string, timestamps []int64, descending bool) error {
	if len(timestamps) < 2 {
		return nil
	}
	for i := 1; i < len(timestamps); i++ {
		prev := timestamps[i-1]
		curr := timestamps[i]
		if descending {
			if curr > prev {
				return fail(suite, fmt.Sprintf("%s created_at is not in descending order", listContext))
			}
			continue
		}
		if curr < prev {
			return fail(suite, fmt.Sprintf("%s created_at is not in ascending order", listContext))
		}
	}
	return nil
}

func isValidChatKitThreadItemType(itemType string) bool {
	switch itemType {
	case "chatkit.user_message", "chatkit.assistant_message", "chatkit.widget",
		"chatkit.client_tool_call", "chatkit.task", "chatkit.task_group":
		return true
	default:
		return false
	}
}

func isValidChatKitThreadStatus(statusType string) bool {
	switch statusType {
	case "active", "locked", "closed":
		return true
	default:
		return false
	}
}

func validateChatKitThreadPage(suite string, page *pagination.ConversationCursorPage[openai.ChatKitThread]) error {
	if page == nil {
		return fail(suite, "thread page is nil")
	}
	if !page.JSON.Data.Valid() {
		return fail(suite, "thread page missing data")
	}
	if !page.JSON.HasMore.Valid() {
		return fail(suite, "thread page missing has_more")
	}
	firstID := ""
	lastID := ""
	if len(page.Data) > 0 {
		firstID = page.Data[0].ID
		lastID = page.Data[len(page.Data)-1].ID
	}
	if err := validateChatKitListEnvelope(suite, "thread list", page.RawJSON(), firstID, lastID); err != nil {
		return err
	}
	createdAts := make([]int64, len(page.Data))
	for i := range page.Data {
		if err := validateChatKitThread(suite, &page.Data[i]); err != nil {
			return err
		}
		if page.Data[i].User != chatkitThreadUser {
			return fail(suite, fmt.Sprintf("thread user is %q, want %q", page.Data[i].User, chatkitThreadUser))
		}
		createdAts[i] = page.Data[i].CreatedAt
	}
	return validateCreatedAtOrder(suite, "thread list", createdAts, true)
}

func validateChatKitThreadItemPage(suite string, threadID string, page *pagination.ConversationCursorPage[openai.ChatKitThreadItemListDataUnion]) error {
	if page == nil {
		return fail(suite, "thread item page is nil")
	}
	if !page.JSON.Data.Valid() {
		return fail(suite, "thread item page missing data")
	}
	if !page.JSON.HasMore.Valid() {
		return fail(suite, "thread item page missing has_more")
	}
	firstID := ""
	lastID := ""
	if len(page.Data) > 0 {
		firstID = page.Data[0].ID
		lastID = page.Data[len(page.Data)-1].ID
	}
	if err := validateChatKitListEnvelope(suite, "thread item list", page.RawJSON(), firstID, lastID); err != nil {
		return err
	}
	createdAts := make([]int64, len(page.Data))
	for i := range page.Data {
		if err := validateChatKitThreadItem(suite, threadID, &page.Data[i]); err != nil {
			return err
		}
		createdAts[i] = page.Data[i].CreatedAt
	}
	if err := validateCreatedAtOrder(suite, "thread item list", createdAts, false); err != nil {
		return err
	}
	return nil
}

func validateChatKitThreadItem(suite string, threadID string, item *openai.ChatKitThreadItemListDataUnion) error {
	if item == nil {
		return fail(suite, "thread item is nil")
	}
	if item.ID == "" {
		return fail(suite, "thread item missing id")
	}
	if item.Type == "" {
		return fail(suite, "thread item missing type")
	}
	if !isValidChatKitThreadItemType(item.Type) {
		return fail(suite, fmt.Sprintf("thread item type is %q, want a known chatkit item discriminator", item.Type))
	}
	if item.AsAny() == nil {
		return fail(suite, fmt.Sprintf("thread item type %q did not parse into a known variant", item.Type))
	}
	if !item.JSON.Object.Valid() {
		return fail(suite, "thread item missing object")
	}
	if string(item.Object) != "chatkit.thread_item" {
		return fail(suite, fmt.Sprintf("thread item object is %q, want chatkit.thread_item", item.Object))
	}
	if !item.JSON.CreatedAt.Valid() {
		return fail(suite, "thread item missing created_at")
	}
	if item.ThreadID == "" {
		return fail(suite, "thread item missing thread_id")
	}
	if item.ThreadID != threadID {
		return fail(suite, fmt.Sprintf("thread item thread_id is %q, want %q", item.ThreadID, threadID))
	}
	return validateChatKitThreadItemVariant(suite, item)
}

func validateChatKitThreadItemVariant(suite string, item *openai.ChatKitThreadItemListDataUnion) error {
	switch item.Type {
	case "chatkit.user_message":
		msg := item.AsChatKitUserMessage()
		if !msg.JSON.Content.Valid() {
			return fail(suite, "user_message item missing content")
		}
		if len(msg.Content) == 0 {
			return fail(suite, "user_message item content is empty")
		}
	case "chatkit.assistant_message":
		msg := item.AsChatKitAssistantMessage()
		if !msg.JSON.Content.Valid() {
			return fail(suite, "assistant_message item missing content")
		}
		if len(msg.Content) == 0 {
			return fail(suite, "assistant_message item content is empty")
		}
	}
	return nil
}

func validateChatKitThreadDeleted(suite string, deleted *openai.BetaChatKitThreadDeleteResponse) error {
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
	if string(deleted.Object) != "chatkit.thread.deleted" {
		return fail(suite, fmt.Sprintf("delete object is %q, want chatkit.thread.deleted", deleted.Object))
	}
	return nil
}
