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
)

const containerCreateName = "compatibility-test-container"

// Containers verifies the Containers API lifecycle via client.Containers.*.
type Containers struct{}

func (Containers) Name() string { return "containers" }
func (Containers) Description() string {
	return "Containers API lifecycle (POST /v1/containers, GET /v1/containers, GET/DELETE /v1/containers/{id})"
}

func (Containers) Run(ctx context.Context, client openai.Client, _ *config.Config) error {
	deleted := false
	var containerID string
	defer func() {
		if containerID != "" && !deleted {
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_ = client.Containers.Delete(cleanupCtx, containerID)
		}
	}()

	created, err := client.Containers.New(ctx, openai.ContainerNewParams{
		Name: containerCreateName,
	})
	if err != nil {
		return fmt.Errorf("container create failed: %w", err)
	}
	if err := validateContainerObject("containers", created); err != nil {
		return err
	}
	containerID = created.ID
	if created.Name != containerCreateName {
		return fail("containers", fmt.Sprintf("create name is %q, want %q", created.Name, containerCreateName))
	}

	got, err := client.Containers.Get(ctx, containerID)
	if err != nil {
		return fmt.Errorf("container get failed: %w", err)
	}
	if err := validateContainerGetObject("containers", got); err != nil {
		return err
	}
	if got.ID != containerID {
		return fail("containers", fmt.Sprintf("get id is %q, want %q", got.ID, containerID))
	}
	if got.Name != containerCreateName {
		return fail("containers", fmt.Sprintf("get name is %q, want %q", got.Name, containerCreateName))
	}

	listPage, err := client.Containers.List(ctx, openai.ContainerListParams{
		Limit: openai.Int(100),
		Order: openai.ContainerListParamsOrderDesc,
	})
	if err != nil {
		return fmt.Errorf("container list failed: %w", err)
	}
	found, err := containerListContains(listPage, containerID)
	if err != nil {
		return err
	}
	if !found {
		return fail("containers", "created container missing from list response")
	}

	if err := client.Containers.Delete(ctx, containerID); err != nil {
		return fmt.Errorf("container delete failed: %w", err)
	}

	_, getErr := client.Containers.Get(ctx, containerID)
	if getErr == nil {
		return fail("containers", "get after delete succeeded; container still exists")
	}
	var apiErr *openai.Error
	if !errors.As(getErr, &apiErr) {
		return fmt.Errorf("get after delete failed: %w", getErr)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		return fail("containers", fmt.Sprintf("get after delete returned status %d, want 404", apiErr.StatusCode))
	}
	deleted = true
	return nil
}

func containerListContains(page *pagination.CursorPage[openai.ContainerListResponse], containerID string) (bool, error) {
	for page != nil {
		if err := validateContainerListPage("containers", page); err != nil {
			return false, err
		}
		for _, item := range page.Data {
			if item.ID == containerID {
				return true, nil
			}
		}
		next, err := page.GetNextPage()
		if err != nil {
			return false, fmt.Errorf("container list next page failed: %w", err)
		}
		page = next
	}
	return false, nil
}

func validateContainerObject(suite string, container *openai.ContainerNewResponse) error {
	if container == nil {
		return fail(suite, "container is nil")
	}
	if container.ID == "" {
		return fail(suite, "container missing id")
	}
	if !container.JSON.CreatedAt.Valid() {
		return fail(suite, "container missing created_at")
	}
	if !container.JSON.Name.Valid() {
		return fail(suite, "container missing name")
	}
	if !container.JSON.Object.Valid() {
		return fail(suite, "container missing object")
	}
	if container.Object != "container" {
		return fail(suite, fmt.Sprintf("container object is %q, want container", container.Object))
	}
	if !container.JSON.Status.Valid() {
		return fail(suite, "container missing status")
	}
	if container.Status == "" {
		return fail(suite, "container status is empty")
	}
	return nil
}

func validateContainerGetObject(suite string, container *openai.ContainerGetResponse) error {
	if container == nil {
		return fail(suite, "container is nil")
	}
	if container.ID == "" {
		return fail(suite, "container missing id")
	}
	if !container.JSON.CreatedAt.Valid() {
		return fail(suite, "container missing created_at")
	}
	if !container.JSON.Name.Valid() {
		return fail(suite, "container missing name")
	}
	if !container.JSON.Object.Valid() {
		return fail(suite, "container missing object")
	}
	if container.Object != "container" {
		return fail(suite, fmt.Sprintf("container object is %q, want container", container.Object))
	}
	if !container.JSON.Status.Valid() {
		return fail(suite, "container missing status")
	}
	if container.Status == "" {
		return fail(suite, "container status is empty")
	}
	return nil
}

func validateContainerListPage(suite string, page *pagination.CursorPage[openai.ContainerListResponse]) error {
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
		item := &page.Data[i]
		if item.ID == "" {
			return fail(suite, "list item missing id")
		}
		if !item.JSON.CreatedAt.Valid() {
			return fail(suite, "list item missing created_at")
		}
		if !item.JSON.Name.Valid() {
			return fail(suite, "list item missing name")
		}
		if !item.JSON.Object.Valid() {
			return fail(suite, "list item missing object")
		}
		if item.Object != "container" {
			return fail(suite, fmt.Sprintf("list item object is %q, want container", item.Object))
		}
		if !item.JSON.Status.Valid() {
			return fail(suite, "list item missing status")
		}
	}
	return nil
}

func createContainerForSuite(ctx context.Context, client openai.Client, suite, name string) (*openai.ContainerNewResponse, error) {
	created, err := client.Containers.New(ctx, openai.ContainerNewParams{
		Name: name,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: container create failed: %w", suite, err)
	}
	if err := validateContainerObject(suite, created); err != nil {
		return nil, err
	}
	return created, nil
}