package migrator

const (
	InitMigrationVersionsTableQuery = `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`

	InsertMigrationVersionQuery = `
		INSERT INTO schema_migrations (version, name) VALUES ($1, $2)
	`

	GetAppliedMigrationsQuery = `
		SELECT version FROM schema_migrations
	`
)
