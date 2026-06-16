package suites

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/conversations"
	"github.com/openai/openai-go/v3/packages/pagination"
	"github.com/openai/openai-go/v3/responses"
	"github.com/openai/openai-go/v3/shared"
)

const conversationItemText = "Remember this compatibility test item."

// Conversations verifies the Conversations API lifecycle via client.Conversations.*.
type Conversations struct{}

func (Conversations) Name() string { return "conversations" }
func (Conversations) Description() string {
	return "Conversations API lifecycle (POST/GET/DELETE /v1/conversations and /v1/conversations/{id}/items)"
}

func (Conversations) Run(ctx context.Context, client openai.Client, _ *config.Config) error {
	deleted := false
	var conversationID string
	defer func() {
		if conversationID != "" && !deleted {
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_, _ = client.Conversations.Delete(cleanupCtx, conversationID)
		}
	}()

	created, err := client.Conversations.New(ctx, conversations.ConversationNewParams{
		Metadata: shared.Metadata{
			"suite": "conversations",
		},
	})
	if err != nil {
		return fmt.Errorf("conversation create failed: %w", err)
	}
	if err := validateConversationObject("conversations", created); err != nil {
		return err
	}
	conversationID = created.ID

	got, err := client.Conversations.Get(ctx, conversationID)
	if err != nil {
		return fmt.Errorf("conversation get failed: %w", err)
	}
	if err := validateConversationObject("conversations", got); err != nil {
		return err
	}
	if got.ID != conversationID {
		return fail("conversations", fmt.Sprintf("get id is %q, want %q", got.ID, conversationID))
	}

	updated, err := client.Conversations.Update(ctx, conversationID, conversations.ConversationUpdateParams{
		Metadata: shared.Metadata{
			"suite":  "conversations",
			"status": "updated",
		},
	})
	if err != nil {
		return fmt.Errorf("conversation update failed: %w", err)
	}
	if err := validateConversationObject("conversations", updated); err != nil {
		return err
	}
	if updated.ID != conversationID {
		return fail("conversations", fmt.Sprintf("update id is %q, want %q", updated.ID, conversationID))
	}

	createdItems, err := client.Conversations.Items.New(ctx, conversationID, conversations.ItemNewParams{
		Items: []responses.ResponseInputItemUnionParam{{
			OfMessage: &responses.EasyInputMessageParam{
				Content: responses.EasyInputMessageContentUnionParam{
					OfString: openai.String(conversationItemText),
				},
				Role: responses.EasyInputMessageRoleUser,
				Type: responses.EasyInputMessageTypeMessage,
			},
		}},
	})
	if err != nil {
		return fmt.Errorf("conversation item create failed: %w", err)
	}
	if err := validateConversationItemList("conversations", createdItems); err != nil {
		return err
	}
	if len(createdItems.Data) == 0 {
		return fail("conversations", "item create returned empty data")
	}
	itemID := createdItems.Data[0].ID
	if !conversationItemsContainText(createdItems.Data, conversationItemText) {
		return fail("conversations", "item create response missing submitted message text")
	}

	itemPage, err := client.Conversations.Items.List(ctx, conversationID, conversations.ItemListParams{
		Limit: openai.Int(10),
		Order: conversations.ItemListParamsOrderAsc,
	})
	if err != nil {
		return fmt.Errorf("conversation item list failed: %w", err)
	}
	if err := validateConversationItemPage("conversations", itemPage); err != nil {
		return err
	}
	if !conversationItemsContainID(itemPage.Data, itemID) {
		return fail("conversations", "created item missing from list response")
	}
	if !conversationItemsContainText(itemPage.Data, conversationItemText) {
		return fail("conversations", "list response missing submitted message text")
	}

	item, err := client.Conversations.Items.Get(ctx, conversationID, itemID, conversations.ItemGetParams{})
	if err != nil {
		return fmt.Errorf("conversation item get failed: %w", err)
	}
	if err := validateConversationItem("conversations", item); err != nil {
		return err
	}
	if item.ID != itemID {
		return fail("conversations", fmt.Sprintf("get item id is %q, want %q", item.ID, itemID))
	}
	if !conversationItemsContainText([]conversations.ConversationItemUnion{*item}, conversationItemText) {
		return fail("conversations", "get response missing submitted message text")
	}

	afterItemDelete, err := client.Conversations.Items.Delete(ctx, conversationID, itemID)
	if err != nil {
		return fmt.Errorf("conversation item delete failed: %w", err)
	}
	if err := validateConversationObject("conversations", afterItemDelete); err != nil {
		return err
	}
	if afterItemDelete.ID != conversationID {
		return fail("conversations", fmt.Sprintf("item delete conversation id is %q, want %q", afterItemDelete.ID, conversationID))
	}
	_, itemGetErr := client.Conversations.Items.Get(ctx, conversationID, itemID, conversations.ItemGetParams{})
	if itemGetErr == nil {
		return fail("conversations", "item get after delete succeeded; item still exists")
	}
	var itemAPIError *openai.Error
	if !errors.As(itemGetErr, &itemAPIError) {
		return fmt.Errorf("item get after delete failed: %w", itemGetErr)
	}
	if itemAPIError.StatusCode != http.StatusNotFound {
		return fail("conversations", fmt.Sprintf("item get after delete returned status %d, want 404", itemAPIError.StatusCode))
	}

	deletedResp, err := client.Conversations.Delete(ctx, conversationID)
	if err != nil {
		return fmt.Errorf("conversation delete failed: %w", err)
	}
	if err := validateConversationDeleted("conversations", deletedResp); err != nil {
		return err
	}
	if deletedResp.ID != conversationID {
		return fail("conversations", fmt.Sprintf("delete id is %q, want %q", deletedResp.ID, conversationID))
	}
	_, convGetErr := client.Conversations.Get(ctx, conversationID)
	if convGetErr == nil {
		return fail("conversations", "conversation get after delete succeeded; conversation still exists")
	}
	var convAPIError *openai.Error
	if !errors.As(convGetErr, &convAPIError) {
		return fmt.Errorf("conversation get after delete failed: %w", convGetErr)
	}
	if convAPIError.StatusCode != http.StatusNotFound {
		return fail("conversations", fmt.Sprintf("conversation get after delete returned status %d, want 404", convAPIError.StatusCode))
	}
	deleted = true
	return nil
}

