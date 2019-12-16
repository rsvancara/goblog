package models

import (
	"github.com/segmentio/ksuid"
)

// Generate a unique identifier
func genUUID() string {
	id := ksuid.New()
	return id.String()
}
