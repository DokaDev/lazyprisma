package types

import (
	"github.com/dokadev/lazyprisma/pkg/i18n"
)

// ConfirmOpts configures a confirmation popup.
type ConfirmOpts struct {
	Title          string
	Prompt         string
	HandleConfirm  func() error
	HandleClose    func() error
}

// PromptOpts configures a text-input popup.
type PromptOpts struct {
	Title            string
	InitialContent   string
	HandleConfirm    func(string) error
	Required         bool
	Subtitle         string
	OnValidationFail func(string)
}

// MenuItem is a single entry in a menu popup.
type MenuItem struct {
	Label       string
	OnPress     func() error
	Description string
}

// MenuOpts configures a menu popup.
type MenuOpts struct {
	Title string
	Items []*MenuItem
}

// IPopupHandler provides methods for displaying popups to the user.
type IPopupHandler interface {
	// Alert shows a simple notification popup.
	Alert(title string, message string)
	// Confirm shows a yes/no confirmation popup.
	Confirm(opts ConfirmOpts)
	// Prompt shows a text-input popup.
	Prompt(opts PromptOpts)
	// Menu shows a list of selectable options.
	Menu(opts MenuOpts) error
	// Toast shows a brief, non-blocking message.
	Toast(message string)
	// ErrorHandler is the global error handler for gocui.
	ErrorHandler(err error) error
}

// IGuiCommon is the common interface available to controllers via dependency injection.
type IGuiCommon interface {
	IPopupHandler

	// LogAction logs a user-visible action to the output panel.
	LogAction(action string, detail ...string)
	// Refresh triggers a data refresh and re-render of all contexts.
	Refresh()
	// OnUIThread schedules a function to run on the UI thread.
	OnUIThread(f func() error)
	// GetTranslationSet returns the current translation set.
	GetTranslationSet() *i18n.TranslationSet
}

// IControllerHost is the interface controllers use to interact with the application.
// It extends IGuiCommon with command lifecycle and refresh methods.
type IControllerHost interface {
	IGuiCommon

	// Command lifecycle
	TryStartCommand(name string) bool
	LogCommandBlocked(name string)
	FinishCommand()

	// Full refresh with callbacks
	RefreshAll(onComplete ...func()) bool
	RefreshPanels()
}
