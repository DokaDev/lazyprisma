package app

import (
	"os"
	"sync/atomic"
	"time"

	"github.com/dokadev/lazyprisma/pkg/commands"
	"github.com/dokadev/lazyprisma/pkg/gui/context"
	"github.com/dokadev/lazyprisma/pkg/gui/types"
	"github.com/jesseduffield/gocui"
)

// StudioController handles Prisma Studio toggle operations.
type StudioController struct {
	c             types.IControllerHost
	g             *gocui.Gui
	outputCtx     *context.OutputContext
	openModal     func(Modal)
	studioCmd     *commands.Command // Running studio command
	studioRunning atomic.Bool       // True if studio is running
}

// NewStudioController creates a new StudioController.
func NewStudioController(
	c types.IControllerHost,
	g *gocui.Gui,
	outputCtx *context.OutputContext,
	openModal func(Modal),
) *StudioController {
	return &StudioController{
		c:         c,
		g:         g,
		outputCtx: outputCtx,
		openModal: openModal,
	}
}

// IsStudioRunning returns whether Prisma Studio is currently running.
func (sc *StudioController) IsStudioRunning() bool {
	return sc.studioRunning.Load()
}

// GetStudioCmd returns the running studio command (for cleanup on app exit).
func (sc *StudioController) GetStudioCmd() *commands.Command {
	return sc.studioCmd
}

// Studio toggles Prisma Studio
func (sc *StudioController) Studio() {
	tr := sc.c.GetTranslationSet()

	// Check if Studio is already running
	if sc.studioRunning.Load() {
		// Stop Studio
		if sc.studioCmd != nil {
			if err := sc.studioCmd.Kill(); err != nil {
				sc.outputCtx.LogAction(tr.LogActionStudio, tr.ModalMsgFailedStopStudio+" "+err.Error())
				modal := NewMessageModal(sc.g, tr, tr.ModalTitleStudioError,
					tr.ModalMsgFailedStopStudio,
					err.Error(),
				).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
				sc.openModal(modal)
				return
			}
			sc.studioCmd = nil
		}
		sc.studioRunning.Store(false)
		sc.outputCtx.LogAction(tr.LogActionStudioStopped, tr.LogMsgStudioHasStopped)

		// Clear subtitle
		sc.outputCtx.SetSubtitle("")

		// Update UI
		sc.c.OnUIThread(func() error {
			// Trigger redraw of status bar
			return nil
		})

		modal := NewMessageModal(sc.g, tr, tr.ModalTitleStudioStopped,
			tr.ModalMsgStudioStopped,
		).WithStyle(MessageModalStyle{TitleColor: ColorYellow, BorderColor: ColorYellow})
		sc.openModal(modal)
		return
	}

	// Start Studio
	// Try to start command - if another command is running, block
	if !sc.c.TryStartCommand("Start Studio") {
		sc.c.LogCommandBlocked("Start Studio")
		return
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		sc.c.FinishCommand()
		sc.outputCtx.LogAction(tr.LogActionStudio, tr.ErrorFailedGetWorkingDir+" "+err.Error())
		modal := NewMessageModal(sc.g, tr, tr.ModalTitleStudioError,
			tr.ErrorFailedGetWorkingDir,
			err.Error(),
		).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		sc.openModal(modal)
		return
	}

	// Log action start
	sc.outputCtx.LogAction(tr.LogActionStudio, tr.LogMsgStartingStudio)

	// Create command builder
	builder := commands.NewCommandBuilder(commands.NewPlatform())

	// Build prisma studio command
	studioCmd := builder.New("npx", "prisma", "studio").
		WithWorkingDir(cwd)

	// Start async
	if err := studioCmd.RunAsync(); err != nil {
		sc.c.FinishCommand()
		sc.outputCtx.LogAction(tr.LogActionStudio, tr.ModalMsgFailedStartStudio+" "+err.Error())
		modal := NewMessageModal(sc.g, tr, tr.ModalTitleStudioError,
			tr.ModalMsgFailedStartStudio,
			err.Error(),
		).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		sc.openModal(modal)
		return
	}

	// Mark studio as running immediately to prevent double-start
	sc.studioRunning.Store(true)
	sc.studioCmd = studioCmd

	// Wait a bit to ensure it started, then finish the "starting" command
	go func() {
		time.Sleep(2 * time.Second)
		sc.c.OnUIThread(func() error {
			sc.c.FinishCommand() // Finish "starting" command

			sc.outputCtx.LogAction(tr.LogActionStudioStarted, tr.LogMsgStudioListeningAt)
			sc.outputCtx.SetSubtitle(tr.LogMsgStudioListeningAt)

			// Show info modal
			modal := NewMessageModal(sc.g, tr, tr.ModalTitleStudioStarted,
				tr.ModalMsgStudioRunningAt,
				tr.ModalMsgPressStopStudio,
			).WithStyle(MessageModalStyle{TitleColor: ColorGreen, BorderColor: ColorGreen})
			sc.openModal(modal)
			return nil
		})
	}()
}
