package app

import (
	"os"
	"time"

	"github.com/dokadev/lazyprisma/pkg/commands"
	"github.com/dokadev/lazyprisma/pkg/gui/context"
	"github.com/jesseduffield/gocui"
)

// Studio toggles Prisma Studio
func (a *App) Studio() {
	outputPanel, ok := a.panels[ViewOutputs].(*context.OutputContext)
	if !ok {
		return
	}

	// Check if Studio is already running
	if a.studioRunning {
		// Stop Studio
		if a.studioCmd != nil {
			if err := a.studioCmd.Kill(); err != nil {
				outputPanel.LogAction(a.Tr.LogActionStudio, a.Tr.ModalMsgFailedStopStudio+" "+err.Error())
				modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleStudioError,
					a.Tr.ModalMsgFailedStopStudio,
					err.Error(),
				).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
				a.OpenModal(modal)
				return
			}
			a.studioCmd = nil
		}
		a.studioRunning = false
		outputPanel.LogAction(a.Tr.LogActionStudioStopped, a.Tr.LogMsgStudioHasStopped)

		// Clear subtitle
		outputPanel.SetSubtitle("")

		// Update UI
		a.g.Update(func(g *gocui.Gui) error {
			// Trigger redraw of status bar
			return nil
		})

		modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleStudioStopped,
			a.Tr.ModalMsgStudioStopped,
		).WithStyle(MessageModalStyle{TitleColor: ColorYellow, BorderColor: ColorYellow})
		a.OpenModal(modal)
		return
	}

	// Start Studio
	// Try to start command - if another command is running, block
	if !a.tryStartCommand("Start Studio") {
		a.logCommandBlocked("Start Studio")
		return
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		a.finishCommand()
		outputPanel.LogAction(a.Tr.LogActionStudio, a.Tr.ErrorFailedGetWorkingDir+" "+err.Error())
		modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleStudioError,
			a.Tr.ErrorFailedGetWorkingDir,
			err.Error(),
		).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		a.OpenModal(modal)
		return
	}

	// Log action start
	outputPanel.LogAction(a.Tr.LogActionStudio, a.Tr.LogMsgStartingStudio)

	// Create command builder
	builder := commands.NewCommandBuilder(commands.NewPlatform())

	// Build prisma studio command
	// Note: We don't use StreamOutput here because Studio is a long-running process
	// and we want to capture the command object to kill it later
	studioCmd := builder.New("npx", "prisma", "studio").
		WithWorkingDir(cwd)

	// Start async
	if err := studioCmd.RunAsync(); err != nil {
		a.finishCommand()
		outputPanel.LogAction(a.Tr.LogActionStudio, a.Tr.ModalMsgFailedStartStudio+" "+err.Error())
		modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleStudioError,
			a.Tr.ModalMsgFailedStartStudio,
			err.Error(),
		).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		a.OpenModal(modal)
		return
	}

	// Wait a bit to ensure it started, then finish the "starting" command
	// The process continues running in background
	go func() {
		time.Sleep(2 * time.Second)
		a.g.Update(func(g *gocui.Gui) error {
			a.finishCommand() // Finish "starting" command
			a.studioRunning = true
			a.studioCmd = studioCmd // Save Command object

			outputPanel.LogAction(a.Tr.LogActionStudioStarted, a.Tr.LogMsgStudioListeningAt)
			outputPanel.SetSubtitle(a.Tr.LogMsgStudioListeningAt)

			// Show info modal
			modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleStudioStarted,
				a.Tr.ModalMsgStudioRunningAt,
				a.Tr.ModalMsgPressStopStudio,
			).WithStyle(MessageModalStyle{TitleColor: ColorGreen, BorderColor: ColorGreen})
			a.OpenModal(modal)
			return nil
		})
	}()
}
