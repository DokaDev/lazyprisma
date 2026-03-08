package app

import (
	"github.com/dokadev/lazyprisma/pkg/gui/types"
	"github.com/jesseduffield/gocui"
)

func (a *App) RegisterKeybindings() error {
	// Quit or close modal (lowercase q)
	if err := a.g.SetKeybinding("", 'q', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if a.HasActiveModal() {
			// Modals that accept text input use 'q' for typing, not for closing
			if !a.activeModal.AcceptsTextInput() {
				a.CloseModal()
				return nil
			}
		} else {
			return gocui.ErrQuit
		}
		return nil
	}); err != nil {
		return err
	}

	// Ctrl+C to quit
	if err := a.g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return gocui.ErrQuit
	}); err != nil {
		return err
	}

	// ESC also closes modal
	if err := a.g.SetKeybinding("", gocui.KeyEsc, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if a.HasActiveModal() {
			a.CloseModal()
			return nil
		}
		return nil
	}); err != nil {
		return err
	}

	// Tab switching within panel (Tab)
	if err := a.g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		// Modal gets priority for key handling
		if a.HasActiveModal() {
			return a.activeModal.HandleKey(gocui.KeyTab, gocui.ModNone)
		}
		// Check if current panel supports tabs
		if panel := a.GetCurrentPanel(); panel != nil {
			if tabbedPanel, ok := panel.(types.ITabbedContext); ok {
				tabbedPanel.NextTab()
			}
		}
		return nil
	}); err != nil {
		return err
	}

	// Previous tab within panel (Shift+Tab)
	if err := a.g.SetKeybinding("", gocui.KeyBacktab, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if a.HasActiveModal() {
			return a.activeModal.HandleKey(gocui.KeyBacktab, gocui.ModNone)
		}
		// Check if current panel supports tabs
		if panel := a.GetCurrentPanel(); panel != nil {
			if tabbedPanel, ok := panel.(types.ITabbedContext); ok {
				tabbedPanel.PrevTab()
			}
		}
		return nil
	}); err != nil {
		return err
	}

	// Arrow keys for navigation
	if err := a.g.SetKeybinding("", gocui.KeyArrowRight, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if a.HasActiveModal() {
			return a.activeModal.HandleKey(gocui.KeyArrowRight, gocui.ModNone)
		}
		a.FocusNext()
		return nil
	}); err != nil {
		return err
	}

	if err := a.g.SetKeybinding("", gocui.KeyArrowLeft, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if a.HasActiveModal() {
			return a.activeModal.HandleKey(gocui.KeyArrowLeft, gocui.ModNone)
		}
		a.FocusPrevious()
		return nil
	}); err != nil {
		return err
	}

	// Arrow Up/Down for modal navigation, list navigation, or scrolling
	if err := a.g.SetKeybinding("", gocui.KeyArrowUp, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if a.HasActiveModal() {
			return a.activeModal.HandleKey(gocui.KeyArrowUp, gocui.ModNone)
		}
		// IListContext (e.g. MigrationsContext) takes priority over IScrollableContext
		if panel := a.GetCurrentPanel(); panel != nil {
			if listPanel, ok := panel.(types.IListContext); ok {
				listPanel.SelectPrev()
			} else if scrollPanel, ok := panel.(types.IScrollableContext); ok {
				scrollPanel.ScrollUp()
			}
		}
		return nil
	}); err != nil {
		return err
	}

	if err := a.g.SetKeybinding("", gocui.KeyArrowDown, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if a.HasActiveModal() {
			return a.activeModal.HandleKey(gocui.KeyArrowDown, gocui.ModNone)
		}
		// IListContext (e.g. MigrationsContext) takes priority over IScrollableContext
		if panel := a.GetCurrentPanel(); panel != nil {
			if listPanel, ok := panel.(types.IListContext); ok {
				listPanel.SelectNext()
			} else if scrollPanel, ok := panel.(types.IScrollableContext); ok {
				scrollPanel.ScrollDown()
			}
		}
		return nil
	}); err != nil {
		return err
	}

	// Enter key for modal
	if err := a.g.SetKeybinding("", gocui.KeyEnter, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if a.HasActiveModal() {
			// Modals that close on Enter (e.g. MessageModal) are dismissed directly
			if a.activeModal.ClosesOnEnter() {
				a.CloseModal()
				return nil
			}
			// Other modals: pass Enter to HandleKey (InputModal, ListModal, etc.)
			return a.activeModal.HandleKey(gocui.KeyEnter, gocui.ModNone)
		}
		return nil
	}); err != nil {
		return err
	}

	// Home key - scroll to top
	if err := a.g.SetKeybinding("", gocui.KeyHome, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if a.HasActiveModal() {
			return a.activeModal.HandleKey(gocui.KeyHome, gocui.ModNone)
		}
		if panel := a.GetCurrentPanel(); panel != nil {
			if scrollPanel, ok := panel.(types.IScrollableContext); ok {
				scrollPanel.ScrollToTop()
			}
		}
		return nil
	}); err != nil {
		return err
	}

	// End key - scroll to bottom
	if err := a.g.SetKeybinding("", gocui.KeyEnd, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if a.HasActiveModal() {
			return a.activeModal.HandleKey(gocui.KeyEnd, gocui.ModNone)
		}
		if panel := a.GetCurrentPanel(); panel != nil {
			if scrollPanel, ok := panel.(types.IScrollableContext); ok {
				scrollPanel.ScrollToBottom()
			}
		}
		return nil
	}); err != nil {
		return err
	}

	// 'r' key - refresh all panels
	if err := a.g.SetKeybinding("", 'r', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if a.HasActiveModal() {
			return nil
		}
		a.RefreshAll()
		return nil
	}); err != nil {
		return err
	}

	// 'd' key - migrate dev
	if err := a.g.SetKeybinding("", 'd', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if a.HasActiveModal() {
			return nil
		}
		a.migrationsController.MigrateDev()
		return nil
	}); err != nil {
		return err
	}

	// 'D' key - migrate deploy
	if err := a.g.SetKeybinding("", 'D', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if a.HasActiveModal() {
			return nil
		}
		a.migrationsController.MigrateDeploy()
		return nil
	}); err != nil {
		return err
	}

	// 'g' key - run prisma generate
	if err := a.g.SetKeybinding("", 'g', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if a.HasActiveModal() {
			return nil
		}
		a.generateController.Generate()
		return nil
	}); err != nil {
		return err
	}

	// 's' key - migrate resolve
	if err := a.g.SetKeybinding("", 's', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if a.HasActiveModal() {
			return nil
		}
		a.migrationsController.MigrateResolve()
		return nil
	}); err != nil {
		return err
	}

	// 'S' key - toggle prisma studio
	if err := a.g.SetKeybinding("", 'S', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if a.HasActiveModal() {
			return nil
		}
		a.studioController.Studio()
		return nil
	}); err != nil {
		return err
	}

	// 'c' key - copy migration info
	if err := a.g.SetKeybinding("", 'c', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if a.HasActiveModal() {
			return nil
		}
		a.clipboardController.CopyMigrationInfo()
		return nil
	}); err != nil {
		return err
	}

	// Delete key - delete pending migration
	deleteHandler := func(g *gocui.Gui, v *gocui.View) error {
		if a.HasActiveModal() {
			return nil
		}
		a.migrationsController.DeleteMigration()
		return nil
	}

	if err := a.g.SetKeybinding("", gocui.KeyDelete, gocui.ModNone, deleteHandler); err != nil {
		return err
	}
	if err := a.g.SetKeybinding("", gocui.KeyBackspace, gocui.ModNone, deleteHandler); err != nil {
		return err
	}
	if err := a.g.SetKeybinding("", gocui.KeyBackspace2, gocui.ModNone, deleteHandler); err != nil {
		return err
	}

	// 'y' key - pass to ConfirmModal for Yes
	if err := a.g.SetKeybinding("", 'y', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if a.HasActiveModal() {
			// Pass 'y' to modal (for ConfirmModal)
			return a.activeModal.HandleKey('y', gocui.ModNone)
		}
		return nil
	}); err != nil {
		return err
	}

	// 'n' key - pass to ConfirmModal for No
	if err := a.g.SetKeybinding("", 'n', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if a.HasActiveModal() {
			// Pass 'n' to modal (for ConfirmModal)
			return a.activeModal.HandleKey('n', gocui.ModNone)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}
