package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"time"

	"github.com/appclacks/maizai/internal/database/queries"
	"github.com/exaring/otelpgx"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type Database struct {
	conn    *pgxpool.Pool
	queries *queries.Queries
}

var CleanupQueries = []string{
	"TRUNCATE context_message CASCADE",
	"TRUNCATE context_source CASCADE",
	"TRUNCATE context CASCADE",
	"TRUNCATE document_chunk CASCADE",
	"TRUNCATE document CASCADE",
	"TRUNCATE schema_migrations CASCADE",
}

func New(config Configuration) (*Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	connectionString := fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=%s", config.Host, config.Port, config.Username, config.Database, config.Password, config.SSLMode)
	cfg, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return nil, err
	}
	cfg.ConnConfig.Tracer = otelpgx.NewTracer()
	conn, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	queries := queries.New(conn)

	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("fail to create source  migration driver: %w", err)
	}

	db := stdlib.OpenDB(*conn.Config().ConnConfig)

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("fail to create postgres migration driver: %w", err)
	}
	m, err := migrate.NewWithInstance(
		"iofs",
		source,
		"postgres",
		driver)

	if err != nil {
		return nil, fmt.Errorf("fail to instantiate migrations: %w", err)
	}
	slog.Info("Applying databases migrations")
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return nil, fmt.Errorf("fail to apply migrations: %w", err)
	}
	slog.Info("Migrations applied")
	return &Database{
		conn:    conn,
		queries: queries,
	}, nil
}

func (d *Database) Exec(query string) (sql.Result, error) {
	db := stdlib.OpenDB(*d.conn.Config().ConnConfig)
	return db.Exec(query)
}

func (d *Database) Stop() {
	d.conn.Close()
}

// crap code
func pgxID(id string) pgtype.UUID {
	googleID := uuid.MustParse(id)
	var uuid pgtype.UUID
	bytes, err := googleID.MarshalBinary()
	if err != nil {
		panic(err)
	}
	for i, b := range bytes { //nolint
		uuid.Bytes[i] = b
	}
	uuid.Valid = true
	return uuid
}

func pgxText(s string) pgtype.Text {
	return pgtype.Text{
		String: s,
		Valid:  true,
	}
}

func pgxTime(t time.Time) pgtype.Timestamp {
	return pgtype.Timestamp{
		Time:  t,
		Valid: true,
	}
}

func (c *Database) beginTx(ctx context.Context, options pgx.TxOptions) (pgx.Tx, *queries.Queries, func(), error) {
	tx, err := c.conn.BeginTx(ctx, options)
	if err != nil {
		return nil, nil, nil, err
	}
	qtx := c.queries.WithTx(tx)
	rollbackFn := func() {
		err := tx.Rollback(ctx)
		if err != nil && err != pgx.ErrTxClosed {
			slog.Error(err.Error())
		}

	}
	return tx, qtx, rollbackFn, nil
}
