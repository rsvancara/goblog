package models

import (
	"github.com/segmentio/ksuid"
)

// Generate a unique identifier
func GenUUID() string {
	id := ksuid.New()
	return id.String()
}
