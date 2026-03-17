package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

func ValidateFile(filename string, fileSize int64, cfg *Config) error {
	ext := strings.ToLower(filepath.Ext(filename))
	if !cfg.AllowedTypes[ext] {
		allowed := make([]string, 0, len(cfg.AllowedTypes))
		for k := range cfg.AllowedTypes {
			allowed = append(allowed, k)
		}
		return &ValidationError{
			Code:    400,
			Message: fmt.Sprintf("Invalid file type. Allowed types: %s", strings.Join(allowed, ", ")),
		}
	}

	if fileSize > cfg.MaxFileSize {
		return &ValidationError{
			Code:    413,
			Message: "File too large. Maximum size is 100MB",
		}
	}

	return nil
}

type ValidationError struct {
	Code    int
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
