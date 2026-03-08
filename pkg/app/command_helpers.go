package app

import (
	"os"

	"github.com/dokadev/lazyprisma/pkg/commands"
	"github.com/dokadev/lazyprisma/pkg/gui/context"
	"github.com/jesseduffield/gocui"
)

// AsyncCommandOpts configures a streaming async command.
type AsyncCommandOpts struct {
	Name         string   // for tryStartCommand / logCommandBlocked
	Args         []string // full command args: ["npx", "prisma", "migrate", "deploy"]
	LogAction    string   // log action label (e.g., "Migrate Deploy")
	LogDetail    string   // log detail text (e.g., "Running prisma migrate deploy...")
	SkipTryStart bool     // true if tryStartCommand was already called by the caller

	// Callbacks — each callback is responsible for calling finishCommand() at the appropriate time.
	// The helper never calls finishCommand() itself.
	OnSuccess func(out *context.OutputContext, cwd string)
	OnFailure func(out *context.OutputContext, cwd string, exitCode int)
	OnError   func(out *context.OutputContext, cwd string, err error)

	// ErrorTitle and ErrorStartMsg are used for the default RunAsync failure modal.
	// If empty, generic error text is used.
	ErrorTitle    string
	ErrorStartMsg string
}

// runStreamingCommand handles the common boilerplate for streaming prisma commands.
// Returns false if the command could not be started (another command running or panel missing).
// The helper does NOT call finishCommand() — each callback is responsible for calling it.
func (a *App) runStreamingCommand(opts AsyncCommandOpts) bool {
	// Phase 1: Guard
	if !opts.SkipTryStart {
		if !a.tryStartCommand(opts.Name) {
			a.logCommandBlocked(opts.Name)
			return false
		}
	}

	// Phase 2: Get output panel
	outputPanel, ok := a.panels[ViewOutputs].(*context.OutputContext)
	if !ok {
		a.finishCommand()
		return false
	}

	// Phase 3: Get cwd
	cwd, err := os.Getwd()
	if err != nil {
		a.finishCommand()
		a.g.Update(func(g *gocui.Gui) error {
			outputPanel.LogAction(opts.LogAction, a.Tr.ErrorFailedGetWorkingDir+" "+err.Error())
			modal := NewMessageModal(a.g, a.Tr, opts.ErrorTitle,
				a.Tr.ErrorFailedGetWorkingDir,
				err.Error(),
			).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
			a.OpenModal(modal)
			return nil
		})
		return false
	}

	// Phase 4: Log action start
	a.g.Update(func(g *gocui.Gui) error {
		if out, ok := a.panels[ViewOutputs].(*context.OutputContext); ok {
			out.LogAction(opts.LogAction, opts.LogDetail)
		}
		return nil
	})

	// Phase 5: Build command
	builder := commands.NewCommandBuilder(commands.NewPlatform())

	cmd := builder.New(opts.Args...).
		WithWorkingDir(cwd).
		StreamOutput().
		OnStdout(func(line string) {
			a.g.Update(func(g *gocui.Gui) error {
				if out, ok := a.panels[ViewOutputs].(*context.OutputContext); ok {
					out.AppendOutput("  " + line)
				}
				return nil
			})
		}).
		OnStderr(func(line string) {
			a.g.Update(func(g *gocui.Gui) error {
				if out, ok := a.panels[ViewOutputs].(*context.OutputContext); ok {
					out.AppendOutput("  " + line)
				}
				return nil
			})
		}).
		OnComplete(func(exitCode int) {
			a.g.Update(func(g *gocui.Gui) error {
				if out, ok := a.panels[ViewOutputs].(*context.OutputContext); ok {
					if exitCode == 0 {
						if opts.OnSuccess != nil {
							opts.OnSuccess(out, cwd)
						} else {
							a.finishCommand()
						}
					} else {
						if opts.OnFailure != nil {
							opts.OnFailure(out, cwd, exitCode)
						} else {
							a.finishCommand()
						}
					}
				} else {
					a.finishCommand()
				}
				return nil
			})
		}).
		OnError(func(err error) {
			a.g.Update(func(g *gocui.Gui) error {
				if out, ok := a.panels[ViewOutputs].(*context.OutputContext); ok {
					if opts.OnError != nil {
						opts.OnError(out, cwd, err)
					} else {
						a.finishCommand()
					}
				} else {
					a.finishCommand()
				}
				return nil
			})
		})

	// Phase 6: RunAsync
	if err := cmd.RunAsync(); err != nil {
		a.finishCommand()
		errorTitle := opts.ErrorTitle
		errorMsg := opts.ErrorStartMsg
		a.g.Update(func(g *gocui.Gui) error {
			outputPanel.LogAction(errorTitle, errorMsg+" "+err.Error())
			modal := NewMessageModal(a.g, a.Tr, errorTitle,
				errorMsg,
				err.Error(),
			).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
			a.OpenModal(modal)
			return nil
		})
		return false
	}

	return true
}
