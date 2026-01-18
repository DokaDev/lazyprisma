# LazyPrisma

<img width="230" alt="lazyprisma_ico_scl" src="https://github.com/user-attachments/assets/5df09da0-61c2-4a10-a58a-d8190a5a4fd6" />


A Terminal UI tool for managing Prisma migrations and the database, designed for developers who prefer the command line.

<!-- <img width="1458" height="918" alt="image" src="https://github.com/user-attachments/assets/71c2db93-b784-4a0b-9740-7bb3710c4caf" /> -->
![2](https://github.com/user-attachments/assets/24f25041-03c5-4bcc-971f-8155e0a1d1d9)



> **Note on Appearance**: The screenshot above features the **Dracula Dark** theme and **JetBrains Mono Nerd Font**. We highly recommend using a [Nerd Font](https://www.nerdfonts.com/) for the best visual experience, as future updates will introduce more icons and symbols.

## Features

- **Visualise Migrations**: View Local, Pending, and DB-Only migrations in a clean, organised TUI.
- **Safe Workflow**: Built-in validations for checksum mismatches and empty migrations to prevent database inconsistencies.
- **Prisma Studio Integration**: Toggle Prisma Studio directly from the app (`S` key) with automatic process management (no more zombie processes).
- **Migration Management**: Create (`d`), Deploy (`D`), and Resolve (`s`) migrations effortlessly.
- **Quick Actions**: Delete pending migrations (`Del`/`Backspace`) and copy migration details to the clipboard (`c`).

## Installation

### Homebrew (macOS/Linux)
```bash
brew tap DokaDev/lazyprisma
brew install lazyprisma
```

### Manual Installation
Download the latest binary from [Releases](https://github.com/DokaDev/lazyprisma/releases).

## Prerequisites

LazyPrisma requires the Prisma CLI to be installed in your project:

```bash
npm install -D prisma
```

> **Note:** LazyPrisma uses `npx prisma` to execute commands. Ensure `npx` is available in your shell path. It supports both the classic `schema.prisma` and the new Prisma v7+ `prisma.config.ts`.

## Usage

Navigate to your project directory and launch the application:

```bash
cd your-prisma-project
lazyprisma
```

Check the version:
```bash
lazyprisma --version
```

### Keyboard Shortcuts

**Navigation**
- `←` / `→`: Switch between panels (Workspace, Migrations, Details, Output).
- `↑` / `↓`: Scroll list or text content.
- `Tab` / `Shift+Tab`: Switch tabs within a panel (e.g., Local / Pending / DB-Only).

**Core Actions**
- `r`: **Refresh** all panels and migration status.
- `d`: **Migrate Dev** – Create a new migration (Schema diff-based or empty Manual migration).
- `D`: **Migrate Deploy** – Apply pending migrations to the database.
- `g`: **Generate** – Run `prisma generate` to update the client.
- `s`: **Resolve** – Fix failed migrations (mark as applied or rolled back).
- `S`: **Studio** – Toggle the Prisma Studio server (opens in your default browser).

**Utilities**
- `c`: **Copy** – Copy the selected migration's name, path, or checksum to the clipboard.
- `⌫` / `Del`: **Delete** – Delete the selected pending local migration folder.
- `q`: **Quit** – Exit the application (safely terminates any background Prisma Studio processes).

## Build from Source

Ensure you have Go installed (1.21+ recommended).

```bash
# Clean and build
make clean
make

# Build and run immediately

make run

```



## Roadmap



- **Automatic `down.sql` Generation**: Initially planned for v0.2.x, but postponed due to technical constraints. We aim to implement this feature in a future sprint, alongside establishing a robust `lazytui` framework and refactoring the codebase.
