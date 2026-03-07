package main

import (
	"fmt"
	"os"

	"github.com/dokadev/lazyprisma/pkg/app"
	"github.com/dokadev/lazyprisma/pkg/config"
	"github.com/dokadev/lazyprisma/pkg/i18n"
	"github.com/dokadev/lazyprisma/pkg/prisma"

	// Register database drivers
	_ "github.com/dokadev/lazyprisma/pkg/database/drivers"
)

const (
	Version   = "v0.2.2"
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
		fmt.Fprintf(os.Stderr, tr.ErrorNotPrismaWorkspace)
		fmt.Fprintf(os.Stderr, tr.ErrorExpectedOneOf)
		fmt.Fprintf(os.Stderr, tr.ErrorExpectedConfigV7Plus)
		fmt.Fprintf(os.Stderr, tr.ErrorExpectedSchemaV7Minus)
		os.Exit(1)
	}

	// App 생성
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

	// 패널 생성 및 등록
	workspace := app.NewWorkspacePanel(tuiApp.GetGui(), tr)
	migrations := app.NewMigrationsPanel(tuiApp.GetGui(), tr)
	details := app.NewDetailsPanel(tuiApp.GetGui(), tr)
	output := app.NewOutputPanel(tuiApp.GetGui(), tr)
	statusbar := app.NewStatusBar(tuiApp.GetGui(), tuiApp)

	// Connect panels
	migrations.SetDetailsPanel(details)
	details.SetApp(tuiApp)

	tuiApp.RegisterPanel(workspace)
	tuiApp.RegisterPanel(migrations)
	tuiApp.RegisterPanel(details)
	tuiApp.RegisterPanel(output)
	tuiApp.RegisterPanel(statusbar)

	// 키바인딩 등록
	if err := tuiApp.RegisterKeybindings(); err != nil {
		fmt.Fprintf(os.Stderr, tr.ErrorFailedRegisterKeybindings, err)
		os.Exit(1)
	}

	// 마우스 바인딩 등록
	tuiApp.RegisterMouseBindings()

	// 실행
	if err := tuiApp.Run(); err != nil {
		fmt.Fprintf(os.Stderr, tr.ErrorAppRuntime, err)
		os.Exit(1)
	}
}
