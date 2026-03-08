package main

import (
	"fmt"
	"os"

	"github.com/dokadev/lazyprisma/pkg/app"
	"github.com/dokadev/lazyprisma/pkg/config"
	"github.com/dokadev/lazyprisma/pkg/gui/context"
	"github.com/dokadev/lazyprisma/pkg/i18n"
	"github.com/dokadev/lazyprisma/pkg/prisma"

	// Register database drivers
	_ "github.com/dokadev/lazyprisma/pkg/database/drivers"
)

const (
	Version   = "v0.3.2"
	Developer = "DokaLab"
)

func main() {
	cfg, _ := config.Load()
	tr := i18n.NewTranslationSet(cfg.Language)

	// Handle version flag
	if len(os.Args) > 1 {
		if os.Args[1] == "--version" || os.Args[1] == "-v" {
			fmt.Printf(tr.VersionOutput, Version, Developer)
			os.Exit(0)
		}
	}

	// Check if current directory is a Prisma workspace
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, tr.ErrorFailedGetCurrentDir, err)
		os.Exit(1)
	}

	if !prisma.IsWorkspace(cwd) {
		fmt.Fprint(os.Stderr, tr.ErrorNotPrismaWorkspace)
		fmt.Fprint(os.Stderr, tr.ErrorExpectedOneOf)
		fmt.Fprint(os.Stderr, tr.ErrorExpectedConfigV7Plus)
		fmt.Fprint(os.Stderr, tr.ErrorExpectedSchemaV7Minus)
		os.Exit(1)
	}

	// Create app
	tuiApp, err := app.NewApp(app.AppConfig{
		DebugMode: false,
		AppName:   "LazyPrisma",
		Version:   Version,
		Developer: Developer,
		Language:  cfg.Language,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, tr.ErrorFailedCreateApp, err)
		os.Exit(1)
	}

	// Create and register panels
	workspace := context.NewWorkspaceContext(context.WorkspaceContextOpts{
		Gui:      tuiApp.GetGui(),
		Tr:       tr,
		ViewName: "workspace",
	})
	migrationsCtx := context.NewMigrationsContext(context.MigrationsContextOpts{
		Gui:      tuiApp.GetGui(),
		Tr:       tr,
		ViewName: "migrations",
	})
	detailsCtx := context.NewDetailsContext(context.DetailsContextOpts{
		Gui:      tuiApp.GetGui(),
		Tr:       tr,
		ViewName: "details",
	})
	output := context.NewOutputContext(context.OutputContextOpts{
		Gui:      tuiApp.GetGui(),
		Tr:       tr,
		ViewName: "outputs",
	})
	statusbar := context.NewStatusBarContext(context.StatusBarContextOpts{
		Gui:      tuiApp.GetGui(),
		Tr:       tr,
		ViewName: "statusbar",
		State:    tuiApp.StatusBarState(),
		Config: context.StatusBarConfig{
			Developer: Developer,
			Version:   Version,
		},
	})

	// Wire callbacks (replaces old bidirectional coupling)
	migrationsCtx.SetOnSelectionChanged(func(mig *prisma.Migration, tab string) {
		detailsCtx.UpdateFromMigration(mig, tab)
	})
	migrationsCtx.SetModalCallbacks(tuiApp.HasActiveModal, func(viewID string) {
		tuiApp.HandlePanelClick(viewID)
	})
	detailsCtx.SetModalCallbacks(tuiApp.HasActiveModal, func(viewID string) {
		tuiApp.HandlePanelClick(viewID)
	})

	// Load action-needed data for details context
	detailsCtx.SetActionNeededMigrations(collectActionNeededMigrations(migrationsCtx.GetCategory()))
	detailsCtx.LoadActionNeededData()

	tuiApp.RegisterPanel(workspace)
	tuiApp.RegisterPanel(migrationsCtx)
	tuiApp.RegisterPanel(detailsCtx)
	tuiApp.RegisterPanel(output)
	tuiApp.RegisterPanel(statusbar)

	// Create and wire controllers
	gui := tuiApp.GetGui()

	migrationsController := app.NewMigrationsController(
		tuiApp, gui, migrationsCtx, output,
		tuiApp.OpenModal, tuiApp.CloseModal,
		tuiApp.RunStreamingCommand,
	)
	generateController := app.NewGenerateController(
		tuiApp, gui, output,
		tuiApp.OpenModal,
		tuiApp.RunStreamingCommand,
	)
	studioController := app.NewStudioController(
		tuiApp, gui, output,
		tuiApp.OpenModal,
	)
	clipboardController := app.NewClipboardController(
		tuiApp, gui, migrationsCtx,
		tuiApp.OpenModal, tuiApp.CloseModal,
	)

	tuiApp.SetControllers(migrationsController, generateController, studioController, clipboardController)

	// Register keybindings
	if err := tuiApp.RegisterKeybindings(); err != nil {
		fmt.Fprintf(os.Stderr, tr.ErrorFailedRegisterKeybindings, err)
		os.Exit(1)
	}

	// Register mouse bindings
	tuiApp.RegisterMouseBindings()

	// Run
	if err := tuiApp.Run(); err != nil {
		fmt.Fprintf(os.Stderr, tr.ErrorAppRuntime, err)
		os.Exit(1)
	}
}

// collectActionNeededMigrations extracts migrations that need action (Empty or Checksum Mismatch)
// from the Local category.
func collectActionNeededMigrations(cat prisma.MigrationCategory) []prisma.Migration {
	var result []prisma.Migration
	for _, mig := range cat.Local {
		if mig.IsEmpty || mig.ChecksumMismatch {
			result = append(result, mig)
		}
	}
	return result
}
