package mysql

import (
	"context"
	"database/sql"
)

type AnalyticsRepository struct{ db *sql.DB }

func NewAnalyticsRepository(db *sql.DB) *AnalyticsRepository { return &AnalyticsRepository{db: db} }

type TeamStats struct {
	TeamID       int64  `json:"team_id"`
	TeamName     string `json:"team_name"`
	MemberCount  int    `json:"member_count"`
	DoneLastWeek int    `json:"done_last_week"`
}

// TeamStatsLastWeek — JOIN 3+ таблиц + агрегация
func (r *AnalyticsRepository) TeamStatsLastWeek(ctx context.Context) ([]*TeamStats, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT t.id AS team_id, t.name AS team_name,
			COUNT(DISTINCT tm.user_id) AS member_count,
			COUNT(DISTINCT CASE
				WHEN tk.status = 'done' AND tk.updated_at >= NOW() - INTERVAL 7 DAY
				THEN tk.id END) AS done_last_week
		FROM teams t
		LEFT JOIN team_members tm ON tm.team_id = t.id
		LEFT JOIN tasks tk ON tk.team_id = t.id
		GROUP BY t.id, t.name
	`)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*TeamStats
	for rows.Next() {
		s := &TeamStats{}
		if err := rows.Scan(&s.TeamID, &s.TeamName, &s.MemberCount, &s.DoneLastWeek); err != nil {
			return nil, err
		}
		result = append(result, s)

	}

	return result, rows.Err()
}

type TopCreator struct {
	TeamID    int64  `json:"team_id"`
	UserID    int64  `json:"user_id"`
	UserName  string `json:"user_name"`
	TaskCount int    `json:"task_count"`
	Rank      int    `json:"rank"`
}

// TopCreatorsPerTeam — оконная функция RANK() OVER (PARTITION BY ...), топ-3 за месяц
func (r *AnalyticsRepository) TopCreatorsPerTeam(ctx context.Context) ([]*TopCreator, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT team_id, user_id, user_name, task_count, rnk FROM (
			SELECT
				tk.team_id,
				u.id        AS user_id,
				u.name      AS user_name,
				COUNT(*)    AS task_count,
				RANK() OVER (PARTITION BY tk.team_id ORDER BY COUNT(*) DESC) AS rnk
			FROM tasks tk
			JOIN users u ON u.id = tk.created_by
			WHERE tk.created_at >= NOW() - INTERVAL 30 DAY
			GROUP BY tk.team_id, u.id, u.name
		) ranked
		WHERE rnk <= 3
	`)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*TopCreator
	for rows.Next() {
		c := &TopCreator{}
		if err := rows.Scan(&c.TeamID, &c.UserID, &c.UserName, &c.TaskCount, &c.Rank); err != nil {
			return nil, err
		}
		result = append(result, c)
	}

	return result, rows.Err()
}

// InvalidAssignees — задачи, где assignee не является членом команды (NOT EXISTS)
func (r *AnalyticsRepository) InvalidAssignees(ctx context.Context) ([]int64, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT t.id FROM tasks t
		WHERE t.assignee_id IS NOT NULL
		  AND NOT EXISTS (
			SELECT 1 FROM team_members tm
			WHERE tm.team_id = t.team_id
			  AND tm.user_id = t.assignee_id
		  )
	`)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, rows.Err()
}
