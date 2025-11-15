package unittests

import (
	"context"
	"errors"
	"testing"

	"review-manager/internal/models/dto"
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
