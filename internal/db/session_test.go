package db

import (
	"goblog/internal/config"
	"testing"
)

func TestSession(t *testing.T) {

	var cfg config.AppConfig
	cfg.Dburi = ""
	var s Session

	err := s.NewSession(cfg)
	if err != nil {
		t.Fatalf(`Session() %v, error`, err)
	}
}
