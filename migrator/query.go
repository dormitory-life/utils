package migrator

const (
	InitMigrationsSchema = `
		CREATE SCHEMA IF NOT EXISTS migrations;
	`

	InitMigrationVersionsTableQuery = `
		CREATE TABLE IF NOT EXISTS migrations.schema_migrations (
			version INTEGER PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`

	InsertMigrationVersionQuery = `
		INSERT INTO migrations.schema_migrations (version, name) VALUES ($1, $2)
	`

	GetAppliedMigrationsQuery = `
		SELECT version FROM migrations.schema_migrations
	`
)
