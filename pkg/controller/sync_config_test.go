package controller

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSyncConfig_NewSyncConfig(t *testing.T) {
	sc := NewSyncConfig()
	assert.True(t, sc.Sync, "Unexpected bool")
	assert.False(t, sc.SkipExisting, "Unexpected bool")
}

func TestSyncConfig_NewSyncConfigFrom(t *testing.T) {
	annotations := map[string]string{
		syncAnnotationKey:         "false",
		skipExistingAnnotationKey: "true",
	}

	sc := NewSyncConfigFrom(annotations)
	assert.False(t, sc.Sync, "Unexpected bool")
	assert.True(t, sc.SkipExisting, "Unexpected bool")
}

func TestSyncConfig_getValFromMap(t *testing.T) {
	m := map[string]string{
		"someKey": "someVal123",
		"myKey":   "myVal123",
		"yourKey": "yourVal123",
	}
	s := getValFromMap("myKey", m)
	assert.Equal(t, "myVal123", s, "Unexpected value")

	s = getValFromMap("wrongKey", m)
	assert.Equal(t, "", s, "Unexpected value")
}
