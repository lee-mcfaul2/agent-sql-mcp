package main

import (
	"testing"
)

func TestEmbeddedSchemaFiles_LoadsAll(t *testing.T) {
	files, err := embeddedSchemaFiles()
	if err != nil {
		t.Fatal(err)
	}
	// Fifteen files: meta+req+resp for five tools (meta.json added by the
	// lib-agent-prompt v1.0 schema propagation).
	if len(files) != 15 {
		t.Errorf("expected 15 files, got %d", len(files))
	}
}
