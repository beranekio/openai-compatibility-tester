package suites

import (
	"context"
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
	return "Beta ChatKit threads (GET /v1/chatkit/threads, GET /v1/chatkit/threads/{id}, GET /v1/chatkit/threads/{id}/items[, DELETE when OPENAI_CHATKIT_TEST_THREAD_ID is set])"
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
	return validateChatKitThreadItemPage("chatkit_threads", itemPage)
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
	for i := range page.Data {
		if err := validateChatKitThread(suite, &page.Data[i]); err != nil {
			return err
		}
	}
	return nil
}

func validateChatKitThreadItemPage(suite string, page *pagination.ConversationCursorPage[openai.ChatKitThreadItemListDataUnion]) error {
	if page == nil {
		return fail(suite, "thread item page is nil")
	}
	if !page.JSON.Data.Valid() {
		return fail(suite, "thread item page missing data")
	}
	if !page.JSON.HasMore.Valid() {
		return fail(suite, "thread item page missing has_more")
	}
	for i := range page.Data {
		if err := validateChatKitThreadItem(suite, &page.Data[i]); err != nil {
			return err
		}
	}
	return nil
}

func validateChatKitThreadItem(suite string, item *openai.ChatKitThreadItemListDataUnion) error {
	if item == nil {
		return fail(suite, "thread item is nil")
	}
	if item.ID == "" {
		return fail(suite, "thread item missing id")
	}
	if item.Type == "" {
		return fail(suite, "thread item missing type")
	}
	if item.ThreadID == "" {
		return fail(suite, "thread item missing thread_id")
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