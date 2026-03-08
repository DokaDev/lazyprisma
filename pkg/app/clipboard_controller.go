package app

import (
	"fmt"

	"github.com/dokadev/lazyprisma/pkg/gui/context"
	"github.com/dokadev/lazyprisma/pkg/gui/types"
	"github.com/jesseduffield/gocui"
)

// ClipboardController handles clipboard-related operations.
type ClipboardController struct {
	c             types.IControllerHost
	g             *gocui.Gui
	migrationsCtx *context.MigrationsContext
	openModal     func(Modal)
	closeModal    func()
}

// NewClipboardController creates a new ClipboardController.
func NewClipboardController(
	c types.IControllerHost,
	g *gocui.Gui,
	migrationsCtx *context.MigrationsContext,
	openModal func(Modal),
	closeModal func(),
) *ClipboardController {
	return &ClipboardController{
		c:             c,
		g:             g,
		migrationsCtx: migrationsCtx,
		openModal:     openModal,
		closeModal:    closeModal,
	}
}

// CopyMigrationInfo copies migration info to clipboard
func (cc *ClipboardController) CopyMigrationInfo() {
	tr := cc.c.GetTranslationSet()

	// Get selected migration
	selected := cc.migrationsCtx.GetSelectedMigration()
	if selected == nil {
		return
	}

	items := []ListModalItem{
		{
			Label:       tr.ListItemCopyName,
			Description: selected.Name,
			OnSelect: func() error {
				cc.closeModal()
				cc.copyTextToClipboard(selected.Name, tr.CopyLabelMigrationName)
				return nil
			},
		},
		{
			Label:       tr.ListItemCopyPath,
			Description: selected.Path,
			OnSelect: func() error {
				cc.closeModal()
				cc.copyTextToClipboard(selected.Path, tr.CopyLabelMigrationPath)
				return nil
			},
		},
	}

	// If it has a checksum, allow copying it
	if selected.Checksum != "" {
		items = append(items, ListModalItem{
			Label:       tr.ListItemCopyChecksum,
			Description: selected.Checksum,
			OnSelect: func() error {
				cc.closeModal()
				cc.copyTextToClipboard(selected.Checksum, tr.CopyLabelChecksum)
				return nil
			},
		})
	}

	modal := NewListModal(cc.g, tr, tr.ModalTitleCopyToClipboard, items,
		func() {
			cc.closeModal()
		},
	).WithStyle(MessageModalStyle{TitleColor: ColorCyan, BorderColor: ColorCyan})

	cc.openModal(modal)
}

func (cc *ClipboardController) copyTextToClipboard(text, label string) {
	tr := cc.c.GetTranslationSet()

	if err := CopyToClipboard(text); err != nil {
		modal := NewMessageModal(cc.g, tr, tr.ModalTitleClipboardError,
			tr.ModalMsgFailedCopyClipboard,
			err.Error(),
		).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		cc.openModal(modal)
		return
	}

	// Show toast/notification via modal for now
	modal := NewMessageModal(cc.g, tr, tr.ModalTitleCopied,
		fmt.Sprintf(tr.ModalMsgCopiedToClipboard, label),
	).WithStyle(MessageModalStyle{TitleColor: ColorGreen, BorderColor: ColorGreen})
	cc.openModal(modal)
}
