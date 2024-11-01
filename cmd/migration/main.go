package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"slices"
	"sort"
	"strings"

	_ "github.com/jackc/pgx/v4/stdlib"
	migrate "github.com/rubenv/sql-migrate"
)

func main() {
	migrate.SetTable("migrations")
	dir := flag.String("dir", "./migrationsxx", "Specify the path to the migrations directory")
	version := flag.String("version", "", "Specify the version to migrate to")
	flag.Parse()

	pgconnStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		"localhost",
		"5432",
		"genome",
		"genome",
		"tester",
		"disable")
	pgconn, err := connect(context.Background(), pgconnStr)
	if err != nil {
		log.Fatalf("Failed to open connection to database: %v\n", err)
	}
	defer pgconn.Close()

	migrationFiles := readMigrationFiles(*dir)
	migrationsApplied := readAppliedMigrations(pgconn)

	latestAppliedMigration := ""
	if len(migrationsApplied) > 0 {
		latestAppliedMigration = migrationsApplied[len(migrationsApplied)-1]
	}

	if *version == "" {
		applyLatest(pgconn, *dir, migrationFiles, migrationsApplied)
	} else if latestAppliedMigration != "" && latestAppliedMigration > *version {
		downgrade(pgconn, *dir, migrationFiles, migrationsApplied, *version)
	} else if (latestAppliedMigration == "" && *version != "") || latestAppliedMigration < *version {
		upgrade(pgconn, *dir, migrationFiles, migrationsApplied, *version)
	} else {
		fmt.Println("Migrations already up to date")
	}
}

func connect(ctx context.Context, connString string) (*sql.DB, error) {
	// Open the connection with pgx as the driver
	db, err := sql.Open("pgx", connString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Ping the database to ensure connectivity
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func downgrade(conn *sql.DB, migrationDir string, migrationFiles []string, migrationsApplied []string, version string) {
	fmt.Printf("migrate to %s\n", version)
	migrationSource := migrate.FileMigrationSource{Dir: migrationDir}

	for i := len(migrationFiles) - 1; i >= 0; i-- {
		migrateFile := migrationFiles[i]

		if migrateFile == version {
			break
		}

		if !slices.Contains(migrationsApplied, migrateFile) {
			fmt.Printf("Migration %s not applied. Skipping...\n", migrateFile)
			continue
		}

		applied, err := migrate.ExecMax(conn, "postgres", migrationSource, migrate.Down, 1)
		if err != nil {
			fmt.Printf("Failed to apply migration %s: %v\n", migrateFile, err)
			return
		}

		fmt.Printf("Applied migration %s %d times\n", migrateFile, applied)
	}
}

func upgrade(conn *sql.DB, migrationDir string, migrationFiles []string, migrationsApplied []string, version string) {
	fmt.Printf("migrate to %s\n", version)
	migrationSource := migrate.FileMigrationSource{Dir: migrationDir}

	for _, migrateFile := range migrationFiles {
		if slices.Contains(migrationsApplied, migrateFile) {
			fmt.Printf("Migration %s already applied. Skipping...\n", migrateFile)
			continue
		}

		applied, err := migrate.ExecMax(conn, "postgres", migrationSource, migrate.Up, 1)
		if err != nil {
			fmt.Printf("Failed to apply migration %s: %v\n", migrateFile, err)
			return
		}

		fmt.Printf("Applied migration %s %d times\n", migrateFile, applied)

		if migrateFile == version {
			break
		}
	}
}

func applyLatest(conn *sql.DB, migrationDir string, migrationFiles []string, migrationsApplied []string) {
	fmt.Println("migrate to latest")
	migrationSource := migrate.FileMigrationSource{Dir: migrationDir}

	for _, migrateFile := range migrationFiles {
		if slices.Contains(migrationsApplied, migrateFile) {
			fmt.Printf("Migration %s already applied. Skipping...\n", migrateFile)
			continue
		}

		applied, err := migrate.ExecMax(conn, "postgres", migrationSource, migrate.Up, 1)
		if err != nil {
			fmt.Printf("Failed to apply migration %s: %v\n", migrateFile, err)
			return
		}

		fmt.Printf("Applied migration %s %d times\n", migrateFile, applied)
	}
}

func readMigrationFiles(dir string) []string {
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Fatalf("Failed to read migration directory: %v\n", err)
	}

	var migrationFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			migrationFiles = append(migrationFiles, file.Name())
		}
	}

	sort.Strings(migrationFiles)

	return migrationFiles
}

func readAppliedMigrations(conn *sql.DB) []string {
	appliedMigrations := []string{}
	rows, err := conn.Query(`Select id from migrations order by id asc`)
	if err != nil {
		log.Fatalf("Failed to query migrations from database: %v\n", err)
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			log.Fatalf("Failed to read migration ID: %v\n", id)
		}
		appliedMigrations = append(appliedMigrations, id)
	}

	return appliedMigrations
}
