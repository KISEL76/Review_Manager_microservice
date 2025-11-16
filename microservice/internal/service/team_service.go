package service

import (
	"context"

	"review-manager/internal/dto"
	"review-manager/internal/repository"
)

type TeamService struct {
	Teams repository.TeamRepo
	Users repository.UserRepo
	Tx    repository.TxManager
}

func NewTeamService(teams repository.TeamRepo, users repository.UserRepo, tx repository.TxManager) *TeamService {
	return &TeamService{
		Teams: teams,
		Users: users,
		Tx:    tx,
	}
}

// логика /team/add - создаем команду и участников в одной транзакции.
func (s *TeamService) TeamAdd(ctx context.Context, req dto.TeamAddRequest) (dto.TeamAddResponse, error) {
	var teamRow repository.Team

	err := s.Tx.WithinTx(ctx, func(ctx context.Context) error {
		var err error

		teamRow, err = s.Teams.Create(ctx, req.TeamName)
		if err != nil {
			return err
		}

		members := make([]repository.User, 0, len(req.Members))
		for _, m := range req.Members {
			members = append(members, repository.User{
				ID:       m.UserID,
				Username: m.Username,
				TeamID:   teamRow.ID,
				IsActive: m.IsActive,
			})
		}

		if err := s.Users.UpsertTeamMembers(ctx, teamRow.ID, members); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return dto.TeamAddResponse{}, err
	}

	dtoTeam := dto.Team{
		TeamName: teamRow.Name,
		Members:  append([]dto.TeamMember(nil), req.Members...),
	}

	return dto.TeamAddResponse{Team: dtoTeam}, nil
}

// логика team/get - выводим команду со всеми участниками
func (s *TeamService) TeamGet(ctx context.Context, teamName string) (dto.Team, error) {
	teamRow, err := s.Teams.GetByName(ctx, teamName)
	if err != nil {
		return dto.Team{}, err
	}

	users, err := s.Users.ListByTeam(ctx, teamRow.ID)
	if err != nil {
		return dto.Team{}, err
	}

	members := make([]dto.TeamMember, 0, len(users))
	for _, u := range users {
		members = append(members, dto.TeamMember{
			UserID:   u.ID,
			Username: u.Username,
			IsActive: u.IsActive,
		})
	}

	return dto.Team{
		TeamName: teamRow.Name,
		Members:  members,
	}, nil
}
