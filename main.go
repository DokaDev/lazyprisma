package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"lazyprisma/pkg/prisma"
	"lazyprisma/pkg/tui"
	"lazyprisma/pkg/version"
)

func main() {
	// Handle --version flag
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("lazyprisma version %s\n", version.Version)
		os.Exit(0)
	}

	status := prisma.GetStatus()

	if !status.CLIAvailable {
		fmt.Println("Prisma CLI is not available!")
		fmt.Println("Please install Prisma first:")
		fmt.Println("  npm install -D prisma")
		os.Exit(1)
	}

	if !status.SchemaExists {
		if askForInit() {
			fmt.Println("\nInitializing Prisma...")
			executor := prisma.NewExecutor()
			output, err := executor.Init()

			if err != nil {
				fmt.Printf("Error: %v\n", err)
				fmt.Println(output)
				os.Exit(1)
			}

			fmt.Println("Prisma initialized successfully!")
			fmt.Println(output)
			fmt.Println("\nPress Enter to continue...")
			bufio.NewReader(os.Stdin).ReadBytes('\n')

			status = prisma.GetStatus()
		}
	}

	app, err := tui.NewApp(status)
	if err != nil {
		log.Fatal(err)
	}

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

func askForInit() bool {
	fmt.Println("No Prisma schema detected.")
	fmt.Println("")
	fmt.Println("Would you like to initialize Prisma now?")
	fmt.Println("This will:")
	fmt.Println("  - Create prisma/schema.prisma")
	fmt.Println("  - Generate .env file with DATABASE_URL")
	fmt.Println("")
	fmt.Print("Initialize Prisma? [Y/n]: ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	input = strings.TrimSpace(strings.ToLower(input))
	return input == "" || input == "y" || input == "yes"
}
