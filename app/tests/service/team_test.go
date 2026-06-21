package servicetest

import (
	"context"
	"testing"

	"github.com/PKSonlem/testtask-secunda-api/internal/domain"
	"github.com/PKSonlem/testtask-secunda-api/internal/service"
	"github.com/PKSonlem/testtask-secunda-api/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeamCreate(t *testing.T) {
	ctx := context.Background()

	teamRepo := &mocks.TeamRepo{
		CreateFn: func(_ context.Context, team *domain.Team) (int64, error) {
			assert.Equal(t, "My Team", team.Name)
			assert.Equal(t, int64(1), team.CreatedBy)
			return 5, nil
		},
	}

	svc := service.NewTeamService(teamRepo, &mocks.UserRepo{}, &mocks.EmailSender{})
	team, err := svc.Create(ctx, 1, "My Team")

	require.NoError(t, err)
	assert.Equal(t, int64(5), team.ID)
	assert.Equal(t, "My Team", team.Name)
}

func TestTeamList(t *testing.T) {
	ctx := context.Background()
	expected := []*domain.Team{{ID: 1, Name: "Alpha"}, {ID: 2, Name: "Beta"}}

	teamRepo := &mocks.TeamRepo{
		ListByUserIDFn: func(_ context.Context, userID int64) ([]*domain.Team, error) {
			assert.Equal(t, int64(3), userID)
			return expected, nil
		},
	}

	svc := service.NewTeamService(teamRepo, &mocks.UserRepo{}, &mocks.EmailSender{})
	teams, err := svc.List(ctx, 3)

	require.NoError(t, err)
	assert.Equal(t, expected, teams)
}

func TestInvite(t *testing.T) {
	ctx := context.Background()

	invitee := &domain.User{ID: 20, Email: "bob@example.com"}
	team := &domain.Team{ID: 1, Name: "Alpha"}

	t.Run("owner_can_invite", func(t *testing.T) {
		emailCalled := false
		svc := buildInviteService(t, domain.RoleOwner, invitee, team, func() { emailCalled = true })

		err := svc.Invite(ctx, 1, 1, "bob@example.com")

		require.NoError(t, err)
		assert.True(t, emailCalled)
	})

	t.Run("admin_can_invite", func(t *testing.T) {
		svc := buildInviteService(t, domain.RoleAdmin, invitee, team, nil)
		require.NoError(t, svc.Invite(ctx, 1, 1, "bob@example.com"))
	})

	t.Run("member_cannot_invite", func(t *testing.T) {
		teamRepo := &mocks.TeamRepo{
			GetMemberFn: func(_ context.Context, _, _ int64) (*domain.TeamMember, error) {
				return &domain.TeamMember{Role: domain.RoleMember}, nil
			},
		}
		svc := service.NewTeamService(teamRepo, &mocks.UserRepo{}, &mocks.EmailSender{})
		assert.ErrorIs(t, svc.Invite(ctx, 1, 1, "bob@example.com"), domain.ErrForbidden)
	})

	t.Run("caller_not_in_team", func(t *testing.T) {
		teamRepo := &mocks.TeamRepo{
			GetMemberFn: func(_ context.Context, _, _ int64) (*domain.TeamMember, error) {
				return nil, domain.ErrNotFound
			},
		}
		svc := service.NewTeamService(teamRepo, &mocks.UserRepo{}, &mocks.EmailSender{})
		assert.ErrorIs(t, svc.Invite(ctx, 99, 1, "bob@example.com"), domain.ErrForbidden)
	})

	t.Run("invitee_not_registered", func(t *testing.T) {
		teamRepo := &mocks.TeamRepo{
			GetMemberFn: func(_ context.Context, _, _ int64) (*domain.TeamMember, error) {
				return &domain.TeamMember{Role: domain.RoleOwner}, nil
			},
		}
		userRepo := &mocks.UserRepo{
			GetByEmailFn: func(_ context.Context, _ string) (*domain.User, error) {
				return nil, domain.ErrNotFound
			},
		}
		svc := service.NewTeamService(teamRepo, userRepo, &mocks.EmailSender{})
		assert.ErrorIs(t, svc.Invite(ctx, 1, 1, "ghost@example.com"), domain.ErrNotFound)
	})
}

func buildInviteService(t *testing.T, role domain.Role, invitee *domain.User, team *domain.Team, onEmail func()) *service.TeamService {
	t.Helper()
	teamRepo := &mocks.TeamRepo{
		GetMemberFn: func(_ context.Context, _, _ int64) (*domain.TeamMember, error) {
			return &domain.TeamMember{Role: role}, nil
		},
		AddMemberFn: func(_ context.Context, _ *domain.TeamMember) error { return nil },
		GetByIDFn:   func(_ context.Context, _ int64) (*domain.Team, error) { return team, nil },
	}
	userRepo := &mocks.UserRepo{
		GetByEmailFn: func(_ context.Context, _ string) (*domain.User, error) { return invitee, nil },
	}
	emailMock := &mocks.EmailSender{
		SendInviteFn: func(_ context.Context, _, _ string) error {
			if onEmail != nil {
				onEmail()
			}
			return nil
		},
	}
	return service.NewTeamService(teamRepo, userRepo, emailMock)
}
