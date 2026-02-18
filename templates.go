// SPDX-License-Identifier: MIT

package main

import (
	"embed"
	"fmt"
	"html/template"
	"os"
)

//go:embed asset/templates/base/*.html asset/templates/bare/*.html asset/templates/dark/*.html asset/templates/light/*.html
var templateFS embed.FS

var validStyles = []string{"dark", "light", "bare"}

// resolveSearchTemplate returns the parsed search template. Custom file override
// (--template-search) takes priority, then the embedded theme (--template-style).
func resolveSearchTemplate(cfg *Config) (*template.Template, error) {
	if cfg.SearchTemplate != "" {
		data, err := os.ReadFile(cfg.SearchTemplate)
		if err != nil {
			return nil, fmt.Errorf("reading custom search template %q: %w", cfg.SearchTemplate, err)
		}
		return template.New("search").Parse(string(data))
	}

	if !isValidStyle(cfg.TemplateStyle) {
		return nil, fmt.Errorf("unknown template style %q (valid: dark, light, bare)", cfg.TemplateStyle)
	}

	basePath := "asset/templates/base/search.html"
	baseData, err := templateFS.ReadFile(basePath)
	if err != nil {
		return nil, fmt.Errorf("reading embedded base search template: %w", err)
	}

	path := fmt.Sprintf("asset/templates/%s/search.html", cfg.TemplateStyle)
	data, err := templateFS.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading embedded search template for style %q: %w", cfg.TemplateStyle, err)
	}

	tmpl, err := template.New("search").Parse(string(baseData))
	if err != nil {
		return nil, fmt.Errorf("parsing base search template: %w", err)
	}

	return tmpl.Parse(string(data))
}

// resolveDisplayTemplate returns the parsed display template. Custom file override
// (--template-display) takes priority, then the embedded theme (--template-style).
func resolveDisplayTemplate(cfg *Config) (*template.Template, error) {
	if cfg.DisplayTemplate != "" {
		data, err := os.ReadFile(cfg.DisplayTemplate)
		if err != nil {
			return nil, fmt.Errorf("reading custom display template %q: %w", cfg.DisplayTemplate, err)
		}
		return template.New("display").Parse(string(data))
	}

	if !isValidStyle(cfg.TemplateStyle) {
		return nil, fmt.Errorf("unknown template style %q (valid: dark, light, bare)", cfg.TemplateStyle)
	}

	path := fmt.Sprintf("asset/templates/%s/display.html", cfg.TemplateStyle)
	data, err := templateFS.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading embedded display template for style %q: %w", cfg.TemplateStyle, err)
	}
	return template.New("display").Parse(string(data))
}

func isValidStyle(style string) bool {
	for _, s := range validStyles {
		if s == style {
			return true
		}
	}
	return false
}
