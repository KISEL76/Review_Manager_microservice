package unit

import (
	"context"
	"errors"
	"testing"
	"time"

	"review-manager/internal/dto"
	"review-manager/internal/repository"
	"review-manager/internal/service"
)

func newTestPRService() (*service.PRService, *MockPrRepo, *MockUserRepo) {
	prRepo := NewMockPrRepo()
	userRepo := NewMockUserRepo()
	txMgr := &MockTxManager{}

	svc := service.NewPRService(prRepo, userRepo, txMgr)
	return svc, prRepo, userRepo
}

// Проверяем, что при создании PR назначаются ревьюверы из команды автора,
// что их не больше двух, и сам автор не попадает в список
func TestPRService_Create_AssignsReviewersFromAuthorsTeam(t *testing.T) {
	svc, _, userRepo := newTestPRService()

	// Команда 1: автор + два активных коллеги
	userRepo.Users["u1"] = repository.User{ID: "u1", Username: "Serega", TeamID: 1, IsActive: true} // автор
	userRepo.Users["u2"] = repository.User{ID: "u2", Username: "Vova", TeamID: 1, IsActive: true}
	userRepo.Users["u3"] = repository.User{ID: "u3", Username: "Boba", TeamID: 1, IsActive: true}

	// Команда 2: активный, но другой team_id
	userRepo.Users["u4"] = repository.User{ID: "u4", Username: "David", TeamID: 2, IsActive: true}

	req := dto.CreatePrRequest{
		PullRequestID:   "pr-1",
		PullRequestName: "Add search",
		AuthorID:        "u1",
	}

	resp, err := svc.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("Create вернул ошибку: %v", err)
	}

	reviewers := resp.PR.AssignedReviewers
	if len(reviewers) == 0 || len(reviewers) > 2 {
		t.Fatalf("ожидали 1 или 2 ревьювера, получили %d", len(reviewers))
	}

	for _, id := range reviewers {
		if id == "u1" {
			t.Fatalf("автор не должен быть ревьювером, но в списке есть %s", id)
		}
		u := userRepo.Users[id]
		if u.TeamID != 1 {
			t.Fatalf("ревьювер должен быть из команды автора team_id=1, но у пользователя %s team_id=%d", id, u.TeamID)
		}
	}
}

// Проверяем, что если кроме автора нет активных коллег, ревьюверы не назначаются.
func TestPRService_Create_NoReviewersWhenNoCandidates(t *testing.T) {
	svc, _, userRepo := newTestPRService()

	// Только автор в своей команде
	userRepo.Users["u1"] = repository.User{ID: "u1", Username: "Lisa", TeamID: 1, IsActive: true}

	req := dto.CreatePrRequest{
		PullRequestID:   "pr-2",
		PullRequestName: "Solo change",
		AuthorID:        "u1",
	}

	resp, err := svc.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("Create вернул ошибку: %v", err)
	}

	if len(resp.PR.AssignedReviewers) != 0 {
		t.Fatalf("ожидали 0 ревьюверов, получили %d", len(resp.PR.AssignedReviewers))
	}
}

