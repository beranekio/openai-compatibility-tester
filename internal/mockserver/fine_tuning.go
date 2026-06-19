package mockserver

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

func (s *Server) handleFineTuningJobCreate(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var req struct {
		Model        string `json:"model"`
		TrainingFile string `json:"training_file"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Model == "" {
		http.Error(w, "missing model", http.StatusBadRequest)
		return
	}
	if req.TrainingFile == "" {
		http.Error(w, "missing training_file", http.StatusBadRequest)
		return
	}
	file, ok := s.fileStore.get(req.TrainingFile)
	if !ok {
		writeNotFound(w, "File not found", "training_file")
		return
	}
	if countJSONLExamples(file.bytes) < 10 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, map[string]any{
			"error": map[string]any{
				"message": "Training file must have at least 10 examples.",
				"type":    "invalid_request_error",
				"param":   "training_file",
				"code":    "insufficient_examples",
			},
		})
		return
	}

	id := s.fineTuningStore.allocateJobID()
	job := storedFineTuningJob{
		id:           id,
		model:        req.Model,
		trainingFile: req.TrainingFile,
		status:       "validating_files",
		createdAt:    1700000000,
	}
	s.fineTuningStore.save(job)
	writeJSON(w, fineTuningJobObjectPayload(job))
}

func (s *Server) handleFineTuningJobList(w http.ResponseWriter, _ *http.Request) {
	items := s.fineTuningStore.list()
	data := make([]map[string]any, len(items))
	firstID := ""
	lastID := ""
	for i, job := range items {
		data[i] = fineTuningJobObjectPayload(job)
		if i == 0 {
			firstID = job.id
		}
		lastID = job.id
	}
	writeJSON(w, map[string]any{
		"object":   "list",
		"data":     data,
		"first_id": firstID,
		"last_id":  lastID,
		"has_more": false,
	})
}

func (s *Server) handleFineTuningJobGet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	job, ok := s.fineTuningStore.advanceStatus(id)
	if !ok {
		writeNotFound(w, "Fine tuning job not found", "id")
		return
	}
	writeJSON(w, fineTuningJobObjectPayload(job))
}

func (s *Server) handleFineTuningJobCancel(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	job, ok, alreadyTerminal := s.fineTuningStore.cancel(id)
	if !ok {
		writeNotFound(w, "Fine tuning job not found", "id")
		return
	}
	if alreadyTerminal {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, map[string]any{
			"error": map[string]any{
				"message": "Cannot cancel a fine-tuning job that has already finished.",
				"type":    "invalid_request_error",
				"param":   "id",
				"code":    "invalid_job_status",
			},
		})
		return
	}
	writeJSON(w, fineTuningJobObjectPayload(job))
}

func (s *Server) handleFineTuningJobCheckpointList(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	job, ok := s.fineTuningStore.get(id)
	if !ok {
		writeNotFound(w, "Fine tuning job not found", "id")
		return
	}
	data := make([]map[string]any, len(job.checkpoints))
	firstID := ""
	lastID := ""
	for i, checkpoint := range job.checkpoints {
		data[i] = fineTuningCheckpointObjectPayload(checkpoint)
		if i == 0 {
			firstID = checkpoint.id
		}
		lastID = checkpoint.id
	}
	writeJSON(w, map[string]any{
		"object":   "list",
		"data":     data,
		"first_id": firstID,
		"last_id":  lastID,
		"has_more": false,
	})
}

func (s *Server) handleFineTuningCheckpointPermissionList(w http.ResponseWriter, r *http.Request) {
	_ = r.PathValue("fine_tuned_model_checkpoint")
	writeJSON(w, map[string]any{
		"object":   "list",
		"data":     []map[string]any{},
		"first_id": "",
		"last_id":  "",
		"has_more": false,
	})
}

func fineTuningJobObjectPayload(job storedFineTuningJob) map[string]any {
	payload := map[string]any{
		"id":              job.id,
		"object":          "fine_tuning.job",
		"created_at":      job.createdAt,
		"model":           job.model,
		"organization_id": "org-mock",
		"training_file":   job.trainingFile,
		"status":          job.status,
		"seed":            42,
		"hyperparameters": map[string]any{
			"n_epochs":                 "auto",
			"batch_size":               "auto",
			"learning_rate_multiplier": "auto",
		},
		"result_files":     []string{},
		"error":            nil,
		"fine_tuned_model": nil,
		"finished_at":      nil,
		"trained_tokens":   nil,
		"validation_file":  nil,
	}
	switch job.status {
	case "succeeded":
		if len(job.checkpoints) > 0 {
			payload["fine_tuned_model"] = job.checkpoints[len(job.checkpoints)-1].fineTunedModelCheckpoint
		} else {
			payload["fine_tuned_model"] = "ft:" + job.model + ":mock:" + job.id
		}
		payload["finished_at"] = job.createdAt + 120
		payload["trained_tokens"] = 128
		payload["result_files"] = []string{"file-mock-ft-results"}
	case "cancelled":
		payload["finished_at"] = job.createdAt + 60
	case "failed":
		payload["error"] = map[string]any{
			"code":    "mock_failure",
			"message": "mock fine-tuning job failed",
			"param":   nil,
		}
		payload["finished_at"] = job.createdAt + 60
	}
	return payload
}

func countJSONLExamples(data []byte) int {
	count := 0
	for _, line := range bytes.Split(data, []byte("\n")) {
		if len(strings.TrimSpace(string(line))) > 0 {
			count++
		}
	}
	return count
}

func fineTuningCheckpointObjectPayload(checkpoint storedFineTuningCheckpoint) map[string]any {
	return map[string]any{
		"id":                          checkpoint.id,
		"object":                      "fine_tuning.job.checkpoint",
		"created_at":                  checkpoint.createdAt,
		"fine_tuned_model_checkpoint": checkpoint.fineTunedModelCheckpoint,
		"fine_tuning_job_id":          checkpoint.jobID,
		"step_number":                 checkpoint.stepNumber,
		"metrics": map[string]any{
			"step":                           float64(checkpoint.stepNumber),
			"train_loss":                     0.12,
			"train_mean_token_accuracy":      0.91,
			"valid_loss":                     0.15,
			"valid_mean_token_accuracy":      0.89,
			"full_valid_loss":                0.14,
			"full_valid_mean_token_accuracy": 0.9,
		},
	}
}