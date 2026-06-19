package mockserver

import (
	"sort"
	"strconv"
	"sync"
)

type storedContainer struct {
	id           string
	name         string
	status       string
	createdAt    int64
	lastActiveAt int64
	memoryLimit  string
	files        map[string]storedContainerFile
}

type storedContainerFile struct {
	id          string
	containerID string
	bytes       []byte
	path        string
	source      string
	createdAt   int64
}

type containerStore struct {
	mu         sync.Mutex
	next       int
	nextFile   int
	containers map[string]storedContainer
}

func newContainerStore() *containerStore {
	return &containerStore{
		containers: make(map[string]storedContainer),
	}
}

func (s *containerStore) create(name string) storedContainer {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.next++
	container := storedContainer{
		id:           "ctr_mock_" + strconv.Itoa(s.next),
		name:         name,
		status:       "active",
		createdAt:    1700000000,
		lastActiveAt: 1700000000,
		memoryLimit:  "1g",
		files:        make(map[string]storedContainerFile),
	}
	s.containers[container.id] = container
	return container
}

func (s *containerStore) get(id string) (storedContainer, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	container, ok := s.containers[id]
	if !ok {
		return storedContainer{}, false
	}
	return cloneContainer(container), true
}

func (s *containerStore) delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.containers[id]; !ok {
		return false
	}
	delete(s.containers, id)
	return true
}

func (s *containerStore) list() []storedContainer {
	s.mu.Lock()
	defer s.mu.Unlock()
	items := make([]storedContainer, 0, len(s.containers))
	for _, container := range s.containers {
		items = append(items, cloneContainer(container))
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].id < items[j].id
	})
	return items
}

func (s *containerStore) addFile(containerID string, data []byte, path, source string) (storedContainerFile, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	container, ok := s.containers[containerID]
	if !ok {
		return storedContainerFile{}, false
	}
	s.nextFile++
	file := storedContainerFile{
		id:          "cfile_mock_" + strconv.Itoa(s.nextFile),
		containerID: containerID,
		bytes:       append([]byte(nil), data...),
		path:        path,
		source:      source,
		createdAt:   1700000000,
	}
	container.files[file.id] = file
	s.containers[containerID] = container
	return cloneContainerFile(file), true
}

func (s *containerStore) getFile(containerID, fileID string) (storedContainerFile, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	container, ok := s.containers[containerID]
	if !ok {
		return storedContainerFile{}, false
	}
	file, ok := container.files[fileID]
	if !ok {
		return storedContainerFile{}, false
	}
	return cloneContainerFile(file), true
}

func (s *containerStore) deleteFile(containerID, fileID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	container, ok := s.containers[containerID]
	if !ok {
		return false
	}
	if _, ok := container.files[fileID]; !ok {
		return false
	}
	delete(container.files, fileID)
	s.containers[containerID] = container
	return true
}

func (s *containerStore) listFiles(containerID string) ([]storedContainerFile, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	container, ok := s.containers[containerID]
	if !ok {
		return nil, false
	}
	items := make([]storedContainerFile, 0, len(container.files))
	for _, file := range container.files {
		items = append(items, cloneContainerFile(file))
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].id < items[j].id
	})
	return items, true
}

func cloneContainer(container storedContainer) storedContainer {
	origFiles := container.files
	container.files = make(map[string]storedContainerFile, len(origFiles))
	for id, file := range origFiles {
		container.files[id] = cloneContainerFile(file)
	}
	return container
}

func cloneContainerFile(file storedContainerFile) storedContainerFile {
	file.bytes = append([]byte(nil), file.bytes...)
	return file
}

func containerPayload(container storedContainer) map[string]any {
	return map[string]any{
		"id":             container.id,
		"object":         "container",
		"created_at":     container.createdAt,
		"name":           container.name,
		"status":         container.status,
		"last_active_at": container.lastActiveAt,
		"memory_limit":   container.memoryLimit,
		"network_policy": map[string]any{
			"type": "disabled",
		},
	}
}

func containerFilePayload(file storedContainerFile) map[string]any {
	return map[string]any{
		"id":           file.id,
		"object":       "container.file",
		"bytes":        len(file.bytes),
		"container_id": file.containerID,
		"created_at":   file.createdAt,
		"path":         file.path,
		"source":       file.source,
	}
}