package unit

import (
	"context"

	"review-manager/internal/repository"
)

// in-memory реализация TeamRepo для тестов
type MockTeamRepo struct {
	TeamsByName map[string]repository.Team
	nextID      int
}

func NewMockTeamRepo() *MockTeamRepo {
	return &MockTeamRepo{
		TeamsByName: make(map[string]repository.Team),
		nextID:      0,
	}
}

// Заполняем интерфейс
func (m *MockTeamRepo) Create(_ context.Context, teamName string) (repository.Team, error) {
	if t, ok := m.TeamsByName[teamName]; ok {
		return t, repository.ErrTeamExists
	}

	m.nextID++
	t := repository.Team{
		ID:   m.nextID,
		Name: teamName,
	}
	m.TeamsByName[teamName] = t
	return t, nil
}

func (m *MockTeamRepo) GetByName(_ context.Context, teamName string) (repository.Team, error) {
	t, ok := m.TeamsByName[teamName]
	if !ok {
		return repository.Team{}, repository.ErrTeamNotFound
	}
	return t, nil
}

// compile time проверка что мок реализует интерфейс TeamRepo
var _ repository.TeamRepo = (*MockTeamRepo)(nil)
