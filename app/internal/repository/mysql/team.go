package mysql

import (
	"context"
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/PKSonlem/testtask-secunda-api/internal/domain"
)

type TeamRepository struct {
	db    *sql.DB
	txMgr *TxManager
}

func NewTeamRepository(db *sql.DB, txMgr *TxManager) *TeamRepository {
	return &TeamRepository{db: db, txMgr: txMgr}
}

// Create создаёт команду и добавляет создателя как owner в одной транзакции.
func (r *TeamRepository) Create(ctx context.Context, team *domain.Team) (int64, error) {
	var id int64

	err := r.txMgr.WithTx(ctx, func(tx *sql.Tx) error {
		res, err := builder.
			Insert("teams").
			Columns("name", "created_by").
			Values(team.Name, team.CreatedBy).
			RunWith(tx).ExecContext(ctx)

		if err != nil {
			return err
		}

		id, err = res.LastInsertId()
		if err != nil {
			return err
		}

		_, err = builder.
			Insert("team_members").
			Columns("team_id", "user_id", "role").
			Values(id, team.CreatedBy, domain.RoleOwner).
			RunWith(tx).ExecContext(ctx)

		return err
	})

	return id, err
}

func (r *TeamRepository) GetByID(ctx context.Context, id int64) (*domain.Team, error) {
	t := &domain.Team{}

	err := builder.
		Select("id", "name", "created_by", "created_at").
		From("teams").
		Where(sq.Eq{"id": id}).
		RunWith(r.db).QueryRowContext(ctx).
		Scan(&t.ID, &t.Name, &t.CreatedBy, &t.CreatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}

	return t, err
}

func (r *TeamRepository) ListByUserID(ctx context.Context, userID int64) ([]*domain.Team, error) {
	rows, err := builder.
		Select("t.id", "t.name", "t.created_by", "t.created_at").
		From("teams t").
		Join("team_members tm ON t.id = tm.team_id").
		Where(sq.Eq{"tm.user_id": userID}).
		RunWith(r.db).QueryContext(ctx)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []*domain.Team

	for rows.Next() {
		t := &domain.Team{}
		if err := rows.Scan(&t.ID, &t.Name, &t.CreatedBy, &t.CreatedAt); err != nil {
			return nil, err
		}
		teams = append(teams, t)
	}

	return teams, rows.Err()
}

func (r *TeamRepository) AddMember(ctx context.Context, m *domain.TeamMember) error {
	_, err := builder.
		Insert("team_members").
		Columns("team_id", "user_id", "role").
		Values(m.TeamID, m.UserID, m.Role).
		RunWith(r.db).ExecContext(ctx)

	return err
}

func (r *TeamRepository) GetMember(ctx context.Context, teamID, userID int64) (*domain.TeamMember, error) {
	m := &domain.TeamMember{}

	err := builder.
		Select("team_id", "user_id", "role", "joined_at").
		From("team_members").
		Where(sq.Eq{"team_id": teamID, "user_id": userID}).
		RunWith(r.db).QueryRowContext(ctx).
		Scan(&m.TeamID, &m.UserID, &m.Role, &m.JoinedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}

	return m, err
}

func (r *TeamRepository) IsMember(ctx context.Context, teamID, userID int64) (bool, error) {
	var n int

	err := builder.
		Select("COUNT(1)").
		From("team_members").
		Where(sq.Eq{"team_id": teamID, "user_id": userID}).
		RunWith(r.db).QueryRowContext(ctx).Scan(&n)

	return n > 0, err
}
