package mockserver

import (
	"sort"
	"strconv"
	"sync"
)

type storedSkillVersion struct {
	id          string
	skillID     string
	version     string
	name        string
	description string
	createdAt   int64
	content     []byte
}

type storedSkill struct {
	id             string
	name           string
	description    string
	createdAt      int64
	defaultVersion string
	latestVersion  string
	versions       map[string]storedSkillVersion
}

type skillStore struct {
	mu        sync.Mutex
	next      int
	nextVer   int
	skills    map[string]storedSkill
}

func newSkillStore() *skillStore {
	return &skillStore{
		skills: make(map[string]storedSkill),
	}
}

func (s *skillStore) create(content []byte) storedSkill {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.next++
	skillID := "skill_mock_" + strconv.Itoa(s.next)
	s.nextVer++
	versionID := "skillver_mock_" + strconv.Itoa(s.nextVer)
	version := "1"
	versionRecord := storedSkillVersion{
		id:          versionID,
		skillID:     skillID,
		version:     version,
		name:        "compatibility-test-skill",
		description: "compatibility test skill",
		createdAt:   1700000000,
		content:     append([]byte(nil), content...),
	}
	skill := storedSkill{
		id:             skillID,
		name:           "compatibility-test-skill",
		description:    "compatibility test skill",
		createdAt:      1700000000,
		defaultVersion: version,
		latestVersion:  version,
		versions: map[string]storedSkillVersion{
			version: versionRecord,
		},
	}
	s.skills[skillID] = skill
	return cloneSkill(skill)
}

func (s *skillStore) get(id string) (storedSkill, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	skill, ok := s.skills[id]
	if !ok {
		return storedSkill{}, false
	}
	return cloneSkill(skill), true
}

func (s *skillStore) updateDefaultVersion(id, version string) (storedSkill, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	skill, ok := s.skills[id]
	if !ok {
		return storedSkill{}, false
	}
	if _, ok := skill.versions[version]; !ok {
		return storedSkill{}, false
	}
	skill.defaultVersion = version
	s.skills[id] = skill
	return cloneSkill(skill), true
}

func (s *skillStore) delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.skills[id]; !ok {
		return false
	}
	delete(s.skills, id)
	return true
}

func (s *skillStore) list() []storedSkill {
	s.mu.Lock()
	defer s.mu.Unlock()
	items := make([]storedSkill, 0, len(s.skills))
	for _, skill := range s.skills {
		items = append(items, cloneSkill(skill))
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].id > items[j].id
	})
	return items
}

func (s *skillStore) createVersion(skillID string, content []byte, setDefault bool) (storedSkillVersion, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	skill, ok := s.skills[skillID]
	if !ok {
		return storedSkillVersion{}, false
	}
	nextVersionNum := 1
	if latest, err := strconv.Atoi(skill.latestVersion); err == nil {
		nextVersionNum = latest + 1
	}
	version := strconv.Itoa(nextVersionNum)
	s.nextVer++
	versionRecord := storedSkillVersion{
		id:          "skillver_mock_" + strconv.Itoa(s.nextVer),
		skillID:     skillID,
		version:     version,
		name:        skill.name,
		description: skill.description,
		createdAt:   1700000000,
		content:     append([]byte(nil), content...),
	}
	skill.versions[version] = versionRecord
	skill.latestVersion = version
	if setDefault {
		skill.defaultVersion = version
	}
	s.skills[skillID] = skill
	return cloneSkillVersion(versionRecord), true
}

func (s *skillStore) getVersion(skillID, version string) (storedSkillVersion, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	skill, ok := s.skills[skillID]
	if !ok {
		return storedSkillVersion{}, false
	}
	versionRecord, ok := skill.versions[version]
	if !ok {
		return storedSkillVersion{}, false
	}
	return cloneSkillVersion(versionRecord), true
}

func (s *skillStore) listVersions(skillID string) ([]storedSkillVersion, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	skill, ok := s.skills[skillID]
	if !ok {
		return nil, false
	}
	items := make([]storedSkillVersion, 0, len(skill.versions))
	for _, version := range skill.versions {
		items = append(items, cloneSkillVersion(version))
	}
	sort.Slice(items, func(i, j int) bool {
		left, errLeft := strconv.Atoi(items[i].version)
		right, errRight := strconv.Atoi(items[j].version)
		if errLeft == nil && errRight == nil {
			return left > right
		}
		return items[i].version > items[j].version
	})
	return items, true
}

func (s *skillStore) deleteVersion(skillID, version string) (storedSkillVersion, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	skill, ok := s.skills[skillID]
	if !ok {
		return storedSkillVersion{}, false
	}
	versionRecord, ok := skill.versions[version]
	if !ok {
		return storedSkillVersion{}, false
	}
	delete(skill.versions, version)
	if skill.defaultVersion == version {
		skill.defaultVersion = ""
	}
	if skill.latestVersion == version {
		skill.latestVersion = highestSkillVersion(skill.versions)
	}
	s.skills[skillID] = skill
	return cloneSkillVersion(versionRecord), true
}

func (s *skillStore) defaultContent(skillID string) ([]byte, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	skill, ok := s.skills[skillID]
	if !ok {
		return nil, false
	}
	versionRecord, ok := skill.versions[skill.defaultVersion]
	if !ok {
		return nil, false
	}
	return append([]byte(nil), versionRecord.content...), true
}

func highestSkillVersion(versions map[string]storedSkillVersion) string {
	highest := 0
	highestVersion := ""
	for version := range versions {
		num, err := strconv.Atoi(version)
		if err != nil {
			if highestVersion == "" || version > highestVersion {
				highestVersion = version
			}
			continue
		}
		if num > highest {
			highest = num
			highestVersion = version
		}
	}
	return highestVersion
}

func cloneSkill(skill storedSkill) storedSkill {
	cloned := storedSkill{
		id:             skill.id,
		name:           skill.name,
		description:    skill.description,
		createdAt:      skill.createdAt,
		defaultVersion: skill.defaultVersion,
		latestVersion:  skill.latestVersion,
		versions:       make(map[string]storedSkillVersion, len(skill.versions)),
	}
	for version, record := range skill.versions {
		cloned.versions[version] = cloneSkillVersion(record)
	}
	return cloned
}

func cloneSkillVersion(version storedSkillVersion) storedSkillVersion {
	return storedSkillVersion{
		id:          version.id,
		skillID:     version.skillID,
		version:     version.version,
		name:        version.name,
		description: version.description,
		createdAt:   version.createdAt,
		content:     append([]byte(nil), version.content...),
	}
}

func skillPayload(skill storedSkill) map[string]any {
	return map[string]any{
		"id":              skill.id,
		"object":          "skill",
		"created_at":      skill.createdAt,
		"default_version": skill.defaultVersion,
		"description":     skill.description,
		"latest_version":  skill.latestVersion,
		"name":            skill.name,
	}
}

func skillVersionPayload(version storedSkillVersion) map[string]any {
	return map[string]any{
		"id":          version.id,
		"object":      "skill.version",
		"created_at":  version.createdAt,
		"description": version.description,
		"name":        version.name,
		"skill_id":    version.skillID,
		"version":     version.version,
	}
}