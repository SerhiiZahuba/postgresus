package sqlquery

import (
	"time"

	"github.com/google/uuid"
)

// ExecuteRequest — query to execute SELECT/CTE
type ExecuteRequest struct {
	DatabaseID uuid.UUID `json:"databaseId" binding:"required"`
	SQL        string    `json:"sql"        binding:"required"`
	MaxRows    int       `json:"maxRows"`
	TimeoutSec int       `json:"timeoutSec"`
}

type ExecuteResponse struct {
	Columns     []string      `json:"columns"`
	Rows        [][]any       `json:"rows"`
	RowCount    int           `json:"rowCount"`
	Truncated   bool          `json:"truncated"`
	ExecutionMs int64         `json:"executionMs"`
}

// Simple structure for history/logs (for future in the DB)
type QueryAudit struct {
	ID         uuid.UUID     `json:"id"`
	UserID     uuid.UUID     `json:"user_id"`
	DatabaseID uuid.UUID     `json:"database_id"`
	SQL        string        `json:"sql"`
	RowCount   int           `json:"row_count"`
	Duration   time.Duration `json:"duration"`
	When       time.Time     `json:"when"`
	Ok         bool          `json:"ok"`
	Error      string        `json:"error,omitempty"`
}