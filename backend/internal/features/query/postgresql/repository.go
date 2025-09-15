package sqlquery

import (
	"context"
	"fmt"
	"strings"
	"time"

	"postgresus-backend/internal/features/databases"
	"postgresus-backend/internal/features/databases/databases/postgresql"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct{}

func NewRepository() *Repository { return &Repository{} }

func (r *Repository) openPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 2
	cfg.MinConns = 0
	cfg.MaxConnIdleTime = 90 * time.Second
	return pgxpool.NewWithConfig(ctx, cfg)
}

func buildPostgresDSN(pg *postgresql.PostgresqlDatabase) (string, error) {
	if pg == nil {
		return "", fmt.Errorf("postgres config is nil")
	}
	if pg.Database == nil {
		return "", fmt.Errorf("database name is nil")
	}

	db := *pg.Database

	ssl := "disable"
	if pg.IsHttps {
		ssl = "require"
	}

	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		pg.Host,
		pg.Port,
		pg.Username,
		pg.Password,
		db,
		ssl,
	), nil
}

type Result struct {
	Columns   []string
	Rows      [][]any
	RowCount  int
	Truncated bool
	Duration  time.Duration
}

// ExecuteSelect — safe SELECT/CTE with forced LIMIT
// ExecuteSQL — один стейтмент: SELECT/CTE виконуємо з лімітом, інші — через Exec
func (r *Repository) ExecuteSQL(ctx context.Context, dbc *databases.Database, sql string, maxRows int) (*Result, error) {
	// підтримуємо лише PostgreSQL (як і раніше)
	t := strings.ToUpper(string(dbc.Type))
	if t != "POSTGRES" && t != "POSTGRESQL" {
		return nil, fmt.Errorf("only PostgreSQL type is supported, got %q", dbc.Type)
	}

	dsn, err := buildPostgresDSN(dbc.Postgresql)
	if err != nil {
		return nil, err
	}

	pool, err := r.openPool(ctx, dsn)
	if err != nil {
		return nil, err
	}
	defer pool.Close()

	stmt := strings.TrimSpace(sql)
	isSelect := isSelectLike(stmt)

	start := time.Now()
	if isSelect {
		// SELECT / WITH ... SELECT — обгортаємо лімітом
		stmt = ensureLimit(stmt, maxRows)

		rows, err := pool.Query(ctx, stmt)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		fds := rows.FieldDescriptions()
		cols := make([]string, len(fds))
		for i, fd := range fds {
			cols[i] = fd.Name
		}

		outRows := make([][]any, 0, 64)
		rowCount := 0
		for rows.Next() {
			vals, err := rows.Values()
			if err != nil {
				return nil, err
			}
			dst := make([]any, len(vals))
			copy(dst, vals)
			outRows = append(outRows, dst)
			rowCount++
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}

		return &Result{
			Columns:   cols,
			Rows:      outRows,
			RowCount:  rowCount,        // кількість повернутих рядків
			Truncated: rowCount >= maxRows,
			Duration:  time.Since(start),
		}, nil
	}

	// DML/DDL: UPDATE/INSERT/DELETE/CREATE/... — просто виконуємо
	tag, err := pool.Exec(ctx, stmt)
	if err != nil {
		return nil, err
	}

	// фронту достатньо часу виконання; таблиця буде порожня
	// (за бажанням можна додати RowsAffected в DTO пізніше)
	return &Result{
		Columns:   []string{},
		Rows:      [][]any{},
		RowCount:  int(tag.RowsAffected()), // для інфи: скільки рядків зачеплено
		Truncated: false,
		Duration:  time.Since(start),
	}, nil
}

// дуже проста евристика: SELECT або WITH без DML-ключових слів
func isSelectLike(sql string) bool {
	up := strings.ToUpper(strings.TrimSpace(sql))
	if strings.HasPrefix(up, "SELECT ") {
		return true
	}
	if strings.HasPrefix(up, "WITH ") {
		// якщо це WITH ... INSERT/UPDATE/DELETE/... — вважаємо не-SELECT
		for _, k := range []string{" INSERT ", " UPDATE ", " DELETE ", " CREATE ", " ALTER ", " DROP ", " TRUNCATE ", " GRANT ", " REVOKE "} {
			if strings.Contains(up, k) {
				return false
			}
		}
		return true
	}
	return false
}

// ensureLimit — guaranteed to wrap in a subscript with LIMIT
func ensureLimit(sql string, maxRows int) string {
	s := strings.TrimSpace(sql)
	if maxRows <= 0 {
		maxRows = 1000
	}
	return fmt.Sprintf("SELECT * FROM (%s) __q LIMIT %d", s, maxRows)
}