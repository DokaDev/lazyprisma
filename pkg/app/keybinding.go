package app

import "github.com/jesseduffield/gocui"

func (a *App) RegisterKeybindings() error {
	// Quit or close modal (lowercase q)
	if err := a.g.SetKeybinding("", 'q', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if a.HasActiveModal() {
			// InputModal uses 'q' for text input, not for closing
			if _, ok := a.activeModal.(*InputModal); !ok {
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

	// Quit or close modal (uppercase Q)
	// if err := a.g.SetKeybinding("", 'Q', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
	// 	if a.HasActiveModal() {
	// 		a.CloseModal()
	// 		return nil
	// 	}
	// 	return gocui.ErrQuit
	// }); err != nil {
	// 	return err
	// }

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
			if migrationsPanel, ok := panel.(*MigrationsPanel); ok {
				migrationsPanel.NextTab()
			} else if detailsPanel, ok := panel.(*DetailsPanel); ok {
				detailsPanel.NextTab()
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
			if migrationsPanel, ok := panel.(*MigrationsPanel); ok {
				migrationsPanel.PrevTab()
			} else if detailsPanel, ok := panel.(*DetailsPanel); ok {
				detailsPanel.PrevTab()
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
		// Handle different panel types
		if panel := a.GetCurrentPanel(); panel != nil {
			switch p := panel.(type) {
			case *MigrationsPanel:
				p.SelectPrev()
			case *WorkspacePanel:
				p.ScrollUp()
			case *DetailsPanel:
				p.ScrollUp()
			case *OutputPanel:
				p.ScrollUp()
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
		// Handle different panel types
		if panel := a.GetCurrentPanel(); panel != nil {
			switch p := panel.(type) {
			case *MigrationsPanel:
				p.SelectNext()
			case *WorkspacePanel:
				p.ScrollDown()
			case *DetailsPanel:
				p.ScrollDown()
			case *OutputPanel:
				p.ScrollDown()
			}
		}
		return nil
	}); err != nil {
		return err
	}

	// Enter key for modal
	if err := a.g.SetKeybinding("", gocui.KeyEnter, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if a.HasActiveModal() {
			// MessageModal: close on Enter
			if _, ok := a.activeModal.(*MessageModal); ok {
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
		// Handle different panel types
		if panel := a.GetCurrentPanel(); panel != nil {
			switch p := panel.(type) {
			case *MigrationsPanel:
				p.ScrollToTop()
			case *WorkspacePanel:
				p.ScrollToTop()
			case *DetailsPanel:
				p.ScrollToTop()
			case *OutputPanel:
				p.ScrollToTop()
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
		// Handle different panel types
		if panel := a.GetCurrentPanel(); panel != nil {
			switch p := panel.(type) {
			case *MigrationsPanel:
				p.ScrollToBottom()
			case *WorkspacePanel:
				p.ScrollToBottom()
			case *DetailsPanel:
				p.ScrollToBottom()
			case *OutputPanel:
				p.ScrollToBottom()
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

	// 'i' key - test ping to google.com
	if err := a.g.SetKeybinding("", 'i', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if a.HasActiveModal() {
			return nil
		}
		a.TestPing()
		return nil
	}); err != nil {
		return err
	}

	// // 't' key - test modal (temporary)
	// if err := a.g.SetKeybinding("", 't', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
	// 	if a.HasActiveModal() {
	// 		return nil
	// 	}
	// 	a.TestModal()
	// 	return nil
	// }); err != nil {
	// 	return err
	// }

	// // 'm' key - test input modal (temporary)
	// if err := a.g.SetKeybinding("", 'm', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
	// 	if a.HasActiveModal() {
	// 		return nil
	// 	}
	// 	a.TestInputModal()
	// 	return nil
	// }); err != nil {
	// 	return err
	// }

	// 'd' key - migrate dev
	if err := a.g.SetKeybinding("", 'd', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if a.HasActiveModal() {
			return nil
		}
		a.MigrateDev()
		return nil
	}); err != nil {
		return err
	}

	// 'D' key - migrate deploy
	if err := a.g.SetKeybinding("", 'D', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if a.HasActiveModal() {
			return nil
		}
		a.MigrateDeploy()
		return nil
	}); err != nil {
		return err
	}

	// 'g' key - run prisma generate
	if err := a.g.SetKeybinding("", 'g', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if a.HasActiveModal() {
			return nil
		}
		a.Generate()
		return nil
	}); err != nil {
		return err
	}

	// 's' key - migrate resolve
	if err := a.g.SetKeybinding("", 's', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if a.HasActiveModal() {
			return nil
		}
		a.MigrateResolve()
		return nil
	}); err != nil {
		return err
	}

	// 'S' key - toggle prisma studio
	if err := a.g.SetKeybinding("", 'S', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if a.HasActiveModal() {
			return nil
		}
		a.Studio()
		return nil
	}); err != nil {
		return err
	}

	// 'c' key - copy migration info
	if err := a.g.SetKeybinding("", 'c', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if a.HasActiveModal() {
			return nil
		}
		a.CopyMigrationInfo()
		return nil
	}); err != nil {
		return err
	}

	// Delete key - delete pending migration
	deleteHandler := func(g *gocui.Gui, v *gocui.View) error {
		if a.HasActiveModal() {
			return nil
		}
		a.DeleteMigration()
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

	// // 'l' key - test list modal (temporary)
	// if err := a.g.SetKeybinding("", 'l', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
	// 	if a.HasActiveModal() {
	// 		return nil
	// 	}
	// 	a.TestListModal()
	// 	return nil
	// }); err != nil {
	// 	return err
	// }

	// // 'y' key - test confirm modal (temporary)
	// if err := a.g.SetKeybinding("", 'y', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
	// 	if a.HasActiveModal() {
	// 		// Pass 'y' to modal (for ConfirmModal)
	// 		return a.activeModal.HandleKey('y', gocui.ModNone)
	// 	}
	// 	a.TestConfirmModal()
	// 	return nil
	// }); err != nil {
	// 	return err
	// }

	return nil
}
