package mysql

import (
	"context"
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/PKSonlem/testtask-secunda-api/internal/domain"
)

type TaskRepository struct {
	db    *sql.DB
	txMgr *TxManager
}

func NewTaskRepository(db *sql.DB, txMgr *TxManager) *TaskRepository {
	return &TaskRepository{db: db, txMgr: txMgr}
}

func (r *TaskRepository) Create(ctx context.Context, t *domain.Task) (int64, error) {
	res, err := builder.
		Insert("tasks").
		Columns("title", "description", "status", "team_id", "assignee_id", "created_by").
		Values(t.Title, t.Description, t.Status, t.TeamID, t.AssigneeID, t.CreatedBy).
		RunWith(r.db).ExecContext(ctx)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func (r *TaskRepository) GetByID(ctx context.Context, id int64) (*domain.Task, error) {
	t := &domain.Task{}
	err := builder.
		Select("id", "title", "description", "status", "team_id", "assignee_id", "created_by", "created_at", "updated_at").
		From("tasks").
		Where(sq.Eq{"id": id}).
		RunWith(r.db).QueryRowContext(ctx).
		Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.TeamID, &t.AssigneeID, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}

	return t, err
}

// Update читает старое состояние задачи внутри той же транзакции, чтобы получить
// консистентный снимок для истории и избежать гонки с параллельными обновлениями.
func (r *TaskRepository) Update(ctx context.Context, task *domain.Task, changedBy int64) error {
	return r.txMgr.WithTx(ctx, func(tx *sql.Tx) error {
		old := &domain.Task{}
		err := builder.
			Select("id", "title", "description", "status", "team_id", "assignee_id", "created_by", "created_at", "updated_at").
			From("tasks").
			Where(sq.Eq{"id": task.ID}).
			RunWith(tx).QueryRowContext(ctx).
			Scan(&old.ID, &old.Title, &old.Description, &old.Status, &old.TeamID, &old.AssigneeID, &old.CreatedBy, &old.CreatedAt, &old.UpdatedAt)
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrNotFound
		}
		if err != nil {
			return err
		}

		if _, err = builder.
			Update("tasks").
			Set("title", task.Title).
			Set("description", task.Description).
			Set("status", task.Status).
			Set("assignee_id", task.AssigneeID).
			Set("updated_at", sq.Expr("NOW()")).
			Where(sq.Eq{"id": task.ID}).
			RunWith(tx).ExecContext(ctx); err != nil {
			return err
		}

		return r.writeHistory(ctx, tx, old, task, changedBy)
	})
}

// List строит один cond для COUNT и SELECT — пагинация не разъедется с total.
func (r *TaskRepository) List(ctx context.Context, f domain.TaskFilter) ([]*domain.Task, int, error) {
	cond := sq.And{}
	if f.TeamID != nil {
		cond = append(cond, sq.Eq{"team_id": *f.TeamID})
	}
	if f.Status != nil {
		cond = append(cond, sq.Eq{"status": *f.Status})
	}
	if f.AssigneeID != nil {
		cond = append(cond, sq.Eq{"assignee_id": *f.AssigneeID})
	}

	var total int
	if err := builder.Select("COUNT(1)").From("tasks").Where(cond).
		RunWith(r.db).QueryRowContext(ctx).Scan(&total); err != nil {
		return nil, 0, err
	}

	if f.PageSize == 0 {
		f.PageSize = 20
	}
	if f.Page < 1 {
		f.Page = 1
	}

	rows, err := builder.
		Select("id", "title", "description", "status", "team_id", "assignee_id", "created_by", "created_at", "updated_at").
		From("tasks").
		Where(cond).
		OrderBy("created_at DESC").
		Limit(uint64(f.PageSize)).
		Offset(uint64((f.Page - 1) * f.PageSize)).
		RunWith(r.db).QueryContext(ctx)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tasks []*domain.Task
	for rows.Next() {
		t := &domain.Task{}
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.TeamID,
			&t.AssigneeID, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, 0, err
		}
		tasks = append(tasks, t)
	}

	return tasks, total, rows.Err()
}

func (r *TaskRepository) GetHistory(ctx context.Context, taskID int64) ([]*domain.TaskHistory, error) {
	rows, err := builder.
		Select("id", "task_id", "changed_by", "field_name", "old_value", "new_value", "changed_at").
		From("task_history").
		Where(sq.Eq{"task_id": taskID}).
		OrderBy("changed_at DESC").
		RunWith(r.db).QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*domain.TaskHistory
	for rows.Next() {
		h := &domain.TaskHistory{}
		if err := rows.Scan(&h.ID, &h.TaskID, &h.ChangedBy, &h.FieldName, &h.OldValue, &h.NewValue, &h.ChangedAt); err != nil {
			return nil, err
		}
		history = append(history, h)
	}

	return history, rows.Err()
}

// writeHistory пишет только изменившиеся поля, а не полный снимок —
// чтобы история не засорялась записями без реального изменения.
func (r *TaskRepository) writeHistory(ctx context.Context, tx *sql.Tx, old, new *domain.Task, changedBy int64) error {
	fields := []struct{ name, old, new string }{
		{"title", old.Title, new.Title},
		{"description", old.Description, new.Description},
		{"status", string(old.Status), string(new.Status)},
	}

	for _, f := range fields {
		if f.old == f.new {
			continue
		}
		if _, err := builder.
			Insert("task_history").
			Columns("task_id", "changed_by", "field_name", "old_value", "new_value").
			Values(new.ID, changedBy, f.name, f.old, f.new).
			RunWith(tx).ExecContext(ctx); err != nil {
			return err
		}
	}

	return nil
}
