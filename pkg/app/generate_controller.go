package app

import (
	"fmt"
	"os"
	"strings"

	"github.com/dokadev/lazyprisma/pkg/commands"
	"github.com/dokadev/lazyprisma/pkg/gui/context"
	"github.com/dokadev/lazyprisma/pkg/prisma"
	"github.com/jesseduffield/gocui"
)

// Generate runs prisma generate and shows result in modal
func (a *App) Generate() {
	// Try to start command - if another command is running, block
	if !a.tryStartCommand("Generate") {
		a.logCommandBlocked("Generate")
		return
	}

	outputPanel, ok := a.panels[ViewOutputs].(*context.OutputContext)
	if !ok {
		a.finishCommand() // Clean up if panel not found
		return
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		a.finishCommand()
		outputPanel.LogAction(a.Tr.LogActionGenerateError, a.Tr.ErrorFailedGetWorkingDir+" "+err.Error())
		modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleGenerateError,
			a.Tr.ErrorFailedGetWorkingDir,
			err.Error(),
		).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		a.OpenModal(modal)
		return
	}

	// Log action start
	outputPanel.LogAction(a.Tr.LogActionGenerate, a.Tr.LogMsgRunningGenerate)

	// Create command builder
	builder := commands.NewCommandBuilder(commands.NewPlatform())

	// Build prisma generate command
	generateCmd := builder.New("npx", "prisma", "generate").
		WithWorkingDir(cwd).
		StreamOutput().
		OnStdout(func(line string) {
			// Update UI on main thread
			a.g.Update(func(g *gocui.Gui) error {
				if out, ok := a.panels[ViewOutputs].(*context.OutputContext); ok {
					out.AppendOutput("  " + line)
				}
				return nil
			})
		}).
		OnStderr(func(line string) {
			// Update UI on main thread
			a.g.Update(func(g *gocui.Gui) error {
				if out, ok := a.panels[ViewOutputs].(*context.OutputContext); ok {
					out.AppendOutput("  " + line)
				}
				return nil
			})
		}).
		OnComplete(func(exitCode int) {
			// Update UI on main thread
			a.g.Update(func(g *gocui.Gui) error {
				if out, ok := a.panels[ViewOutputs].(*context.OutputContext); ok {
					if exitCode == 0 {
						a.finishCommand() // Finish immediately on success
						out.LogAction(a.Tr.LogActionGenerateComplete, a.Tr.LogMsgPrismaClientGeneratedSuccess)
						// Show success modal
						modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleGenerateSuccess,
							a.Tr.ModalMsgPrismaClientGenerated,
						).WithStyle(MessageModalStyle{TitleColor: ColorGreen, BorderColor: ColorGreen})
						a.OpenModal(modal)
					} else {
						// Failed - run validate to check schema (keep spinner running)
						out.LogAction(a.Tr.LogActionGenerateFailed, a.Tr.LogMsgCheckingSchemaErrors)

						// Run validate in goroutine to not block UI updates
						go func() {
							validateResult, err := prisma.Validate(cwd)

							// Update UI on main thread after validate completes
							a.g.Update(func(g *gocui.Gui) error {
								a.finishCommand() // Finish after validate completes

								if out, ok := a.panels[ViewOutputs].(*context.OutputContext); ok {
									if err == nil && !validateResult.Valid {
										// Schema has validation errors - show them
										out.LogAction(a.Tr.LogActionSchemaValidationFailed, fmt.Sprintf(a.Tr.LogMsgFoundSchemaErrors, len(validateResult.Errors)))

										// Show validation errors in modal
										modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleSchemaValidationFailed,
											a.Tr.ModalMsgGenerateFailedSchemaErrors,
										).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
										a.OpenModal(modal)
									} else {
										// Schema is valid but generate failed for other reasons
										out.LogAction(a.Tr.LogActionGenerateFailed, fmt.Sprintf(a.Tr.LogMsgFoundSchemaErrors, exitCode))
										modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleGenerateFailed,
											fmt.Sprintf(a.Tr.ModalMsgGenerateFailedWithCode, exitCode),
											a.Tr.ModalMsgSchemaValidCheckOutput,
										).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
										a.OpenModal(modal)
									}
								}
								return nil
							})
						}()
					}
				}
				return nil
			})
		}).
		OnError(func(err error) {
			// Update UI on main thread
			a.g.Update(func(g *gocui.Gui) error {
				if out, ok := a.panels[ViewOutputs].(*context.OutputContext); ok {
					// Check if it's an exit status error (command ran but failed)
					if strings.Contains(err.Error(), "exit status") {
						// Failed - run validate to check schema (keep spinner running)
						out.LogAction(a.Tr.LogActionGenerateFailed, a.Tr.LogMsgCheckingSchemaErrors)

						// Run validate in goroutine to not block UI updates
						go func() {
							validateResult, validateErr := prisma.Validate(cwd)

							// Update UI on main thread after validate completes
							a.g.Update(func(g *gocui.Gui) error {
								a.finishCommand() // Finish after validate completes

								if out, ok := a.panels[ViewOutputs].(*context.OutputContext); ok {
									if validateErr == nil && !validateResult.Valid {
										// Schema has validation errors - show them
										out.LogAction(a.Tr.LogActionSchemaValidationFailed, fmt.Sprintf(a.Tr.LogMsgFoundSchemaErrors, len(validateResult.Errors)))

										// Show validation errors in modal
										modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleSchemaValidationFailed,
											a.Tr.ModalMsgGenerateFailedSchemaErrors,
										).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
										a.OpenModal(modal)
									} else {
										// Schema is valid but generate failed for other reasons
										out.LogAction(a.Tr.LogActionGenerateFailed, err.Error())
										modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleGenerateFailed,
											a.Tr.ModalMsgFailedRunGenerate,
											a.Tr.ModalMsgSchemaValidCheckOutput,
										).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
										a.OpenModal(modal)
									}
								}
								return nil
							})
						}()
					} else {
						// Other error (command couldn't start, etc.)
						a.finishCommand() // Finish immediately on startup error
						out.LogAction(a.Tr.LogActionGenerateError, err.Error())
						modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleGenerateError,
							a.Tr.ModalMsgFailedRunGenerate,
							err.Error(),
						).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
						a.OpenModal(modal)
					}
				}
				return nil
			})
		})

	// Run async to avoid blocking UI (spinner will show automatically)
	if err := generateCmd.RunAsync(); err != nil {
		a.finishCommand() // Clean up if command fails to start
		outputPanel.LogAction(a.Tr.LogActionGenerateError, a.Tr.ModalMsgFailedStartGenerate+" "+err.Error())
		modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleGenerateError,
			a.Tr.ModalMsgFailedStartGenerate,
			err.Error(),
		).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		a.OpenModal(modal)
	}
}
