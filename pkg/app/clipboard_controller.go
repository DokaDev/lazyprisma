package app

import (
	"fmt"

	"github.com/dokadev/lazyprisma/pkg/gui/context"
)

// CopyMigrationInfo copies migration info to clipboard
func (a *App) CopyMigrationInfo() {
	// Get migrations panel
	migrationsPanel, ok := a.panels[ViewMigrations].(*context.MigrationsContext)
	if !ok {
		return
	}

	// Get selected migration
	selected := migrationsPanel.GetSelectedMigration()
	if selected == nil {
		return
	}

	items := []ListModalItem{
		{
			Label:       a.Tr.ListItemCopyName,
			Description: selected.Name,
			OnSelect: func() error {
				a.CloseModal()
				a.copyTextToClipboard(selected.Name, a.Tr.CopyLabelMigrationName)
				return nil
			},
		},
		{
			Label:       a.Tr.ListItemCopyPath,
			Description: selected.Path,
			OnSelect: func() error {
				a.CloseModal()
				a.copyTextToClipboard(selected.Path, a.Tr.CopyLabelMigrationPath)
				return nil
			},
		},
	}

	// If it has a checksum, allow copying it
	if selected.Checksum != "" {
		items = append(items, ListModalItem{
			Label:       a.Tr.ListItemCopyChecksum,
			Description: selected.Checksum,
			OnSelect: func() error {
				a.CloseModal()
				a.copyTextToClipboard(selected.Checksum, a.Tr.CopyLabelChecksum)
				return nil
			},
		})
	}

	modal := NewListModal(a.g, a.Tr, a.Tr.ModalTitleCopyToClipboard, items,
		func() {
			a.CloseModal()
		},
	).WithStyle(MessageModalStyle{TitleColor: ColorCyan, BorderColor: ColorCyan})

	a.OpenModal(modal)
}

func (a *App) copyTextToClipboard(text, label string) {
	if err := CopyToClipboard(text); err != nil {
		modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleClipboardError,
			a.Tr.ModalMsgFailedCopyClipboard,
			err.Error(),
		).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		a.OpenModal(modal)
		return
	}

	// Show toast/notification via modal for now
	// Ideally we would have a toast system
	modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleCopied,
		fmt.Sprintf(a.Tr.ModalMsgCopiedToClipboard, label),
	).WithStyle(MessageModalStyle{TitleColor: ColorGreen, BorderColor: ColorGreen})
	a.OpenModal(modal)
}