func validateConversationObject(suite string, conversation *conversations.Conversation) error {
	if conversation == nil {
		return fail(suite, "conversation is nil")
	}
	if conversation.ID == "" {
		return fail(suite, "conversation missing id")
	}
	if !conversation.JSON.CreatedAt.Valid() {
		return fail(suite, "conversation missing created_at")
	}
	if !conversation.JSON.Metadata.Valid() {
		return fail(suite, "conversation missing metadata")
	}
	if !conversation.JSON.Object.Valid() {
		return fail(suite, "conversation missing object")
	}
	if string(conversation.Object) != "conversation" {
		return fail(suite, fmt.Sprintf("conversation object is %q, want conversation", conversation.Object))
	}
	return nil
}

func validateConversationDeleted(suite string, deleted *conversations.ConversationDeletedResource) error {
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
	if string(deleted.Object) != "conversation.deleted" {
		return fail(suite, fmt.Sprintf("delete object is %q, want conversation.deleted", deleted.Object))
	}
	return nil
}

func validateConversationItemList(suite string, list *conversations.ConversationItemList) error {
	if list == nil {
		return fail(suite, "item list is nil")
	}
	if !list.JSON.Data.Valid() {
		return fail(suite, "item list missing data")
	}
	if !list.JSON.HasMore.Valid() {
		return fail(suite, "item list missing has_more")
	}
	if !list.JSON.Object.Valid() {
		return fail(suite, "item list missing object")
	}
	if string(list.Object) != "list" {
		return fail(suite, fmt.Sprintf("item list object is %q, want list", list.Object))
	}
	for i := range list.Data {
		if err := validateConversationItem(suite, &list.Data[i]); err != nil {
			return err
		}
	}
	return nil
}

func validateConversationItemPage(suite string, page *pagination.ConversationCursorPage[conversations.ConversationItemUnion]) error {
	if page == nil {
		return fail(suite, "item page is nil")
	}
	if !page.JSON.HasMore.Valid() {
		return fail(suite, "item page missing has_more")
	}
	for i := range page.Data {
		if err := validateConversationItem(suite, &page.Data[i]); err != nil {
			return err
		}
	}
	return nil
}

func validateConversationItem(suite string, item *conversations.ConversationItemUnion) error {
	if item == nil {
		return fail(suite, "conversation item is nil")
	}
	if item.ID == "" {
		return fail(suite, "conversation item missing id")
	}
	if item.Type == "" {
		return fail(suite, "conversation item missing type")
	}
	if item.Type == "message" && item.Role == "" {
		return fail(suite, "conversation message item missing role")
	}
	return nil
}

func conversationItemsContainID(items []conversations.ConversationItemUnion, itemID string) bool {
	for _, item := range items {
		if item.ID == itemID {
			return true
		}
	}
	return false
}

func conversationItemsContainText(items []conversations.ConversationItemUnion, want string) bool {
	for _, item := range items {
		if item.Type != "message" {
			continue
		}
		for _, content := range item.Content.OfMessageContentArray {
			if content.Text == want {
				return true
			}
		}
	}
	return false
}
