package main

import (
	"os"
	"sync"
	"time"
)

var defaultCaseEntryCache = newCaseEntryCache()

type caseEntryCache struct {
	mu      sync.Mutex
	entries map[caseEntryCacheKey]cachedCaseEntry
}

type caseEntryCacheKey struct {
	BaselinePath    string
	BaselineSize    int64
	BaselineModTime time.Time
	ArtifactPath    string
	ArtifactSize    int64
	ArtifactModTime time.Time
}

type cachedCaseEntry struct {
	Stats     metrics
	RefWidth  int
	RefHeight int
	ActWidth  int
	ActHeight int
}

func newCaseEntryCache() *caseEntryCache {
	return &caseEntryCache{entries: map[caseEntryCacheKey]cachedCaseEntry{}}
}

func newCaseEntryCacheKey(baselinePath, artifactPath string) (caseEntryCacheKey, bool) {
	baselineInfo, err := os.Stat(baselinePath)
	if err != nil {
		return caseEntryCacheKey{}, false
	}
	artifactInfo, err := os.Stat(artifactPath)
	if err != nil {
		return caseEntryCacheKey{}, false
	}
	return caseEntryCacheKey{
		BaselinePath:    baselinePath,
		BaselineSize:    baselineInfo.Size(),
		BaselineModTime: baselineInfo.ModTime(),
		ArtifactPath:    artifactPath,
		ArtifactSize:    artifactInfo.Size(),
		ArtifactModTime: artifactInfo.ModTime(),
	}, true
}

func (c *caseEntryCache) lookup(key caseEntryCacheKey) (cachedCaseEntry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.entries[key]
	return entry, ok
}

func (c *caseEntryCache) store(key caseEntryCacheKey, entry cachedCaseEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = entry
}
