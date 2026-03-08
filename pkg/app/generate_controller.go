package app

import (
	"fmt"
	"strings"

	"github.com/dokadev/lazyprisma/pkg/gui/context"
	"github.com/dokadev/lazyprisma/pkg/gui/types"
	"github.com/dokadev/lazyprisma/pkg/prisma"
	"github.com/jesseduffield/gocui"
)

// GenerateController handles prisma generate operations.
type GenerateController struct {
	c             types.IControllerHost
	g             *gocui.Gui
	outputCtx     *context.OutputContext
	openModal     func(Modal)
	runStreamCmd  func(AsyncCommandOpts) bool
}

// NewGenerateController creates a new GenerateController.
func NewGenerateController(
	c types.IControllerHost,
	g *gocui.Gui,
	outputCtx *context.OutputContext,
	openModal func(Modal),
	runStreamCmd func(AsyncCommandOpts) bool,
) *GenerateController {
	return &GenerateController{
		c:            c,
		g:            g,
		outputCtx:    outputCtx,
		openModal:    openModal,
		runStreamCmd: runStreamCmd,
	}
}

// Generate runs prisma generate and shows result in modal
func (gc *GenerateController) Generate() {
	tr := gc.c.GetTranslationSet()

	gc.runStreamCmd(AsyncCommandOpts{
		Name:          "Generate",
		Args:          []string{"npx", "prisma", "generate"},
		LogAction:     tr.LogActionGenerate,
		LogDetail:     tr.LogMsgRunningGenerate,
		ErrorTitle:    tr.ModalTitleGenerateError,
		ErrorStartMsg: tr.ModalMsgFailedStartGenerate,
		OnSuccess: func(out *context.OutputContext, cwd string) {
			gc.c.FinishCommand() // Finish immediately on success
			out.LogAction(tr.LogActionGenerateComplete, tr.LogMsgPrismaClientGeneratedSuccess)
			modal := NewMessageModal(gc.g, tr, tr.ModalTitleGenerateSuccess,
				tr.ModalMsgPrismaClientGenerated,
			).WithStyle(MessageModalStyle{TitleColor: ColorGreen, BorderColor: ColorGreen})
			gc.openModal(modal)
		},
		OnFailure: func(out *context.OutputContext, cwd string, exitCode int) {
			// Don't finishCommand yet -- validate first (keep spinner running)
			out.LogAction(tr.LogActionGenerateFailed, tr.LogMsgCheckingSchemaErrors)

			go func() {
				validateResult, err := prisma.Validate(cwd)

				gc.c.OnUIThread(func() error {
					gc.c.FinishCommand() // Finish after validate completes

					if err == nil && !validateResult.Valid {
						gc.outputCtx.LogAction(tr.LogActionSchemaValidationFailed, fmt.Sprintf(tr.LogMsgFoundSchemaErrors, len(validateResult.Errors)))
						modal := NewMessageModal(gc.g, tr, tr.ModalTitleSchemaValidationFailed,
							tr.ModalMsgGenerateFailedSchemaErrors,
						).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
						gc.openModal(modal)
					} else {
						gc.outputCtx.LogAction(tr.LogActionGenerateFailed, fmt.Sprintf(tr.ModalMsgGenerateFailedWithCode, exitCode))
						modal := NewMessageModal(gc.g, tr, tr.ModalTitleGenerateFailed,
							fmt.Sprintf(tr.ModalMsgGenerateFailedWithCode, exitCode),
							tr.ModalMsgSchemaValidCheckOutput,
						).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
						gc.openModal(modal)
					}
					return nil
				})
			}()
		},
		OnError: func(out *context.OutputContext, cwd string, err error) {
			// Check if it's an exit status error (command ran but failed)
			if strings.Contains(err.Error(), "exit status") {
				// Don't finishCommand yet -- validate first (keep spinner running)
				out.LogAction(tr.LogActionGenerateFailed, tr.LogMsgCheckingSchemaErrors)

				go func() {
					validateResult, validateErr := prisma.Validate(cwd)

					gc.c.OnUIThread(func() error {
						gc.c.FinishCommand() // Finish after validate completes

						if validateErr == nil && !validateResult.Valid {
							gc.outputCtx.LogAction(tr.LogActionSchemaValidationFailed, fmt.Sprintf(tr.LogMsgFoundSchemaErrors, len(validateResult.Errors)))
							modal := NewMessageModal(gc.g, tr, tr.ModalTitleSchemaValidationFailed,
								tr.ModalMsgGenerateFailedSchemaErrors,
							).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
							gc.openModal(modal)
						} else {
							gc.outputCtx.LogAction(tr.LogActionGenerateFailed, err.Error())
							modal := NewMessageModal(gc.g, tr, tr.ModalTitleGenerateFailed,
								tr.ModalMsgFailedRunGenerate,
								tr.ModalMsgSchemaValidCheckOutput,
							).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
							gc.openModal(modal)
						}
						return nil
					})
				}()
			} else {
				// Other error (command couldn't start, etc.)
				gc.c.FinishCommand() // Finish immediately on startup error
				out.LogAction(tr.LogActionGenerateError, err.Error())
				modal := NewMessageModal(gc.g, tr, tr.ModalTitleGenerateError,
					tr.ModalMsgFailedRunGenerate,
					err.Error(),
				).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
				gc.openModal(modal)
			}
		},
	})
}
