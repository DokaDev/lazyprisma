package app

import (
	"github.com/dokadev/lazyprisma/pkg/gui/context"
	"github.com/dokadev/lazyprisma/pkg/gui/types"
	"github.com/dokadev/lazyprisma/pkg/i18n"
	"github.com/jesseduffield/gocui"
)

// Compile-time interface satisfaction checks.
var _ types.IGuiCommon = (*App)(nil)

// --- IPopupHandler methods ---

// Alert shows a simple notification popup (unstyled).
func (a *App) Alert(title string, message string) {
	modal := NewMessageModal(a.g, a.Tr, title, message)
	a.OpenModal(modal)
}

// Confirm shows a yes/no confirmation popup.
// The ConfirmOpts callbacks return error; the underlying ConfirmModal expects func().
// We wrap them with adapters that discard the error.
func (a *App) Confirm(opts types.ConfirmOpts) {
	adaptedOnYes := func() {
		if opts.HandleConfirm != nil {
			_ = opts.HandleConfirm()
		}
	}
	adaptedOnNo := func() {
		if opts.HandleClose != nil {
			_ = opts.HandleClose()
		}
	}

	modal := NewConfirmModal(a.g, a.Tr, opts.Title, opts.Prompt, adaptedOnYes, adaptedOnNo)
	a.OpenModal(modal)
}

// Prompt shows a text-input popup.
// The PromptOpts callback returns error; the underlying InputModal expects func(string).
// We wrap with an adapter that discards the error.
func (a *App) Prompt(opts types.PromptOpts) {
	adaptedOnSubmit := func(input string) {
		if opts.HandleConfirm != nil {
			_ = opts.HandleConfirm(input)
		}
	}
	adaptedOnCancel := func() {
		a.CloseModal()
	}

	modal := NewInputModal(a.g, a.Tr, opts.Title, adaptedOnSubmit, adaptedOnCancel)

	if opts.Subtitle != "" {
		modal = modal.WithSubtitle(opts.Subtitle)
	}
	if opts.Required {
		modal = modal.WithRequired(true)
	}
	if opts.OnValidationFail != nil {
		modal = modal.OnValidationFail(opts.OnValidationFail)
	}

	a.OpenModal(modal)
}

// Menu shows a list of selectable options.
// Maps types.MenuItem to ListModalItem.
func (a *App) Menu(opts types.MenuOpts) error {
	items := make([]ListModalItem, len(opts.Items))
	for i, mi := range opts.Items {
		items[i] = ListModalItem{
			Label:       mi.Label,
			Description: mi.Description,
			OnSelect:    mi.OnPress,
		}
	}

	modal := NewListModal(a.g, a.Tr, opts.Title, items, func() {
		a.CloseModal()
	})
	a.OpenModal(modal)
	return nil
}

// Toast shows a brief message. Currently delegates to Alert as there is
// no auto-dismiss toast system yet.
func (a *App) Toast(message string) {
	a.Alert("", message)
}

// ErrorHandler shows an error modal with red styling.
func (a *App) ErrorHandler(err error) error {
	if err == nil {
		return nil
	}
	modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleError,
		err.Error(),
	).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
	a.OpenModal(modal)
	return nil
}

// --- IGuiCommon methods ---

// LogAction logs a user-visible action to the output panel.
// Wrapped in g.Update() for thread safety — OutputContext.LogAction mutates
// o.content without mutex, so it must run on the UI thread.
func (a *App) LogAction(action string, detail ...string) {
	a.g.Update(func(g *gocui.Gui) error {
		if outputPanel, ok := a.panels[ViewOutputs].(*context.OutputContext); ok {
			outputPanel.LogAction(action, detail...)
		}
		return nil
	})
}

// Refresh triggers a data refresh and re-render of all contexts.
// This is the controller-facing API; it ignores the return value of RefreshAll.
func (a *App) Refresh() {
	a.RefreshAll()
}

// OnUIThread schedules a function to run on the UI thread.
func (a *App) OnUIThread(f func() error) {
	a.g.Update(func(g *gocui.Gui) error {
		return f()
	})
}

// GetTranslationSet returns the current translation set.
func (a *App) GetTranslationSet() *i18n.TranslationSet {
	return a.Tr
}
