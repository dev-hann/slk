package themeswitcher

import (
	"testing"

	"github.com/gammons/slk/internal/config"
	"github.com/gammons/slk/internal/ui/styles"
)

// stylesCfg is a zero-value override used so Apply tests don't mutate user
// overrides; restored to "dark" at the end of the styles-touching tests.
var stylesCfg config.Theme

func TestSelectedName_FirstItemOnOpen(t *testing.T) {
	styles.Apply("dark", stylesCfg)
	m := New()
	m.SetItems([]string{"Dark", "Light", "Dracula"})
	m.Open()
	if got := m.SelectedName(); got != "Dark" {
		t.Fatalf("SelectedName() = %q, want Dark on fresh open", got)
	}
}

func TestSelectedName_AfterNavigation(t *testing.T) {
	m := New()
	m.SetItems([]string{"Dark", "Light", "Dracula"})
	m.Open()
	m.HandleKey("down")
	if got := m.SelectedName(); got != "Light" {
		t.Fatalf("SelectedName() after down = %q, want Light", got)
	}
	m.HandleKey("down")
	if got := m.SelectedName(); got != "Dracula" {
		t.Fatalf("SelectedName() after 2x down = %q, want Dracula", got)
	}
	m.HandleKey("up")
	if got := m.SelectedName(); got != "Light" {
		t.Fatalf("SelectedName() after up = %q, want Light", got)
	}
}

func TestSelectedName_AfterFilterShowsFirstMatch(t *testing.T) {
	m := New()
	m.SetItems([]string{"Dark", "Light", "Dracula", "Nord"})
	m.Open()
	m.HandleKey("d")
	if got := m.SelectedName(); got != "Dark" {
		t.Fatalf("SelectedName() after filter 'd' = %q, want Dark (prefix match first)", got)
	}
}

func TestSelectedName_EmptyWhenNoMatches(t *testing.T) {
	m := New()
	m.SetItems([]string{"Dark", "Light"})
	m.Open()
	m.HandleKey("z")
	if got := m.SelectedName(); got != "" {
		t.Fatalf("SelectedName() with no matches = %q, want empty", got)
	}
}

func TestSelectedName_EmptyWhenClosed(t *testing.T) {
	m := New()
	m.SetItems([]string{"Dark"})
	if got := m.SelectedName(); got != "" {
		t.Fatalf("SelectedName() before Open = %q, want empty", got)
	}
}

func TestOriginalName_SnapshotsCurrentAtOpen(t *testing.T) {
	styles.Apply("dracula", stylesCfg)
	m := New()
	m.SetItems([]string{"Dark", "Light", "Dracula"})
	m.OpenWithScope(ScopeWorkspace, "")
	if got := m.OriginalName(); got != "dracula" {
		t.Fatalf("OriginalName() = %q, want dracula (snapshot of styles.CurrentName at open)", got)
	}
}

func TestOriginalName_UnchangedByNavigation(t *testing.T) {
	styles.Apply("dracula", stylesCfg)
	m := New()
	m.SetItems([]string{"Dark", "Light", "Dracula"})
	m.OpenWithScope(ScopeWorkspace, "")
	m.HandleKey("down")
	m.HandleKey("down")
	if got := m.OriginalName(); got != "dracula" {
		t.Fatalf("OriginalName() after navigation = %q, want dracula (must stay frozen)", got)
	}
}

func TestOpenClose(t *testing.T) {
	m := New()
	m.SetItems([]string{"Dark", "Light", "Dracula"})
	if m.IsVisible() {
		t.Error("should not be visible initially")
	}
	m.Open()
	if !m.IsVisible() {
		t.Error("should be visible after Open")
	}
	m.Close()
	if m.IsVisible() {
		t.Error("should not be visible after Close")
	}
}

func TestSelectTheme(t *testing.T) {
	m := New()
	m.SetItems([]string{"Dark", "Light", "Dracula"})
	m.Open()
	result := m.HandleKey("enter")
	if result == nil || result.Name != "Dark" {
		t.Errorf("expected Dark, got %v", result)
	}
}

func TestNavigation(t *testing.T) {
	m := New()
	m.SetItems([]string{"Dark", "Light", "Dracula"})
	m.Open()
	m.HandleKey("down")
	result := m.HandleKey("enter")
	if result == nil || result.Name != "Light" {
		t.Errorf("expected Light after down, got %v", result)
	}
}

func TestFilter(t *testing.T) {
	m := New()
	m.SetItems([]string{"Dark", "Light", "Dracula", "Nord"})
	m.Open()
	m.HandleKey("d")
	// Should match "Dark" and "Dracula" (prefix first)
	result := m.HandleKey("enter")
	if result == nil || result.Name != "Dark" {
		t.Errorf("expected Dark (prefix match first), got %v", result)
	}
}

func TestEscapeCloses(t *testing.T) {
	m := New()
	m.SetItems([]string{"Dark"})
	m.Open()
	result := m.HandleKey("esc")
	if result != nil {
		t.Error("expected nil result on escape")
	}
	if m.IsVisible() {
		t.Error("should be closed after escape")
	}
}

func TestOpenWithScope(t *testing.T) {
	m := New()
	m.SetItems([]string{"Dark", "Light"})

	m.OpenWithScope(ScopeWorkspace, "Theme for ACME")
	if !m.IsVisible() {
		t.Error("expected picker to be visible")
	}
	if m.Scope() != ScopeWorkspace {
		t.Errorf("Scope = %v, want ScopeWorkspace", m.Scope())
	}
	if m.HeaderText() != "Theme for ACME" {
		t.Errorf("HeaderText = %q, want Theme for ACME", m.HeaderText())
	}

	m.Close()
	m.OpenWithScope(ScopeGlobal, "Default theme")
	if m.Scope() != ScopeGlobal {
		t.Errorf("Scope after re-open = %v, want ScopeGlobal", m.Scope())
	}
	if m.HeaderText() != "Default theme" {
		t.Errorf("HeaderText = %q, want Default theme", m.HeaderText())
	}
}

func TestSelectionReturnsScope(t *testing.T) {
	m := New()
	m.SetItems([]string{"Dark"})
	m.OpenWithScope(ScopeWorkspace, "Theme for X")

	result := m.HandleKey("enter")
	if result == nil {
		t.Fatal("expected ThemeResult, got nil")
	}
	if result.Name != "Dark" {
		t.Errorf("result.Name = %q, want Dark", result.Name)
	}
	if result.Scope != ScopeWorkspace {
		t.Errorf("result.Scope = %v, want ScopeWorkspace", result.Scope)
	}
}

func TestLegacyOpenStillWorks(t *testing.T) {
	// Open() (no args) should default to ScopeGlobal with no header.
	m := New()
	m.SetItems([]string{"Dark"})
	m.Open()
	if m.Scope() != ScopeGlobal {
		t.Errorf("Open() default scope = %v, want ScopeGlobal", m.Scope())
	}
}
