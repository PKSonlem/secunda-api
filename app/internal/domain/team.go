package domain

import (
	"context"
	"time"
)

type Role string

const (
	RoleOwner  Role = "owner"
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
)

type Team struct {
	ID        int64     `json:"id"         db:"id"`
	Name      string    `json:"name"       db:"name"`
	CreatedBy int64     `json:"created_by" db:"created_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type TeamMember struct {
	TeamID   int64     `json:"team_id"   db:"team_id"`
	UserID   int64     `json:"user_id"   db:"user_id"`
	Role     Role      `json:"role"      db:"role"`
	JoinedAt time.Time `json:"joined_at" db:"joined_at"`
}

type TeamRepository interface {
	Create(ctx context.Context, team *Team) (int64, error)
	GetByID(ctx context.Context, id int64) (*Team, error)
	ListByUserID(ctx context.Context, userID int64) ([]*Team, error)
	AddMember(ctx context.Context, member *TeamMember) error
	GetMember(ctx context.Context, teamID, userID int64) (*TeamMember, error)
	IsMember(ctx context.Context, teamID, userID int64) (bool, error)
}
