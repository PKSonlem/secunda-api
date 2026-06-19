package mysql

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"
)

var builder = sq.StatementBuilder.PlaceholderFormat(sq.Question)

type TxManager struct{ db *sql.DB }

func NewTxManager(db *sql.DB) *TxManager { return &TxManager{db: db} }

func (m *TxManager) WithTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}
