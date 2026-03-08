package app

import (
	"fmt"
	"strings"

	"github.com/dokadev/lazyprisma/pkg/gui/context"
	"github.com/dokadev/lazyprisma/pkg/prisma"
	"github.com/jesseduffield/gocui"
)

// Generate runs prisma generate and shows result in modal
func (a *App) Generate() {
	a.runStreamingCommand(AsyncCommandOpts{
		Name:          "Generate",
		Args:          []string{"npx", "prisma", "generate"},
		LogAction:     a.Tr.LogActionGenerate,
		LogDetail:     a.Tr.LogMsgRunningGenerate,
		ErrorTitle:    a.Tr.ModalTitleGenerateError,
		ErrorStartMsg: a.Tr.ModalMsgFailedStartGenerate,
		OnSuccess: func(out *context.OutputContext, cwd string) {
			a.finishCommand() // Finish immediately on success
			out.LogAction(a.Tr.LogActionGenerateComplete, a.Tr.LogMsgPrismaClientGeneratedSuccess)
			modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleGenerateSuccess,
				a.Tr.ModalMsgPrismaClientGenerated,
			).WithStyle(MessageModalStyle{TitleColor: ColorGreen, BorderColor: ColorGreen})
			a.OpenModal(modal)
		},
		OnFailure: func(out *context.OutputContext, cwd string, exitCode int) {
			// Don't finishCommand yet — validate first (keep spinner running)
			out.LogAction(a.Tr.LogActionGenerateFailed, a.Tr.LogMsgCheckingSchemaErrors)

			go func() {
				validateResult, err := prisma.Validate(cwd)

				a.g.Update(func(g *gocui.Gui) error {
					a.finishCommand() // Finish after validate completes

					if out, ok := a.panels[ViewOutputs].(*context.OutputContext); ok {
						if err == nil && !validateResult.Valid {
							out.LogAction(a.Tr.LogActionSchemaValidationFailed, fmt.Sprintf(a.Tr.LogMsgFoundSchemaErrors, len(validateResult.Errors)))
							modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleSchemaValidationFailed,
								a.Tr.ModalMsgGenerateFailedSchemaErrors,
							).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
							a.OpenModal(modal)
						} else {
							out.LogAction(a.Tr.LogActionGenerateFailed, fmt.Sprintf(a.Tr.ModalMsgGenerateFailedWithCode, exitCode))
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
		},
		OnError: func(out *context.OutputContext, cwd string, err error) {
			// Check if it's an exit status error (command ran but failed)
			if strings.Contains(err.Error(), "exit status") {
				// Don't finishCommand yet — validate first (keep spinner running)
				out.LogAction(a.Tr.LogActionGenerateFailed, a.Tr.LogMsgCheckingSchemaErrors)

				go func() {
					validateResult, validateErr := prisma.Validate(cwd)

					a.g.Update(func(g *gocui.Gui) error {
						a.finishCommand() // Finish after validate completes

						if out, ok := a.panels[ViewOutputs].(*context.OutputContext); ok {
							if validateErr == nil && !validateResult.Valid {
								out.LogAction(a.Tr.LogActionSchemaValidationFailed, fmt.Sprintf(a.Tr.LogMsgFoundSchemaErrors, len(validateResult.Errors)))
								modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleSchemaValidationFailed,
									a.Tr.ModalMsgGenerateFailedSchemaErrors,
								).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
								a.OpenModal(modal)
							} else {
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
		},
	})
}
