package unit

import (
	"context"

	"review-manager/internal/repository"
)

// in-memory реализация UserRepo для тестов
type MockUserRepo struct {
	Users     map[string]repository.User
	ReviewPRs map[string][]repository.PullRequestShort
	StatsRows []repository.ReviewerStatRow
}

func NewMockUserRepo() *MockUserRepo {
	return &MockUserRepo{
		Users:     make(map[string]repository.User),
		ReviewPRs: make(map[string][]repository.PullRequestShort),
		StatsRows: nil,
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

func (m *MockUserRepo) GetReviewPrs(_ context.Context, userID string) ([]repository.PullRequestShort, error) {
	prs, ok := m.ReviewPRs[userID]
	if !ok {
		return nil, nil
	}

	out := make([]repository.PullRequestShort, len(prs))
	copy(out, prs)
	return out, nil
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
	if len(m.StatsRows) == 0 {
		return nil, nil
	}
	out := make([]repository.ReviewerStatRow, len(m.StatsRows))
	copy(out, m.StatsRows)
	return out, nil
}

// compile-time проверка соответствия интерфейсу
var _ repository.UserRepo = (*MockUserRepo)(nil)
