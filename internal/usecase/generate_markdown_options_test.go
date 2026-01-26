package usecase_test

import (
	"strings"
	"testing"

	"github.com/annenpolka/cclog/internal/usecase"
)

func TestGenerateMarkdownFromPath_Directory(t *testing.T) {
	dir := "../../testdata"
	md, err := usecase.GenerateMarkdownFromPath(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(md, "# Claude Conversation Logs") {
		t.Fatalf("expected multi-conversation header in output, got prefix: %q", md[:min(80, len(md))])
	}
}

func TestGenerateMarkdown_Options_ShowUUID(t *testing.T) {
	file := "../../testdata/sample.jsonl"

	md, err := usecase.GenerateMarkdownFromPath(file, usecase.Options{ShowUUID: true})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !strings.Contains(md, "UUID:") {
		t.Errorf("expected UUID to appear when ShowUUID=true")
	}

	md2, err := usecase.GenerateMarkdownFromPath(file, usecase.Options{ShowUUID: false})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if strings.Contains(md2, "UUID:") {
		t.Errorf("did not expect UUID to appear when ShowUUID=false")
	}
}

func TestGenerateMarkdown_Options_ShowTitle(t *testing.T) {
	file := "../../testdata/sample.jsonl"
	md, err := usecase.GenerateMarkdownFromPath(file, usecase.Options{ShowTitle: true})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !strings.HasPrefix(md, "# ") {
		t.Fatalf("expected markdown to start with a title, got: %q", md[:min(40, len(md))])
	}
	if !strings.Contains(md, "# Conversation Log\n\n") {
		t.Fatalf("expected underlying conversation header to still be present")
	}
}

func TestGenerateMarkdown_IncludeAll_AffectsFiltering(t *testing.T) {
	file := "../../testdata/sample.jsonl"

	// Include all: should include system messages and placeholders
	mdAll, err := usecase.GenerateMarkdownFromPath(file, usecase.Options{IncludeAll: true})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !strings.Contains(mdAll, "## System") {
		t.Errorf("expected system messages to appear when IncludeAll=true")
	}
	if !strings.Contains(mdAll, "*[Command executed:") {
		t.Errorf("expected placeholder for command execution when IncludeAll=true")
	}

	// Filtering enabled: system should be filtered out
	mdFiltered, err := usecase.GenerateMarkdownFromPath(file, usecase.Options{IncludeAll: false})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if strings.Contains(mdFiltered, "### System") {
		t.Errorf("did not expect system messages when IncludeAll=false")
	}
}

// helper (duplicate definition avoided by reusing from other _test file)
