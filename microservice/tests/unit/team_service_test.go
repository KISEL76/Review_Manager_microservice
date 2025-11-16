package unit

import (
	"context"
	"testing"

	"review-manager/internal/dto"
	"review-manager/internal/repository"
	"review-manager/internal/service"
)

func newTestTeamService() (*service.TeamService, *MockTeamRepo, *MockUserRepo) {
	teamRepo := NewMockTeamRepo()
	userRepo := NewMockUserRepo()
	txMgr := &MockTxManager{}

	svc := service.NewTeamService(teamRepo, userRepo, txMgr)
	return svc, teamRepo, userRepo
}

// Проверяем, что TeamAdd создает команду и апсертит участников в репозиторий пользователей
func TestTeamService_TeamAdd_CreatesTeamAndMembers(t *testing.T) {
	svc, _, userRepo := newTestTeamService()

	req := dto.TeamAddRequest{
		TeamName: "backend",
		Members: []dto.TeamMember{
			{UserID: "u1", Username: "Misha", IsActive: true},
			{UserID: "u2", Username: "Dima", IsActive: false},
		},
	}

	resp, err := svc.TeamAdd(context.Background(), req)
	if err != nil {
		t.Fatalf("TeamAdd вернул ошибку: %v", err)
	}

	if resp.Team.TeamName != "backend" {
		t.Fatalf("ожидали team_name=backend, получили %s", resp.Team.TeamName)
	}
	if len(resp.Team.Members) != 2 {
		t.Fatalf("ожидали 2 участника в ответе, получили %d", len(resp.Team.Members))
	}

	u1, ok1 := userRepo.Users["u1"]
	u2, ok2 := userRepo.Users["u2"]
	if !ok1 || !ok2 {
		t.Fatalf("ожидали что пользователи u1 и u2 будут в репозитории, ok1=%v ok2=%v", ok1, ok2)
	}

	if !u1.IsActive {
		t.Fatalf("ожидали что u1.IsActive=true, получили false")
	}
	if u2.IsActive {
		t.Fatalf("ожидали что u2.IsActive=false, получили true")
	}
}

// Проверяем, что TeamGet возвращает команду и только тех участников, у кого team_id этой команды
func TestTeamService_TeamGet_ReturnsTeamWithMembers(t *testing.T) {
	svc, teamRepo, userRepo := newTestTeamService()

	// Готовим команду и участников
	teamRepo.TeamsByName["payments"] = repository.Team{ID: 10, Name: "payments"}
	userRepo.Users["u1"] = repository.User{ID: "u1", Username: "Alina", TeamID: 10, IsActive: true}
	userRepo.Users["u2"] = repository.User{ID: "u2", Username: "Roma", TeamID: 10, IsActive: false}
	userRepo.Users["u3"] = repository.User{ID: "u3", Username: "Gleb", TeamID: 99, IsActive: true} // другая команда

	resp, err := svc.TeamGet(context.Background(), "payments")
	if err != nil {
		t.Fatalf("TeamGet вернул ошибку: %v", err)
	}

	if resp.TeamName != "payments" {
		t.Fatalf("ожидали team_name=payments, получили %s", resp.TeamName)
	}
	if len(resp.Members) != 2 {
		t.Fatalf("ожидали 2 участника для команды payments, получили %d", len(resp.Members))
	}

	ids := map[string]bool{}
	for _, m := range resp.Members {
		ids[m.UserID] = true
	}
	if !ids["u1"] || !ids["u2"] {
		t.Fatalf("ожидали что в команде будут u1 и u2, получили %v", ids)
	}
}
