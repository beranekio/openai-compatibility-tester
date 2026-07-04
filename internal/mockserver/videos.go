package mockserver

import (
	"encoding/json"
	"io"
	"net/http"
)

func (s *Server) handleVideoCreate(w http.ResponseWriter, r *http.Request) {
	model := "sora-2"
	prompt := "mock video prompt"
	seconds := "4"
	size := "720x1280"

	if err := r.ParseMultipartForm(1 << 20); err == nil && r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
		if values := r.MultipartForm.Value; values != nil {
			if v := firstFormValue(values["model"]); v != "" {
				model = v
			}
			if v := firstFormValue(values["prompt"]); v != "" {
				prompt = v
			}
			if v := firstFormValue(values["seconds"]); v != "" {
				seconds = v
			}
			if v := firstFormValue(values["size"]); v != "" {
				size = v
			}
		}
	}

	video := s.videoStore.create(model, prompt, seconds, size)
	writeJSON(w, videoPayload(video))
}

func (s *Server) handleVideoGet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	video, ok := s.videoStore.get(id)
	if !ok {
		writeNotFound(w, "Video not found", "video_id")
		return
	}
	writeJSON(w, videoPayload(video))
}

func (s *Server) handleVideoList(w http.ResponseWriter, _ *http.Request) {
	items := s.videoStore.list()
	data := make([]map[string]any, len(items))
	firstID := ""
	lastID := ""
	for i, video := range items {
		data[i] = videoPayload(video)
		if i == 0 {
			firstID = video.id
		}
		lastID = video.id
	}
	writeJSON(w, map[string]any{
		"object":   "list",
		"data":     data,
		"first_id": firstID,
		"last_id":  lastID,
		"has_more": false,
	})
}

func (s *Server) handleVideoDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !s.videoStore.delete(id) {
		writeNotFound(w, "Video not found", "video_id")
		return
	}
	writeJSON(w, map[string]any{
		"id":      id,
		"object":  "video.deleted",
		"deleted": true,
	})
}

// handleVideoSubResource dispatches the two-segment GET routes under
// /v1/videos/{id}/{rest}: content download (rest == "content") and character
// retrieval (id == "characters", rest == character id). They are merged into a
// single handler because Go 1.22 ServeMux considers "characters/{id}" and
// "{id}/content" mutually ambiguous when registered separately.
func (s *Server) handleVideoSubResource(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	rest := r.PathValue("rest")
	if id == "characters" {
		char, ok := s.videoStore.getCharacter(rest)
		if !ok {
			writeNotFound(w, "Video character not found", "character_id")
			return
		}
		writeJSON(w, videoCharacterPayload(char))
		return
	}
	if rest == "content" {
		if _, ok := s.videoStore.get(id); !ok {
			writeNotFound(w, "Video not found", "video_id")
			return
		}
		w.Header().Set("Content-Type", "video/mp4")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(mockVideoContent)
		return
	}
	writeNotFound(w, "Unknown video sub-resource", "path")
}

// videoReferenceID extracts the referenced video id from a multipart "video"
// form field, which the SDK serializes as a JSON object ({"id":"..."}). Returns
// "" when the field is absent or unparseable; the mock is lenient about the
// reference existing.
func videoReferenceID(values []string) string {
	raw := firstFormValue(values)
	if raw == "" {
		return ""
	}
	var ref struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(raw), &ref); err != nil {
		return ""
	}
	return ref.ID
}

func (s *Server) handleVideoEdit(w http.ResponseWriter, r *http.Request) {
	prompt := "mock edit prompt"
	refID := ""
	if err := r.ParseMultipartForm(1 << 20); err == nil && r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
		if v := firstFormValue(r.MultipartForm.Value["prompt"]); v != "" {
			prompt = v
		}
		refID = videoReferenceID(r.MultipartForm.Value["video"])
	}
	video := s.videoStore.createVariant("sora-2", prompt, "4", "720x1280", refID)
	writeJSON(w, videoPayload(video))
}

func (s *Server) handleVideoExtend(w http.ResponseWriter, r *http.Request) {
	prompt := "mock extend prompt"
	seconds := "4"
	refID := ""
	if err := r.ParseMultipartForm(1 << 20); err == nil && r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
		if v := firstFormValue(r.MultipartForm.Value["prompt"]); v != "" {
			prompt = v
		}
		if v := firstFormValue(r.MultipartForm.Value["seconds"]); v != "" {
			seconds = v
		}
		refID = videoReferenceID(r.MultipartForm.Value["video"])
	}
	video := s.videoStore.createVariant("sora-2", prompt, seconds, "720x1280", refID)
	writeJSON(w, videoPayload(video))
}

func (s *Server) handleVideoRemix(w http.ResponseWriter, r *http.Request) {
	videoID := r.PathValue("id")
	prompt := "mock remix prompt"
	if r.Body != nil {
		var req struct {
			Prompt string `json:"prompt"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil && req.Prompt != "" {
			prompt = req.Prompt
		}
	}
	video := s.videoStore.createVariant("sora-2", prompt, "4", "720x1280", videoID)
	writeJSON(w, videoPayload(video))
}

func (s *Server) handleVideoCharacterCreate(w http.ResponseWriter, r *http.Request) {
	name := "mock-character"
	if err := r.ParseMultipartForm(1 << 20); err == nil && r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
		if v := firstFormValue(r.MultipartForm.Value["name"]); v != "" {
			name = v
		}
		if file, _, err := r.FormFile("video"); err == nil {
			_, _ = io.Copy(io.Discard, file)
			_ = file.Close()
		}
	}
	char := s.videoStore.createCharacter(name)
	writeJSON(w, videoCharacterPayload(char))
}

func firstFormValue(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}