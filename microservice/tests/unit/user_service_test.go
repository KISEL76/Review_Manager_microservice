package unit

import (
	"context"
	"errors"
	"testing"

	"review-manager/internal/dto"
	"review-manager/internal/repository"
	"review-manager/internal/service"
)

// newTestUserService собирает UserService с моковым репозиторием
func newTestUserService() (*service.UserService, *MockUserRepo) {
	userRepo := NewMockUserRepo()
	svc := service.NewUserService(userRepo)
	return svc, userRepo
}

// Проверяем, что SetIsActive меняет флаг активности пользователя и возвращает корректный dto
func TestUserService_SetIsActive_UpdatesFlag(t *testing.T) {
	svc, userRepo := newTestUserService()

	// Исходный пользователь
	userRepo.Users["u1"] = repository.User{
		ID:       "u1",
		Username: "Petya",
		TeamID:   1,
		TeamName: "backend",
		IsActive: true,
	}

	req := dto.SetUserIsActiveRequest{
		UserID:   "u1",
		IsActive: false,
	}

	resp, err := svc.SetIsActive(context.Background(), req)
	if err != nil {
		t.Fatalf("SetIsActive вернул ошибку: %v", err)
	}

	if resp.UserID != "u1" {
		t.Fatalf("ожидали user_id=u1 в ответе, получили %s", resp.UserID)
	}
	if resp.IsActive != false {
		t.Fatalf("ожидали IsActive=false в ответе, получили %v", resp.IsActive)
	}

	u := userRepo.Users["u1"]
	if u.IsActive != false {
		t.Fatalf("ожидали что в репозитории IsActive=false, получили %v", u.IsActive)
	}
}

// Проверяем, что при попытке изменить флаг активности неизвестного пользователя возвращается ErrUserNotFound
func TestUserService_SetIsActive_UserNotFound(t *testing.T) {
	svc, _ := newTestUserService()

	req := dto.SetUserIsActiveRequest{
		UserID:   "unknown",
		IsActive: true,
	}

	_, err := svc.SetIsActive(context.Background(), req)
	if err == nil {
		t.Fatalf("ожидали ошибку, но получили nil")
	}
	if !errors.Is(err, repository.ErrUserNotFound) {
		t.Fatalf("ожидали ErrUserNotFound, получили %v", err)
	}
}

// Проверяем, что GetReviewPRs мапит статусы и поля в DTO
func TestUserService_GetReviewPRs_MapsStatuses(t *testing.T) {
	svc, repo := newTestUserService()

	repo.ReviewPRs["u1"] = []repository.PullRequestShort{
		{ID: "pr1", Name: "Otkrytyj PR", AuthorID: "a1", StatusName: "OPEN"},
		{ID: "pr2", Name: "Smerzhennyj PR", AuthorID: "a2", StatusName: "MERGED"},
	}

	resp, err := svc.GetReviewPRs(context.Background(), "u1")
	if err != nil {
		t.Fatalf("GetReviewPRs вернул ошибку: %v", err)
	}

	if resp.UserID != "u1" {
		t.Fatalf("ожидали user_id = u1 в ответе, получили %s", resp.UserID)
	}
	if len(resp.PullRequests) != 2 {
		t.Fatalf("ожидали 2 PR в ответе, получили %d", len(resp.PullRequests))
	}

	pr1 := resp.PullRequests[0]
	if pr1.PullRequestID != "pr1" || pr1.Status != dto.PrStatusOpen {
		t.Fatalf("первый PR преобразован некорректно, получили %+v", pr1)
	}

	pr2 := resp.PullRequests[1]
	if pr2.PullRequestID != "pr2" || pr2.Status != dto.PrStatusMerged {
		t.Fatalf("второй PR преобразован некорректно, получили %+v", pr2)
	}
}

// Проверяем, что GetReviewerStats собирает DTO из строк репозитория
func TestUserService_GetReviewerStats_BuildsDTO(t *testing.T) {
	svc, repo := newTestUserService()

	repo.StatsRows = []repository.ReviewerStatRow{
		{UserID: "u1", Username: "Ivan", ReviewsCount: 3},
		{UserID: "u2", Username: "Sonya", ReviewsCount: 1},
	}

	resp, err := svc.GetReviewerStats(context.Background())
	if err != nil {
		t.Fatalf("GetReviewerStats вернул ошибку: %v", err)
	}

	if len(resp.Stats) != 2 {
		t.Fatalf("ожидали 2 записи в статистике, получили %d", len(resp.Stats))
	}

	if resp.Stats[0].Username != "Ivan" || resp.Stats[0].ReviewsCount != 3 {
		t.Fatalf("первая запись статистики собрана некорректно, получили %+v", resp.Stats[0])
	}

	if resp.Stats[1].Username != "Sonya" || resp.Stats[1].ReviewsCount != 1 {
		t.Fatalf("вторая запись статистики собрана некорректно, получили %+v", resp.Stats[1])
	}
}
