package database

import (
	"fmt"
	"time"
)

// RunTestSuite runs database module tests
func RunTestSuite() {
	fmt.Println("=== LazyPrisma Database Module Test ===\n")

	TestRegistryList()
	TestConfigBuilders()
	TestClientCreation()

	// Real DB connection tests
	TestMySQLPrismaMigrations()
	TestPostgresPrismaMigrations()

	fmt.Println("=== Database Tests Completed ===")
}

// TestMySQLPrismaMigrations connects to MySQL and fetches Prisma migrations
func TestMySQLPrismaMigrations() {
	fmt.Println("Test 4: MySQL Prisma Migrations")
	fmt.Println("--------------------------------")

	client, err := NewClient("mysql")
	if err != nil {
		fmt.Printf("✗ Failed to create client: %v\n", err)
		return
	}

	cfg := NewConfig().
		WithHost("localhost").
		WithPort(3308).
		WithUser("root").
		WithPassword("1234").
		WithDatabase("linkareer_local_dev_db")

	fmt.Printf("Connecting to: mysql://%s:****@%s:%d/%s\n", cfg.User, cfg.Host, cfg.Port, cfg.Database)

	if err := client.Connect(cfg); err != nil {
		fmt.Printf("✗ Failed to connect: %v\n", err)
		return
	}
	defer client.Close()

	fmt.Println("✓ Connected to MySQL\n")

	// Query Prisma migrations table
	rows, err := client.Query(`
		SELECT
			id,
			migration_name,
			started_at,
			finished_at,
			applied_steps_count
		FROM _prisma_migrations
		ORDER BY started_at DESC
		LIMIT 10
	`)
	if err != nil {
		fmt.Printf("✗ Failed to query migrations: %v\n", err)
		return
	}
	defer rows.Close()

	fmt.Println("\nPrisma Migrations (latest 10):")

	count := 0
	for rows.Next() {
		var id, name string
		var startedAt time.Time
		var finishedAt *time.Time
		var steps int

		if err := rows.Scan(&id, &name, &startedAt, &finishedAt, &steps); err != nil {
			fmt.Printf("✗ Scan error: %v\n", err)
			continue
		}

		fmt.Printf("[%d] %s\n", count+1, name)
		fmt.Printf("    ID: %s\n", id[:8])
		fmt.Printf("    Started: %s\n", startedAt.Format("2006-01-02 15:04:05"))
		if finishedAt != nil {
			fmt.Printf("    Finished: %s\n", finishedAt.Format("2006-01-02 15:04:05"))
		} else {
			fmt.Printf("    Finished: (failed)\n")
		}
		fmt.Printf("    Steps: %d\n\n", steps)
		count++
	}

	if count == 0 {
		fmt.Println("No migrations found.")
	} else {
		fmt.Printf("✓ Total: %d migration(s)\n", count)
	}
	fmt.Println()
}

// TestPostgresPrismaMigrations connects to PostgreSQL and fetches Prisma migrations
func TestPostgresPrismaMigrations() {
	fmt.Println("Test 5: PostgreSQL Prisma Migrations")
	fmt.Println("-------------------------------------")

	client, err := NewClient("postgres")
	if err != nil {
		fmt.Printf("✗ Failed to create client: %v\n", err)
		return
	}

	cfg := NewConfig().
		WithHost("localhost").
		WithPort(6432).
		WithUser("linkareer").
		WithPassword("1234").
		WithDatabase("linkareer_chat_local_db")

	fmt.Printf("Connecting to: postgresql://%s:****@%s:%d/%s\n", cfg.User, cfg.Host, cfg.Port, cfg.Database)

	if err := client.Connect(cfg); err != nil {
		fmt.Printf("✗ Failed to connect: %v\n\n", err)
		return
	}
	defer client.Close()

	fmt.Println("✓ Connected to PostgreSQL")

	// Query Prisma migrations table
	rows, err := client.Query(`
		SELECT
			id,
			migration_name,
			started_at,
			finished_at,
			applied_steps_count
		FROM _prisma_migrations
		ORDER BY started_at DESC
		LIMIT 10
	`)
	if err != nil {
		fmt.Printf("✗ Failed to query migrations: %v\n\n", err)
		return
	}
	defer rows.Close()

	fmt.Println("\nPrisma Migrations (latest 10):")

	count := 0
	for rows.Next() {
		var id, name string
		var startedAt time.Time
		var finishedAt *time.Time
		var steps int

		if err := rows.Scan(&id, &name, &startedAt, &finishedAt, &steps); err != nil {
			fmt.Printf("✗ Scan error: %v\n", err)
			continue
		}

		fmt.Printf("[%d] %s\n", count+1, name)
		fmt.Printf("    ID: %s\n", id[:8])
		fmt.Printf("    Started: %s\n", startedAt.Format("2006-01-02 15:04:05"))
		if finishedAt != nil {
			fmt.Printf("    Finished: %s\n", finishedAt.Format("2006-01-02 15:04:05"))
		} else {
			fmt.Printf("    Finished: (failed)\n")
		}
		fmt.Printf("    Steps: %d\n\n", steps)
		count++
	}

	if count == 0 {
		fmt.Println("No migrations found.")
	} else {
		fmt.Printf("✓ Total: %d migration(s)\n", count)
	}
	fmt.Println()
}

// TestRegistryList tests driver registration
func TestRegistryList() {
	fmt.Println("Test 1: Registry - List Registered Drivers")
	fmt.Println("-------------------------------------------")

	drivers := List()
	fmt.Printf("Registered drivers: %v\n", drivers)

	for _, name := range drivers {
		fmt.Printf("  - %s: Has=%v\n", name, Has(name))
	}
	fmt.Println()
}

// TestConfigBuilders tests config DSN builders
func TestConfigBuilders() {
	fmt.Println("Test 2: Config - DSN Builders")
	fmt.Println("-----------------------------")

	cfg := NewConfig().
		WithHost("localhost").
		WithPort(5432).
		WithUser("admin").
		WithPassword("secret").
		WithDatabase("testdb").
		WithSSLMode("disable")

	fmt.Printf("PostgreSQL DSN:\n  %s\n\n", cfg.PostgresDSN())

	cfg.WithPort(3306)
	fmt.Printf("MySQL DSN:\n  %s\n\n", cfg.MySQLDSN())
}

// TestClientCreation tests client creation without actual connection
func TestClientCreation() {
	fmt.Println("Test 3: Client - Creation (no connection)")
	fmt.Println("------------------------------------------")

	// Test postgres client
	pgClient, err := NewClient("postgres")
	if err != nil {
		fmt.Printf("✗ Failed to create postgres client: %v\n", err)
	} else {
		fmt.Printf("✓ Created postgres client (driver: %s)\n", pgClient.DriverName())
	}

	// Test mysql client
	mysqlClient, err := NewClient("mysql")
	if err != nil {
		fmt.Printf("✗ Failed to create mysql client: %v\n", err)
	} else {
		fmt.Printf("✓ Created mysql client (driver: %s)\n", mysqlClient.DriverName())
	}

	// Test unknown driver
	_, err = NewClient("unknown")
	if err != nil {
		fmt.Printf("✓ Correctly rejected unknown driver: %v\n", err)
	} else {
		fmt.Printf("✗ Should have rejected unknown driver\n")
	}

	fmt.Println()
}
