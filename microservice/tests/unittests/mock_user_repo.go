package unittests

import (
	"context"

	"review-manager/internal/repository"
)

// MockUserRepo - реализация UserRepo для юнит-тестов
type MockUserRepo struct {
	Users map[string]repository.User
}

func NewMockUserRepo() *MockUserRepo {
	return &MockUserRepo{
		Users: make(map[string]repository.User),
	}
}

// Заполняем интерфейс
func (m *MockUserRepo) GetByID(_ context.Context, userID string) (repository.User, error) {
	u, ok := m.Users[userID]
	if !ok {
		return repository.User{}, repository.ErrUserNotFound
	}
	return u, nil
}

func (m *MockUserRepo) SetIsActive(_ context.Context, userID string, isActive bool) (repository.User, error) {
	u, ok := m.Users[userID]
	if !ok {
		return repository.User{}, repository.ErrUserNotFound
	}
	u.IsActive = isActive
	m.Users[userID] = u
	return u, nil
}

func (m *MockUserRepo) GetReviewPrs(_ context.Context, _ string) ([]repository.PullRequestShort, error) {
	// для тестов PRService не используется
	return nil, nil
}

func (m *MockUserRepo) FindActiveInTeamExcept(
	_ context.Context,
	teamID int,
	excludeUserID string,
	limit int,
) ([]repository.User, error) {
	var res []repository.User
	for _, u := range m.Users {
		if u.TeamID != teamID {
			continue
		}
		if !u.IsActive {
			continue
		}
		if u.ID == excludeUserID {
			continue
		}
		res = append(res, u)
		if len(res) == limit {
			break
		}
	}
	return res, nil
}

func (m *MockUserRepo) UpsertTeamMembers(_ context.Context, teamID int, members []repository.User) error {
	for _, u := range members {
		u.TeamID = teamID
		m.Users[u.ID] = u
	}
	return nil
}

func (m *MockUserRepo) ListByTeam(_ context.Context, teamID int) ([]repository.User, error) {
	var res []repository.User
	for _, u := range m.Users {
		if u.TeamID == teamID {
			res = append(res, u)
		}
	}
	return res, nil
}

func (m *MockUserRepo) GetReviewerStats(_ context.Context) ([]repository.ReviewerStatRow, error) {
	return nil, nil
}

// compile-time проверка соответствия интерфейсу
var _ repository.UserRepo = (*MockUserRepo)(nil)
