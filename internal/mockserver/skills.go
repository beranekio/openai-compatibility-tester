package mockserver

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
)

const skillBundleFolder = "compatibility-test-skill"

func (s *Server) handleSkillCreate(w http.ResponseWriter, r *http.Request) {
	content, err := readSkillMultipartFiles(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	skill := s.skillStore.create(content)
	writeJSON(w, skillPayload(skill))
}

func (s *Server) handleSkillList(w http.ResponseWriter, _ *http.Request) {
	items := s.skillStore.list()
	data := make([]map[string]any, len(items))
	firstID := ""
	lastID := ""
	for i, skill := range items {
		data[i] = skillPayload(skill)
		if i == 0 {
			firstID = skill.id
		}
		lastID = skill.id
	}
	writeJSON(w, map[string]any{
		"object":   "list",
		"data":     data,
		"first_id": firstID,
		"last_id":  lastID,
		"has_more": false,
	})
}

func (s *Server) handleSkillGet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	skill, ok := s.skillStore.get(id)
	if !ok {
		writeNotFound(w, "Skill not found", "skill_id")
		return
	}
	writeJSON(w, skillPayload(skill))
}

func (s *Server) handleSkillUpdate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		DefaultVersion string `json:"default_version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.DefaultVersion) == "" {
		http.Error(w, "missing default_version", http.StatusBadRequest)
		return
	}
	skill, ok := s.skillStore.updateDefaultVersion(id, req.DefaultVersion)
	if !ok {
		writeNotFound(w, "Skill not found", "skill_id")
		return
	}
	writeJSON(w, skillPayload(skill))
}

func (s *Server) handleSkillDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !s.skillStore.delete(id) {
		writeNotFound(w, "Skill not found", "skill_id")
		return
	}
	writeJSON(w, map[string]any{
		"id":      id,
		"object":  "skill.deleted",
		"deleted": true,
	})
}

func (s *Server) handleSkillContent(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	content, ok := s.skillStore.defaultContent(id)
	if !ok {
		writeNotFound(w, "Skill not found", "skill_id")
		return
	}
	zipBytes, err := skillBundleZip(content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/zip")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(zipBytes)
}

func (s *Server) handleSkillVersionCreate(w http.ResponseWriter, r *http.Request) {
	skillID := r.PathValue("id")
	content, setDefault, err := readSkillVersionMultipart(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	version, ok := s.skillStore.createVersion(skillID, content, setDefault)
	if !ok {
		writeNotFound(w, "Skill not found", "skill_id")
		return
	}
	writeJSON(w, skillVersionPayload(version))
}

func (s *Server) handleSkillVersionList(w http.ResponseWriter, r *http.Request) {
	skillID := r.PathValue("id")
	versions, ok := s.skillStore.listVersions(skillID)
	if !ok {
		writeNotFound(w, "Skill not found", "skill_id")
		return
	}
	data := make([]map[string]any, len(versions))
	firstID := ""
	lastID := ""
	for i, version := range versions {
		data[i] = skillVersionPayload(version)
		if i == 0 {
			firstID = version.id
		}
		lastID = version.id
	}
	writeJSON(w, map[string]any{
		"object":   "list",
		"data":     data,
		"first_id": firstID,
		"last_id":  lastID,
		"has_more": false,
	})
}

func (s *Server) handleSkillVersionGet(w http.ResponseWriter, r *http.Request) {
	skillID := r.PathValue("id")
	version := r.PathValue("version")
	versionRecord, ok := s.skillStore.getVersion(skillID, version)
	if !ok {
		writeNotFound(w, "Skill version not found", "version")
		return
	}
	writeJSON(w, skillVersionPayload(versionRecord))
}

func (s *Server) handleSkillVersionDelete(w http.ResponseWriter, r *http.Request) {
	skillID := r.PathValue("id")
	version := r.PathValue("version")
	versionRecord, ok := s.skillStore.deleteVersion(skillID, version)
	if !ok {
		writeNotFound(w, "Skill version not found", "version")
		return
	}
	writeJSON(w, map[string]any{
		"id":      versionRecord.id,
		"object":  "skill.version.deleted",
		"deleted": true,
		"version": versionRecord.version,
	})
}

func (s *Server) handleSkillVersionContent(w http.ResponseWriter, r *http.Request) {
	skillID := r.PathValue("id")
	version := r.PathValue("version")
	versionRecord, ok := s.skillStore.getVersion(skillID, version)
	if !ok {
		writeNotFound(w, "Skill version not found", "version")
		return
	}
	zipBytes, err := skillBundleZip(versionRecord.content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/zip")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(zipBytes)
}

func readSkillMultipartFiles(r *http.Request) ([]byte, error) {
	content, _, err := readSkillVersionMultipart(r)
	return content, err
}

func readSkillVersionMultipart(r *http.Request) ([]byte, bool, error) {
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		return nil, false, err
	}
	defer r.MultipartForm.RemoveAll()

	setDefault := false
	if value := strings.TrimSpace(r.FormValue("default")); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err == nil {
			setDefault = parsed
		}
	}

	var content []byte
	for _, field := range []string{"files[]", "files"} {
		if files := r.MultipartForm.File[field]; len(files) > 0 {
			for _, header := range files {
				file, err := header.Open()
				if err != nil {
					return nil, false, err
				}
				data, err := io.ReadAll(file)
				_ = file.Close()
				if err != nil {
					return nil, false, err
				}
				content = append(content, data...)
			}
			return content, setDefault, nil
		}
	}

	for _, field := range []string{"files[]", "files"} {
		file, _, err := r.FormFile(field)
		if err == nil {
			defer file.Close()
			content, err = io.ReadAll(file)
			return content, setDefault, err
		}
	}
	return nil, false, errors.New("missing files")
}

func skillBundleZip(skillMD []byte) ([]byte, error) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	entry, err := zw.Create(skillBundleFolder + "/SKILL.md")
	if err != nil {
		return nil, err
	}
	if _, err := entry.Write(skillMD); err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}