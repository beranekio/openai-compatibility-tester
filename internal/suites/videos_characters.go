package suites

import (
	"bytes"
	"context"
	"fmt"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
)

// VideosCharacters verifies the video Characters sub-resource via
// client.Videos.NewCharacter and client.Videos.GetCharacter.
type VideosCharacters struct{}

func (VideosCharacters) Name() string { return "videos_characters" }
func (VideosCharacters) Description() string {
	return "Videos API characters (POST /v1/videos/characters, then GET /v1/videos/characters/{id})"
}

func (VideosCharacters) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	const characterName = "compat-test-character"
	created, err := client.Videos.NewCharacter(ctx, openai.VideoNewCharacterParams{
		Name: characterName,
		Video: bytes.NewReader([]byte("mock-video-source")),
	})
	if err != nil {
		return fmt.Errorf("video character create failed: %w", err)
	}
	if created == nil {
		return fail("videos_characters", "create character is nil")
	}
	if created.ID == "" {
		return fail("videos_characters", "create character missing id")
	}
	if !created.JSON.CreatedAt.Valid() {
		return fail("videos_characters", "create character missing created_at")
	}
	if created.Name != characterName {
		return fail("videos_characters", fmt.Sprintf("create character name is %q, want %q", created.Name, characterName))
	}

	got, err := client.Videos.GetCharacter(ctx, created.ID)
	if err != nil {
		return fmt.Errorf("video character get failed: %w", err)
	}
	if got == nil {
		return fail("videos_characters", "get character is nil")
	}
	if got.ID != created.ID {
		return fail("videos_characters", fmt.Sprintf("get character id is %q, want %q", got.ID, created.ID))
	}
	if !got.JSON.CreatedAt.Valid() {
		return fail("videos_characters", "get character missing created_at")
	}
	if got.Name != characterName {
		return fail("videos_characters", fmt.Sprintf("get character name is %q, want %q", got.Name, characterName))
	}
	return nil
}
