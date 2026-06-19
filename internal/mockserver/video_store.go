package mockserver

import (
	"sort"
	"strconv"
	"sync"
)

// mockVideoContent is a tiny placeholder returned by GET /v1/videos/{id}/content.
var mockVideoContent = []byte("mock-video-mp4-bytes")

type storedVideo struct {
	id                 string
	model              string
	prompt             string
	seconds            string
	size               string
	status             string
	progress           int64
	createdAt          int64
	completedAt        int64
	expiresAt          int64
	remixedFromVideoID string
}

type videoStore struct {
	mu     sync.Mutex
	next   int
	videos map[string]storedVideo
}

func newVideoStore() *videoStore {
	return &videoStore{
		videos: make(map[string]storedVideo),
	}
}

func (s *videoStore) create(model, prompt, seconds, size string) storedVideo {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.next++
	if model == "" {
		model = "sora-2"
	}
	if prompt == "" {
		prompt = "mock video prompt"
	}
	if seconds == "" {
		seconds = "4"
	}
	if size == "" {
		size = "720x1280"
	}
	createdAt := int64(1700000000)
	video := storedVideo{
		id:          "video_mock_" + strconv.Itoa(s.next),
		model:       model,
		prompt:      prompt,
		seconds:     seconds,
		size:        size,
		status:      "completed",
		progress:    100,
		createdAt:   createdAt,
		completedAt: createdAt + 5,
		expiresAt:   createdAt + 86400,
	}
	s.videos[video.id] = video
	return video
}

func (s *videoStore) get(id string) (storedVideo, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	video, ok := s.videos[id]
	return video, ok
}

func (s *videoStore) list() []storedVideo {
	s.mu.Lock()
	defer s.mu.Unlock()
	items := make([]storedVideo, 0, len(s.videos))
	for _, video := range s.videos {
		items = append(items, video)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].createdAt > items[j].createdAt
	})
	return items
}

func (s *videoStore) delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.videos[id]; !ok {
		return false
	}
	delete(s.videos, id)
	return true
}

func videoPayload(video storedVideo) map[string]any {
	return map[string]any{
		"id":                    video.id,
		"object":                "video",
		"created_at":            video.createdAt,
		"completed_at":          video.completedAt,
		"expires_at":            video.expiresAt,
		"model":                 video.model,
		"prompt":                video.prompt,
		"progress":              video.progress,
		"remixed_from_video_id": video.remixedFromVideoID,
		"seconds":               video.seconds,
		"size":                  video.size,
		"status":                video.status,
		"error": map[string]any{
			"code":    "",
			"message": "",
		},
	}
}