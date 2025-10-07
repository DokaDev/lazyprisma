# LazyPrisma

Terminal UI tool for managing Prisma migrations and database.

## Installation

### Homebrew (macOS/Linux)
```bash
brew tap linkareer_123/lazyprisma
brew install lazyprisma
```

### Manual Installation
Download the latest binary from [Releases](https://github.com/linkareer_123/lazyprisma/releases).

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
- `h` - Help
- `q` - Quit

## Build from Source
```bash
brew install make

make clean
make
```