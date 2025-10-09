# LazyPrisma

Terminal UI tool for managing Prisma migrations and database.

<img width="1458" height="918" alt="image" src="https://github.com/user-attachments/assets/71c2db93-b784-4a0b-9740-7bb3710c4caf" />

## Installation

### Homebrew (macOS/Linux)
```bash
brew tap DokaDev/lazyprisma
brew install lazyprisma
```

### Manual Installation
Download the latest binary from [Releases](https://github.com/DokaDev/lazyprisma/releases).

## Prerequisites
LazyPrisma requires Prisma CLI to be installed in your project:
```bash
npm install -D prisma
```

> **Note:** LazyPrisma uses `npx prisma` to execute commands. Global Prisma installation is not currently supported.

## Usage
```bash
cd your-prisma-project
lazyprisma
```

### Keyboard Shortcuts
- `←/→` - Switch panels
- `↑/↓` - Scroll/Select
- `r` - Refresh migration status
- `g` - Generate Prisma Client
- `d` - Migrate Dev (create new migration)
- `D` - Migrate Deploy
- `f` - Format schema
- `t` - Open Prisma Studio
- `?` - Help
- `q` - Quit

## Build from Source
```bash
brew install make

make clean
make
```
