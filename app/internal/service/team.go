package service

import (
	"context"
	"log/slog"

	"github.com/PKSonlem/testtask-secunda-api/internal/domain"
)

type emailSender interface {
	SendInvite(ctx context.Context, email, teamName string) error
}

type TeamService struct {
	teams domain.TeamRepository
	users domain.UserRepository
	email emailSender
}

func NewTeamService(teams domain.TeamRepository, users domain.UserRepository, email emailSender) *TeamService {
	return &TeamService{teams: teams, users: users, email: email}
}

func (s *TeamService) Create(ctx context.Context, userID int64, name string) (*domain.Team, error) {
	team := &domain.Team{Name: name, CreatedBy: userID}

	id, err := s.teams.Create(ctx, team)
	if err != nil {
		return nil, err
	}

	team.ID = id

	return team, nil
}

func (s *TeamService) List(ctx context.Context, userID int64) ([]*domain.Team, error) {
	return s.teams.ListByUserID(ctx, userID)
}

// Invite добавляет участника и отправляет email. Ошибка email не возвращается —
// добавление уже зафиксировано в БД, отправка письма некритична.
func (s *TeamService) Invite(ctx context.Context, callerID, teamID int64, inviteeEmail string) error {
	member, err := s.teams.GetMember(ctx, teamID, callerID)
	if err != nil {
		return domain.ErrForbidden
	}
	if member.Role != domain.RoleOwner && member.Role != domain.RoleAdmin {
		return domain.ErrForbidden
	}

	invitee, err := s.users.GetByEmail(ctx, inviteeEmail)
	if err != nil {
		return domain.ErrNotFound
	}

	if err := s.teams.AddMember(ctx, &domain.TeamMember{
		TeamID: teamID, UserID: invitee.ID, Role: domain.RoleMember,
	}); err != nil {
		return err
	}

	team, err := s.teams.GetByID(ctx, teamID)
	if err != nil {
		return err
	}

	if err := s.email.SendInvite(ctx, inviteeEmail, team.Name); err != nil {
		slog.Warn("send invite email failed", "err", err, "email", inviteeEmail)
	}

	return nil
}
