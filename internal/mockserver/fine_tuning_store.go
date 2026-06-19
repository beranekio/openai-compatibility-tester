package mockserver

import (
	"sort"
	"strconv"
	"sync"
)

type storedFineTuningCheckpoint struct {
	id                       string
	fineTunedModelCheckpoint string
	jobID                    string
	stepNumber               int64
	createdAt                int64
}

type storedFineTuningJob struct {
	id           string
	model        string
	trainingFile string
	status       string
	createdAt    int64
	checkpoints  []storedFineTuningCheckpoint
}

type fineTuningStore struct {
	mu      sync.Mutex
	nextJob int
	nextCP  int
	jobs    map[string]storedFineTuningJob
}

func newFineTuningStore() *fineTuningStore {
	return &fineTuningStore{
		jobs: make(map[string]storedFineTuningJob),
	}
}

func (s *fineTuningStore) allocateJobID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextJob++
	return "ftjob-mock-" + strconv.Itoa(s.nextJob)
}

func (s *fineTuningStore) save(job storedFineTuningJob) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[job.id] = job
}

func (s *fineTuningStore) get(id string) (storedFineTuningJob, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	job, ok := s.jobs[id]
	if !ok {
		return storedFineTuningJob{}, false
	}
	return job, true
}

func (s *fineTuningStore) list() []storedFineTuningJob {
	s.mu.Lock()
	defer s.mu.Unlock()
	items := make([]storedFineTuningJob, 0, len(s.jobs))
	for _, job := range s.jobs {
		items = append(items, job)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].id > items[j].id
	})
	return items
}

func (s *fineTuningStore) advanceStatus(id string) (storedFineTuningJob, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	job, ok := s.jobs[id]
	if !ok {
		return storedFineTuningJob{}, false
	}
	switch job.status {
	case "validating_files":
		job.status = "queued"
	case "queued":
		job.status = "running"
		job.checkpoints = append(job.checkpoints, s.newCheckpointLocked(job))
	case "running":
		job.status = "succeeded"
	}
	s.jobs[id] = job
	return job, true
}

func (s *fineTuningStore) cancel(id string) (storedFineTuningJob, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	job, ok := s.jobs[id]
	if !ok {
		return storedFineTuningJob{}, false
	}
	job.status = "cancelled"
	s.jobs[id] = job
	return job, true
}

func (s *fineTuningStore) newCheckpointLocked(job storedFineTuningJob) storedFineTuningCheckpoint {
	s.nextCP++
	checkpointID := "ftckpt-mock-" + strconv.Itoa(s.nextCP)
	return storedFineTuningCheckpoint{
		id:                       checkpointID,
		fineTunedModelCheckpoint: "ft:" + job.model + ":mock:" + job.id + ":" + checkpointID,
		jobID:                    job.id,
		stepNumber:               1,
		createdAt:                job.createdAt + 30,
	}
}