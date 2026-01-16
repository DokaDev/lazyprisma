# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

LazyPrisma is a terminal UI (TUI) application for Prisma ORM management, built with Go. It uses gocui for terminal UI and supports multiple database backends.

## Commands

```bash
# Build
go build ./...

# Run
go run main.go

# Run tests
go test ./...

# Add dependencies
go get <package>
```

## Architecture

### Core Packages

**pkg/prisma/** - Prisma integration layer
- Project detection (v6 with schema.prisma, v7 with prisma.config.ts)
- Version detection (local vs global Prisma installations)
- Migration inspection and creation
- Schema parsing and datasource extraction
- Environment variable resolution following Prisma's search order

**pkg/database/** - Database abstraction layer (Strategy + Registry pattern)
- `DBDriver` interface defines database operations (Connect, Query, Exec, Close)
- `BaseDriver` provides common sql.DB wrapper implementation
- `Registry` manages driver registration with factory functions
- `Config` builds connection DSN strings with fluent API
- `Client` wrapper converts Prisma URLs to Go driver DSN formats

**pkg/database/drivers/** - Database driver implementations
- Each driver registers itself via `init()` using `MustRegister()`
- Import with blank identifier to auto-register: `_ "github.com/dokadev/lazyprisma/pkg/database/drivers"`
- Supports: PostgreSQL, MySQL, SQLite, SQL Server, CockroachDB

**pkg/commands/** - Shell command execution system
- `CommandBuilder` creates commands with fluent API
- `CommandRunner` interface handles sync/async execution
- `StreamHandler` buffers real-time stdout/stderr for UI updates
- Supports context cancellation and callbacks (OnStdout, OnStderr, OnComplete, OnError)

**pkg/config/** - Application configuration
- YAML-based config at `~/.config/lazyprisma/config.yaml`
- Controls scan depth and directory exclusions
- Auto-creates with defaults if missing

### Key Patterns

1. **Strategy Pattern**: DBDriver interface allows swappable database implementations
2. **Registry Pattern**: Drivers self-register on import, retrieved by name at runtime
3. **Fluent API**: Command and Config builders use method chaining
4. **Callback Pattern**: Commands emit real-time output via callbacks for streaming to UI

### Important Conventions

#### Prisma Version Detection
1. Check local installation: `node_modules/.bin/prisma`
2. Fall back to global installation: `prisma` command
3. Distinguish versions: v7 has `prisma.config.ts`, v6 has only `schema.prisma`

#### Database Driver Registration
```go
// In driver file
func init() {
    database.MustRegister(DriverName, NewDriver)
}

// In main
import _ "github.com/dokadev/lazyprisma/pkg/database/drivers"
```

#### Environment Variable Resolution (Prisma Order)
Follows Prisma's resolution order:
1. OS environment variables (highest priority)
2. `.env` in project root directory
3. `.env` in schema directory
4. `.env` in `prisma/` directory

#### Command Building
```go
cmd := cmdBuilder.New("npx", "prisma", "--version").
    WithWorkingDir(projectDir).
    OnStdout(func(line string) {
        // Handle real-time output
    }).
    RunWithOutput()
```

#### Migration Tracking
- Migrations stored in `_prisma_migrations` table
- Checksums detect modified migrations
- Failed migrations have `finished_at IS NULL`
- Rollback tracked in `rolled_back_at` field

### Dependencies

**Core:**
- `github.com/jesseduffield/gocui` - Terminal UI framework (from lazygit author)
- `github.com/jesseduffield/lazycore/pkg/boxlayout` - Layout engine for UI panels

**Database Drivers:**
- `github.com/lib/pq` - PostgreSQL driver
- `github.com/go-sql-driver/mysql` - MySQL driver

## Development Patterns

### Adding a New Database Driver

1. Create file `pkg/database/drivers/{dbname}.go`
2. Implement DBDriver interface (or embed BaseDriver and implement Connect)
3. Self-register in init():
   ```go
   func init() {
       database.MustRegister("driver-name", NewDriver)
   }
   ```
4. Driver automatically available via `database.Get("driver-name")`

### Running Prisma Commands

```go
// Via high-level functions
err := prisma.Generate(projectDir, &prisma.GenerateOptions{
    Schema: schemaPath,
    Watch: true,
})

// Via CommandBuilder (for custom commands)
cmd := cmdBuilder.New("npx", "prisma", "migrate", "dev", "--name", "init").
    WithWorkingDir(projectDir).
    StreamOutput()
```

### Detecting Projects

```go
projects, err := prisma.DetectProjects(dir, &prisma.ScanOptions{
    MaxDepth: 10,
    ExcludeDirs: prisma.DefaultExcludeDirs,
})

for _, project := range projects {
    // Check capabilities
    ds, _ := prisma.GetProjectDatasource(&project)
    dbType := prisma.DBType(ds.Provider)
    caps := prisma.GetCapabilities(dbType)

    if caps.Migrate {
        // Database supports prisma migrate
    }
}
```

### Querying Migrations

```go
// Connect to database using Prisma datasource
ds, _ := prisma.GetProjectDatasource(project)
client, _ := database.NewClientFromDSN(ds.Provider, ds.URL)
defer client.Close()

// Get migrations
migrations, _ := prisma.GetSuccessfulMigrations(client)
for _, m := range migrations {
    fmt.Printf("%s: %s\n", m.Name, m.Status)
}
```

### Database Capabilities

Different database providers support different Prisma features:

```go
caps := prisma.GetCapabilities(prisma.DBType("postgresql"))
// caps.Migrate: true (supports prisma migrate)
// caps.Branching: false (doesn't use branching workflow)
// caps.ShadowDB: true (supports shadow database)

caps := prisma.GetCapabilities(prisma.DBType("planetscale"))
// caps.Migrate: false
// caps.Branching: true (uses db branching instead)
```

Branching databases (PlanetScale, Turso, Neon) don't use traditional migrations.
