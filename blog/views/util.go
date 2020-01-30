package views

import (
	"blog/blog/config"
	"fmt"
)

// TemplateLoader loads the correct template directory for the site
func SiteTemplate(path string) (string, error) {

	cfg, err := config.GetConfig()

	if err != nil {
		return "", fmt.Errorf("error loading template directory %s", err)
	}

	return cfg.Site + path, nil
}
