package ui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/gammons/slk/internal/ui/workspace"
)

func gKey() tea.KeyPressMsg { return tea.KeyPressMsg{Code: 'g', Text: "g"} }

func railWithItems(n int) []workspace.WorkspaceItem {
	items := make([]workspace.WorkspaceItem, n)
	for i := range items {
		items[i] = workspace.WorkspaceItem{
			ID:       teamID(i),
			Initials: "WS",
		}
	}
	return items
}

func teamID(i int) string {
	return "T" + string(rune('1'+i))
}

func newAppForGotoTest() *App {
	a := NewApp()
	_, _ = a.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	return a
}

// handleGoToTop on the workspace rail jumps selection to the first item.
func TestHandleGoToTop_WorkspaceRail(t *testing.T) {
	a := newAppForGotoTest()
	a.workspaceRail.SetItems(railWithItems(3))
	a.workspaceRail.Select(2) // start at the bottom
	a.focusedPanel = PanelWorkspace

	a.handleGoToTop()

	if got := a.workspaceRail.SelectedIndex(); got != 0 {
		t.Errorf("GoToTop on rail: SelectedIndex=%d want 0", got)
	}
}

// handleGoToBottom on the workspace rail jumps selection to the last item.
func TestHandleGoToBottom_WorkspaceRail(t *testing.T) {
	a := newAppForGotoTest()
	a.workspaceRail.SetItems(railWithItems(3))
	a.workspaceRail.Select(0) // start at the top
	a.focusedPanel = PanelWorkspace

	a.handleGoToBottom()

	if got := a.workspaceRail.SelectedIndex(); got != 2 {
		t.Errorf("GoToBottom on rail: SelectedIndex=%d want 2", got)
	}
}

// handleGoToTop is a no-op when the messages pane is focused (channel chat
// and the thread list are excluded from gg): the rail must not move.
func TestHandleGoToTop_MessagesPaneIsNoOp(t *testing.T) {
	a := newAppForGotoTest()
	a.workspaceRail.SetItems(railWithItems(3))
	a.workspaceRail.Select(1)
	a.focusedPanel = PanelMessages

	cmd := a.handleGoToTop()

	if cmd != nil {
		t.Errorf("GoToTop on messages pane: expected nil cmd, got %v", cmd)
	}
	if got := a.workspaceRail.SelectedIndex(); got != 1 {
		t.Errorf("GoToTop on messages pane moved rail: SelectedIndex=%d want 1", got)
	}
}

// A single 'g' arms the pending state but must not jump.
func TestGG_SingleGDoesNotJump(t *testing.T) {
	a := newAppForGotoTest()
	a.workspaceRail.SetItems(railWithItems(3))
	a.workspaceRail.Select(2)
	a.focusedPanel = PanelWorkspace
	a.SetMode(ModeNormal)

	handleNormalMode(a, gKey())

	if got := a.workspaceRail.SelectedIndex(); got != 2 {
		t.Errorf("single g moved rail: SelectedIndex=%d want 2", got)
	}
}

// Two consecutive 'g's jump the focused panel to the top.
func TestGG_DoubleGJumpsToTop(t *testing.T) {
	a := newAppForGotoTest()
	a.workspaceRail.SetItems(railWithItems(3))
	a.workspaceRail.Select(2)
	a.focusedPanel = PanelWorkspace
	a.SetMode(ModeNormal)

	handleNormalMode(a, gKey())
	handleNormalMode(a, gKey())

	if got := a.workspaceRail.SelectedIndex(); got != 0 {
		t.Errorf("gg: rail SelectedIndex=%d want 0", got)
	}
}

// A non-'g' key after the first 'g' cancels the pending state, so a
// later lone 'g' does not jump.
func TestGG_NonGCancelsPending(t *testing.T) {
	a := newAppForGotoTest()
	a.workspaceRail.SetItems(railWithItems(3))
	a.workspaceRail.Select(2)
	a.focusedPanel = PanelWorkspace
	a.SetMode(ModeNormal)

	handleNormalMode(a, gKey())
	// 'x' is unbound in normal mode; it should cancel pending and do nothing.
	handleNormalMode(a, tea.KeyPressMsg{Code: 'x', Text: "x"})
	handleNormalMode(a, gKey())

	if got := a.workspaceRail.SelectedIndex(); got != 2 {
		t.Errorf("g x g: rail SelectedIndex=%d want 2 (pending cancelled)", got)
	}
}

// gg is inert when the messages pane is focused (channel chat / thread
// list are excluded from gg).
func TestGG_MessagesPaneInert(t *testing.T) {
	a := newAppForGotoTest()
	a.workspaceRail.SetItems(railWithItems(3))
	a.workspaceRail.Select(1)
	a.focusedPanel = PanelMessages
	a.SetMode(ModeNormal)

	handleNormalMode(a, gKey())
	handleNormalMode(a, gKey())

	if got := a.workspaceRail.SelectedIndex(); got != 1 {
		t.Errorf("gg on messages moved rail: SelectedIndex=%d want 1", got)
	}
}
