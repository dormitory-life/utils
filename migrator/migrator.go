package migrator

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Migration struct {
	Version int
	Name    string
	SQL     string
}

func MigrateDB(connectionString, migrationsPath string) error {
	log.Println("Migrations started to run...")

	// Check directory existence
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		return fmt.Errorf("migrations path does not exist: %s", migrationsPath)
	}

	// Connecting to db
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return fmt.Errorf("error connecting to db: %w", err)
	}

	defer db.Close()

	// Check connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("error ping db: %w", err)
	}

	// Create migrations versions table
	if err := createMigrationsTable(db, InitMigrationVersionsTableQuery); err != nil {
		return fmt.Errorf("error creating migrations table: %w", err)
	}

	// Get applied migrations
	alreadyApplied, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Get migrations from migrations directory
	migrations, err := readMigrationsFromPath(migrationsPath)
	if err != nil {
		return fmt.Errorf("error reading migrations: %w", err)
	}

	if len(migrations) == 0 {
		log.Println("No migrations found")
	}

	// Apply not applied migrations
	appliedCounter := 0

	for _, migration := range migrations {
		if alreadyApplied[migration.Version] {
			log.Printf("Migration %s already applied", migration.Name)
			continue
		}

		if err := applyMigration(db, migration); err != nil {
			return fmt.Errorf("error apply migration: %w", err)
		}

		appliedCounter++
	}

	log.Printf("Migrations applied %d", appliedCounter)
	return nil
}

func createMigrationsTable(db *sql.DB, query string) error {
	log.Println("Init migrations table...")
	if _, err := db.Exec(query); err != nil {
		return fmt.Errorf("error init table: %w", err)
	}

	return nil
}

func readMigrationsFromPath(migrationsPath string) ([]Migration, error) {
	log.Println("Reading migrations from path...")
	entries, err := os.ReadDir(migrationsPath)
	if err != nil {
		return nil, fmt.Errorf("error reading files from path: %w", err)
	}

	migrations := make([]Migration, 0)

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".up.sql") {
			log.Printf("incorrect migration name: %s", entry.Name())
			continue
		}

		migrationVersion, err := strconv.Atoi(strings.Split(entry.Name(), "_")[0])
		if err != nil {
			return nil, fmt.Errorf("invalid migration filename %s: %w", entry.Name(), err)
		}

		sqlBytes, err := os.ReadFile(filepath.Join(migrationsPath, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read migration %s: %w", filepath.Join(migrationsPath, entry.Name()), err)
		}

		migrationName := strings.TrimSuffix(entry.Name(), ".up.sql")

		migrations = append(migrations, Migration{
			Version: migrationVersion,
			Name:    migrationName,
			SQL:     string(sqlBytes),
		})
	}

	return migrations, nil
}

func applyMigration(db *sql.DB, migration Migration) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}

	defer tx.Rollback()

	if _, err := tx.Exec(migration.SQL); err != nil {
		return fmt.Errorf("error executing migration %s: %w", migration.Name, err)
	}

	if _, err := tx.Exec(InsertMigrationVersionQuery, migration.Version, migration.Name); err != nil {
		return fmt.Errorf("error inserting version %d: %w", migration.Version, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction for version %s: %w", migration.Version, err)
	}

	return nil
}

func getAppliedMigrations(db *sql.DB) (map[int]bool, error) {
	rows, err := db.Query(GetAppliedMigrationsQuery)
	if err != nil {
		return nil, fmt.Errorf("error getting applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[int]bool)
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}
	return applied, nil
}
