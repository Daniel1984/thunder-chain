package peernode

import (
	"context"
	"database/sql"
	"time"

	"com.perkunas/internal/db"
)

type Peer struct {
	ID          int    `json:"id,omitempty" db:"id,omitempty"`
	IP          string `json:"ip" db:"ip"`
	Port        string `json:"port" db:"port"`
	IsBootstrap bool   `json:"is_bootstrap" db:"is_bootstrap"`
	IsActive    bool   `json:"is_active" db:"is_active"`
	LastSeen    int64  `json:"last_seen" db:"last_seen"`
}

type Model struct {
	DB *db.DB
}

func (m *Model) MarkInactive(ctx context.Context) (sql.Result, error) {
	cutoff := time.Now().Add(-5 * time.Minute).Unix()

	query := `
		UPDATE peers
		SET is_active = FALSE
		WHERE last_seen < ? AND is_active = TRUE
	`

	return m.DB.WriteDB.ExecContext(ctx, query, cutoff)
}

func (m *Model) Insert(ctx context.Context, p Peer) error {
	query := `
		INSERT OR REPLACE INTO peers
		(ip, port, is_active, is_bootstrap, last_seen)
		VALUES (:ip, :port, :is_active, :is_bootstrap, :last_seen)
  `

	_, err := m.DB.WriteDB.NamedExecContext(ctx, query, p)
	return err
}

func (m *Model) QueryActive(ctx context.Context) ([]Peer, error) {
	query := `
		SELECT ip, port, is_bootstrap, is_active, last_seen
		FROM peers
		WHERE is_active = 1
		AND is_bootstrap = 0
	`

	res := []Peer{}
	err := m.DB.ReadDB.SelectContext(ctx, &res, query)
	return res, err
}

func (m *Model) ActiveCount(ctx context.Context) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM peers
		WHERE is_active = 1
		AND is_bootstrap = 0
	`

	var count int
	err := m.DB.ReadDB.GetContext(ctx, &count, query, 0)
	return count, err
}