// Проверяем, что Merge переводит PR в MERGED и повторный вызов не меняет merged_at
func TestPRService_Merge_Idempotent(t *testing.T) {
	svc, prRepo, userRepo := newTestPRService()

	userRepo.Users["u1"] = repository.User{ID: "u1", Username: "Alice", TeamID: 1, IsActive: true}

	// PR в статусе OPEN
	prRepo.PRs["pr-3"] = repository.PullRequest{
		ID:        "pr-3",
		Name:      "Feature",
		AuthorID:  "u1",
		StatusID:  repository.StatusOpen,
		CreatedAt: time.Now().UTC(),
	}

	ctx := context.Background()

	// Первый merge
	resp1, err := svc.Merge(ctx, dto.MergePrRequest{PullRequestID: "pr-3"})
	if err != nil {
		t.Fatalf("первый Merge вернул ошибку: %v", err)
	}
	if resp1.PR.Status != dto.PrStatusMerged {
		t.Fatalf("ожидали статус MERGED после первого merge, получили %s", resp1.PR.Status)
	}
	if resp1.PR.MergedAt == nil {
		t.Fatalf("ожидали, что mergedAt будет установлен после первого merge")
	}
	first := *resp1.PR.MergedAt

	// Второй merge  не должен ничего сломать и менять время
	resp2, err := svc.Merge(ctx, dto.MergePrRequest{PullRequestID: "pr-3"})
	if err != nil {
		t.Fatalf("второй Merge вернул ошибку: %v", err)
	}
	if resp2.PR.Status != dto.PrStatusMerged {
		t.Fatalf("ожидали статус MERGED после второго merge, получили %s", resp2.PR.Status)
	}
	if resp2.PR.MergedAt == nil {
		t.Fatalf("ожидали, что mergedAt будет установлен после второго merge")
	}
	second := *resp2.PR.MergedAt

	if !first.Equal(second) {
		t.Fatalf("ожидали, что mergedAt не изменится, было=%v стало=%v", first, second)
	}
}

// Проверяем, что Reassign заменяет старого ревьювера на нового из той же команды
func TestPRService_Reassign_Success(t *testing.T) {
	svc, prRepo, userRepo := newTestPRService()

	// Команда ревьюверов
	userRepo.Users["u2"] = repository.User{ID: "u2", Username: "Vitya", TeamID: 1, IsActive: true}
	userRepo.Users["u3"] = repository.User{ID: "u3", Username: "Jeka", TeamID: 1, IsActive: true}

	// Автор — другая команда, чтобы не мешал
	userRepo.Users["u1"] = repository.User{ID: "u1", Username: "Alice", TeamID: 2, IsActive: true}

	now := time.Now().UTC()
	prRepo.PRs["pr-4"] = repository.PullRequest{
		ID:        "pr-4",
		Name:      "Reassign me",
		AuthorID:  "u1",
		StatusID:  repository.StatusOpen,
		CreatedAt: now,
	}
	prRepo.Reviewers["pr-4"] = []string{"u2"} // старый ревьювер

	req := dto.ReassignPrRequest{
		PullRequestID: "pr-4",
		OldUserID:     "u2",
	}

	resp, err := svc.Reassign(context.Background(), req)
	if err != nil {
		t.Fatalf("Reassign вернул ошибку: %v", err)
	}

	if resp.ReplacedBy == "" {
		t.Fatalf("ожидали, что поле ReplacedBy будет заполнено")
	}
	if resp.ReplacedBy == "u2" {
		t.Fatalf("ожидали нового ревьювера, отличного от старого, но получили u2")
	}

	_, reviewers, err := prRepo.GetWithReviewers(context.Background(), "pr-4")
	if err != nil {
		t.Fatalf("GetWithReviewers вернул ошибку: %v", err)
	}
	if len(reviewers) != 1 {
		t.Fatalf("ожидали 1 ревьювера после Reassign, получили %d", len(reviewers))
	}
	if reviewers[0] != resp.ReplacedBy {
		t.Fatalf("ожидали ревьювера %s, в репозитории %s", resp.ReplacedBy, reviewers[0])
	}
}

// Проверяем, что для MERGED PR метод Reassign возвращает ErrPRMerged
func TestPRService_Reassign_MergedPR(t *testing.T) {
	svc, prRepo, _ := newTestPRService()

	now := time.Now().UTC()
	prRepo.PRs["pr-5"] = repository.PullRequest{
		ID:        "pr-5",
		Name:      "Already merged",
		AuthorID:  "u1",
		StatusID:  repository.StatusMerged,
		CreatedAt: now,
		MergedAt:  &now,
	}
	prRepo.Reviewers["pr-5"] = []string{"u2"}

	req := dto.ReassignPrRequest{
		PullRequestID: "pr-5",
		OldUserID:     "u2",
	}

	_, err := svc.Reassign(context.Background(), req)
	if err == nil {
		t.Fatalf("ожидали ошибку, но получили nil")
	}
	if !errors.Is(err, repository.ErrPRMerged) {
		t.Fatalf("ожидали ошибку ErrPRMerged, получили %v", err)
	}
}

