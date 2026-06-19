package mockserver

import "net/http"

func (s *Server) handleVideoCreate(w http.ResponseWriter, r *http.Request) {
	model := "sora-2"
	prompt := "mock video prompt"
	seconds := "4"
	size := "720x1280"

	if err := r.ParseMultipartForm(1 << 20); err == nil && r.MultipartForm != nil {
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
	lastID := ""
	for i, video := range items {
		data[i] = videoPayload(video)
		lastID = video.id
	}
	writeJSON(w, map[string]any{
		"data":     data,
		"has_more": false,
		"last_id":  lastID,
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

func (s *Server) handleVideoContent(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, ok := s.videoStore.get(id); !ok {
		writeNotFound(w, "Video not found", "video_id")
		return
	}
	w.Header().Set("Content-Type", "video/mp4")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(mockVideoContent)
}

func firstFormValue(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}