package main

import (
	"fmt"
	"os"

	"github.com/dokadev/lazyprisma/pkg/app"
	"github.com/dokadev/lazyprisma/pkg/prisma"

	// Register database drivers
	_ "github.com/dokadev/lazyprisma/pkg/database/drivers"
)

func main() {
	// Check if current directory is a Prisma workspace
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get current directory: %v\n", err)
		os.Exit(1)
	}

	if !prisma.IsWorkspace(cwd) {
		fmt.Fprintf(os.Stderr, "Error: Current directory is not a Prisma workspace.\n")
		fmt.Fprintf(os.Stderr, "\nExpected one of:\n")
		fmt.Fprintf(os.Stderr, "  - prisma.config.ts (Prisma v7.0+)\n")
		fmt.Fprintf(os.Stderr, "  - prisma/schema.prisma (Prisma < v7.0)\n")
		os.Exit(1)
	}

	// App 생성
	tuiApp, err := app.NewApp(app.AppConfig{
		DebugMode: false,
		AppName:   "LazyPrisma",
		Version:   "v0.2.0",
		Developer: "DokaLab",
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create app: %v\n", err)
		os.Exit(1)
	}

	// 패널 생성 및 등록
	workspace := app.NewWorkspacePanel(tuiApp.GetGui())
	migrations := app.NewMigrationsPanel(tuiApp.GetGui())
	details := app.NewDetailsPanel(tuiApp.GetGui())
	output := app.NewOutputPanel(tuiApp.GetGui())
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
		fmt.Fprintf(os.Stderr, "Failed to register keybindings: %v\n", err)
		os.Exit(1)
	}

	// 마우스 바인딩 등록
	tuiApp.RegisterMouseBindings()

	// 실행
	if err := tuiApp.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "App error: %v\n", err)
		os.Exit(1)
	}
}
