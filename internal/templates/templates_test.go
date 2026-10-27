package templates_test

import (
	"file_sync/internal/templates"
	"strings"
	"testing"
)

func TestTemplates_ThemeLinearOnyx(t *testing.T) {
	data, err := templates.Assets.ReadFile("index.html")
	if err != nil {
		t.Fatalf("failed to read embedded index.html: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "bg-zinc-950") {
		t.Fatalf("expected Linear deep onyx theme (bg-zinc-950) in index.html, not found")
	}
	if !strings.Contains(content, "from-violet-600") {
		t.Fatalf("expected Linear violet gradient (from-violet-600) in index.html, not found")
	}

	uiData, err := templates.Assets.ReadFile("ui.js")
	if err != nil {
		t.Fatalf("failed to read embedded ui.js: %v", err)
	}
	uiContent := string(uiData)
	if !strings.Contains(uiContent, "bg-violet-600/20") {
		t.Fatalf("expected violet preset active badge in ui.js, not found")
	}
	if !strings.Contains(uiContent, "bg-violet-950/60") {
		t.Fatalf("expected violet target badge in ui.js, not found")
	}
}
