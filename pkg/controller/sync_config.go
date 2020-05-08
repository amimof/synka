package controller

import (
	"strconv"
)

const (
	syncAnnotationKey         = "synka.io/sync"
	clustersAnnotationKey     = "synka.io/clusters"
	skipExistingAnnotationKey = "synka.io/skip-existing"
)

// SyncConfig is the configuration of the sync process. It defines how a resource is synchronised.
type SyncConfig struct {
	Sync         bool
	SkipExisting bool
}

// NewSyncConfig returns a SyncConfig with default values
func NewSyncConfig() SyncConfig {
	return SyncConfig{
		Sync:         true,
		SkipExisting: false,
	}
}

// NewSyncConfigFrom creates a SyncConfig from a map[string]string
func NewSyncConfigFrom(m map[string]string) SyncConfig {
	sync, _ := strconv.ParseBool(getValFromMap(syncAnnotationKey, m))
	skipExisting, _ := strconv.ParseBool(getValFromMap(skipExistingAnnotationKey, m))
	return SyncConfig{
		Sync:         sync,
		SkipExisting: skipExisting,
	}
}

// getValFromMap returns the value of a key in a map of strings
func getValFromMap(key string, m map[string]string) string {
	if val, ok := m[key]; ok {
		return val
	}
	return ""
}
