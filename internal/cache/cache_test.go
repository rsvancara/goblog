package cache

import (
	"goblog/internal/config"
	"testing"
)

func TestInitPool(t *testing.T) {

	var cfg config.AppConfig
	cfg.Dburi = ""
	var c Cache

	err := c.InitPool(cfg)
	if err != nil {
		t.Fatalf(`InitPool() %v, error`, err)
	}
}
