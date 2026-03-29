package tui

import (
	"reflect"
	"testing"

	"github.com/joelazar/kagi/internal/config"
)

func TestBuildSearchArgs(t *testing.T) {
	got := buildSearchArgs(map[string]string{
		"query":           "golang generics",
		"limit":           "5",
		"include_content": "yes",
	})
	want := []string{"search", "golang generics", "-n", "5", "--content"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("buildSearchArgs() = %#v, want %#v", got, want)
	}
}

func TestBuildSummarizeArgsForTextMode(t *testing.T) {
	got := buildSummarizeArgs(map[string]string{
		"input":        "Long pasted text",
		"input_mode":   "text",
		"engine":       "muriel",
		"summary_type": "takeaway",
		"lang":         "DE",
	})
	want := []string{"summarize", "--text", "Long pasted text", "--engine", "muriel", "--type", "takeaway", "--lang", "DE"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("buildSummarizeArgs() = %#v, want %#v", got, want)
	}
}

func TestBuildEnrichArgsDefaultsToWeb(t *testing.T) {
	got := buildEnrichArgs(map[string]string{"query": "indie web"})
	want := []string{"enrich", "web", "indie web"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("buildEnrichArgs() = %#v, want %#v", got, want)
	}
}

func TestBuildAssistantArgsIncludesThread(t *testing.T) {
	got := buildAssistantArgs(map[string]string{
		"query":     "Tell me more",
		"thread_id": "thread-123",
	})
	want := []string{"assistant", "Tell me more", "--thread", "thread-123"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("buildAssistantArgs() = %#v, want %#v", got, want)
	}
}

func TestMergeConfigInputsPreservesBlankFields(t *testing.T) {
	current := config.Config{
		APIKey:       "api-old",
		SessionToken: "session-old",
		Defaults: config.Defaults{
			Format: "json",
			Search: config.SearchDefaults{Region: "us-en"},
		},
	}

	got := mergeConfigInputs(current, map[string]string{
		"default_format": "pretty",
	})

	want := config.Config{
		APIKey:       "api-old",
		SessionToken: "session-old",
		Defaults: config.Defaults{
			Format: "pretty",
			Search: config.SearchDefaults{Region: "us-en"},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("mergeConfigInputs() = %#v, want %#v", got, want)
	}
}

func TestMenuCommandsCoverInteractiveStories(t *testing.T) {
	commands := MenuCommands()
	seen := map[string]bool{}
	for _, cmd := range commands {
		seen[cmd.Name] = true
	}

	required := []string{
		"search",
		"search content",
		"fastgpt",
		"summarize",
		"enrich",
		"quick",
		"translate",
		"news",
		"smallweb",
		"askpage",
		"assistant",
		"assistant threads",
		"assistant delete thread",
		"balance",
		"auth",
		"config",
		"completion",
		"version",
	}

	for _, name := range required {
		if !seen[name] {
			t.Fatalf("expected menu command %q to exist", name)
		}
	}

	if seen["batch"] {
		t.Fatalf("batch should remain CLI-only and not appear in the TUI menu")
	}
}
