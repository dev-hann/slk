// internal/ui/mode_theme_switcher.go
//
// Theme-switcher mode key handler.
//
// Live-preview semantics (Pattern B — preview-then-commit, à la Neovim
// Telescope theme pickers):
//
//   - On open, themeswitcher.Model snapshots styles.CurrentName() as the
//     "original" theme (see OriginalName).
//   - Arrow / filter navigation applies the highlighted theme immediately via
//     applyTheme, so the app behind the dimmed overlay re-colors in real
//     time. No disk write happens here.
//   - Enter commits: applyTheme + themeSaveFn (persists to config). This is
//     the only path that writes.
//   - Escape (or any dismissal that closes the picker without a result)
//     reverts to OriginalName via applyTheme — previewed-but-unsaved changes
//     are discarded so the user lands back where they started.
//
// applyTheme is the single funnel for styles.Apply + cache invalidation +
// compose style refresh, so live-preview, commit, and revert all stay in
// sync.
package ui

import (
	tea "charm.land/bubbletea/v2"

	"github.com/gammons/slk/internal/ui/styles"
)

// applyTheme applies the named theme and refreshes every component whose
// rendering caches theme colors. Used by the theme-switcher handler for
// live preview, commit, and revert — all three need identical refresh
// work, so it lives here once.
func applyTheme(a *App, name string) {
	styles.Apply(name, a.themeOverrides)
	// Invalidate render caches so they rebuild with new theme colors.
	a.invalidateAllWinModelCaches()
	a.threadPanel.InvalidateCache()
	a.sidebar.InvalidateCache()
	// Refresh compose textarea styles for new theme.
	a.compose.RefreshStyles()
	a.threadCompose.RefreshStyles()
}

func handleThemeSwitcherMode(a *App, msg tea.KeyMsg) tea.Cmd {
	keyStr := msg.String()
	switch msg.Key().Code {
	case tea.KeyEnter:
		keyStr = "enter"
	case tea.KeyEscape:
		keyStr = "esc"
	case tea.KeyUp:
		keyStr = "up"
	case tea.KeyDown:
		keyStr = "down"
	case tea.KeyBackspace:
		keyStr = "backspace"
	}

	wasVisible := a.themeSwitcher.IsVisible()
	result := a.themeSwitcher.HandleKey(keyStr)

	// Enter -> commit: apply the chosen theme and persist it.
	if result != nil {
		a.themeSwitcher.Close()
		a.SetMode(ModeNormal)
		applyTheme(a, result.Name)
		if a.themeSaveFn != nil {
			a.themeSaveFn(result.Name, result.Scope)
		}
		return nil
	}

	// Picker closed without a selection (Escape). If we applied any
	// previews, revert to the theme that was active when the picker opened
	// so the user isn't left looking at a theme they didn't commit.
	if wasVisible && !a.themeSwitcher.IsVisible() {
		if orig := a.themeSwitcher.OriginalName(); orig != "" && orig != styles.CurrentName() {
			applyTheme(a, orig)
		}
		a.SetMode(ModeNormal)
		return nil
	}

	// Navigation / filter change -> live preview. Only re-apply when the
	// highlighted theme actually differs from the currently-applied one,
	// so identical keys (or keys that don't move the selection) are no-ops.
	if sel := a.themeSwitcher.SelectedName(); sel != "" && sel != styles.CurrentName() {
		applyTheme(a, sel)
	}
	return nil
}
