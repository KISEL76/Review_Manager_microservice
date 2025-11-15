package unittests

import (
	"context"
	"time"

	"review-manager/internal/repository"
)

// Простая in-memory реализация на мапах PrRepo для тестов
type MockPrRepo struct {
	PRs       map[string]repository.PullRequest
	Reviewers map[string][]string
}

func NewMockPrRepo() *MockPrRepo {
	return &MockPrRepo{
		PRs:       make(map[string]repository.PullRequest),
		Reviewers: make(map[string][]string),
	}
}

// Заполнение интерфейса
func (m *MockPrRepo) Exists(_ context.Context, prID string) (bool, error) {
	_, ok := m.PRs[prID]
	return ok, nil
}

func (m *MockPrRepo) Insert(_ context.Context, pr repository.PullRequest) error {
	m.PRs[pr.ID] = pr
	return nil
}

func (m *MockPrRepo) AddReviewers(_ context.Context, prID string, reviewerIDs []string) error {
	if len(reviewerIDs) == 0 {
		return nil
	}
	cp := make([]string, len(reviewerIDs))
	copy(cp, reviewerIDs)
	m.Reviewers[prID] = cp
	return nil
}

func (m *MockPrRepo) GetWithReviewers(_ context.Context, prID string) (repository.PullRequest, []string, error) {
	pr, ok := m.PRs[prID]
	if !ok {
		return repository.PullRequest{}, nil, repository.ErrPRNotFound
	}
	revs := append([]string(nil), m.Reviewers[prID]...)
	return pr, revs, nil
}

func (m *MockPrRepo) SetMerged(_ context.Context, prID string, mergedAt time.Time) (repository.PullRequest, error) {
	pr, ok := m.PRs[prID]
	if !ok {
		return repository.PullRequest{}, repository.ErrPRNotFound
	}
	pr.StatusID = repository.StatusMerged
	if pr.MergedAt == nil {
		pr.MergedAt = &mergedAt
	}
	m.PRs[prID] = pr
	return pr, nil
}

func (m *MockPrRepo) ReplaceReviewer(_ context.Context, prID, oldReviewerID, newReviewerID string) error {
	revs, ok := m.Reviewers[prID]
	if !ok {
		return repository.ErrReviewerNotSet
	}
	replaced := false
	for i, id := range revs {
		if id == oldReviewerID {
			revs[i] = newReviewerID
			replaced = true
			break
		}
	}
	if !replaced {
		return repository.ErrReviewerNotSet
	}
	m.Reviewers[prID] = revs
	return nil
}