// Проверяем, что при попытке создать уже существующий PR возвращается ErrPRExists
func TestPRService_Create_PRAlreadyExists(t *testing.T) {
	svc, prRepo, userRepo := newTestPRService()

	// автор существует
	userRepo.Users["u1"] = repository.User{ID: "u1", Username: "Author", TeamID: 1, IsActive: true}

	// PR уже есть в базе
	prRepo.PRs["pr-dup"] = repository.PullRequest{
		ID:        "pr-dup",
		Name:      "Old",
		AuthorID:  "u1",
		StatusID:  repository.StatusOpen,
		CreatedAt: time.Now().UTC(),
	}

	req := dto.CreatePrRequest{
		PullRequestID:   "pr-dup",
		PullRequestName: "New name",
		AuthorID:        "u1",
	}

	_, err := svc.Create(context.Background(), req)
	if err == nil {
		t.Fatalf("ожидали ошибку, получили nil")
	}
	if !errors.Is(err, repository.ErrPRExists) {
		t.Fatalf("ожидали ErrPRExists, получили %v", err)
	}
}

// Проверяем, что если нет ни одного кандидата для замены ревьювера, возвращается ErrNoCandidate
func TestPRService_Reassign_NoCandidates(t *testing.T) {
	svc, prRepo, userRepo := newTestPRService()

	// один пользователь, он же ревьювер и он же единственный в команде
	userRepo.Users["u1"] = repository.User{ID: "u1", Username: "Solo", TeamID: 1, IsActive: true}

	now := time.Now().UTC()
	prRepo.PRs["pr-solo"] = repository.PullRequest{
		ID:        "pr-solo",
		Name:      "Change",
		AuthorID:  "u1",
		StatusID:  repository.StatusOpen,
		CreatedAt: now,
	}
	prRepo.Reviewers["pr-solo"] = []string{"u1"}

	req := dto.ReassignPrRequest{
		PullRequestID: "pr-solo",
		OldUserID:     "u1",
	}

	_, err := svc.Reassign(context.Background(), req)
	if err == nil {
		t.Fatalf("ожидали ошибку, получили nil")
	}
	if !errors.Is(err, repository.ErrNoCandidate) {
		t.Fatalf("ожидали ErrNoCandidate, получили %v", err)
	}
}

// Проверяем, что Merge возвращает ErrPRNotFound, если PR не существует
func TestPRService_Merge_PRNotFound(t *testing.T) {
	svc, _, _ := newTestPRService()

	_, err := svc.Merge(context.Background(), dto.MergePrRequest{
		PullRequestID: "unknown-pr",
	})
	if err == nil {
		t.Fatalf("ожидали ошибку, но получили nil")
	}
	if !errors.Is(err, repository.ErrPRNotFound) {
		t.Fatalf("ожидали ErrPRNotFound, получили %v", err)
	}
}

// Проверяем, что Reassign возвращает ErrReviewerNotSet,
// если переданный old_user_id не назначен ревьювером на этот PR
func TestPRService_Reassign_ReviewerNotAssigned(t *testing.T) {
	svc, prRepo, _ := newTestPRService()

	// PR в статусе OPEN c одним ревьювером u1
	prRepo.PRs["pr-6"] = repository.PullRequest{
		ID:       "pr-6",
		Name:     "Reassign error",
		AuthorID: "author",
		StatusID: repository.StatusOpen,
	}
	prRepo.Reviewers["pr-6"] = []string{"u1"} // единственный ревьювер

	req := dto.ReassignPrRequest{
		PullRequestID: "pr-6",
		OldUserID:     "u999", // такого ревьювера у PR нет
	}

	_, err := svc.Reassign(context.Background(), req)
	if err == nil {
		t.Fatalf("ожидали ошибку, но получили nil")
	}
	if !errors.Is(err, repository.ErrReviewerNotSet) {
		t.Fatalf("ожидали ErrReviewerNotSet, получили %v", err)
	}
}
