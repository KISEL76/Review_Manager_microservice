package service

import (
	"context"

	"review-manager/internal/dto"
	"review-manager/internal/repository"
)

type UserService struct {
	Users repository.UserRepo
}

func NewUserService(users repository.UserRepo) *UserService {
	return &UserService{Users: users}
}

// логика для /users/setIsActive.
func (s *UserService) SetIsActive(ctx context.Context, req dto.SetUserIsActiveRequest) (dto.User, error) {
	u, err := s.Users.SetIsActive(ctx, req.UserID, req.IsActive)
	if err != nil {
		return dto.User{}, err
	}

	return dto.User{
		UserID:   u.ID,
		Username: u.Username,
		TeamName: u.TeamName,
		IsActive: u.IsActive,
	}, nil
}

// логика для /users/getReview.
func (s *UserService) GetReviewPRs(ctx context.Context, userID string) (dto.UserGetReviewResponse, error) {
	prs, err := s.Users.GetReviewPrs(ctx, userID)
	if err != nil {
		return dto.UserGetReviewResponse{}, err
	}

	resp := dto.UserGetReviewResponse{
		UserID:       userID,
		PullRequests: make([]dto.PullRequestShort, 0, len(prs)),
	}

	for _, p := range prs {
		var status dto.PullRequestStatus
		switch p.StatusName {
		case "OPEN":
			status = dto.PrStatusOpen
		case "MERGED":
			status = dto.PrStatusMerged
		default:
			status = dto.PullRequestStatus(p.StatusName)
		}

		resp.PullRequests = append(resp.PullRequests, dto.PullRequestShort{
			PullRequestID:   p.ID,
			PullRequestName: p.Name,
			AuthorID:        p.AuthorID,
			Status:          status,
		})
	}

	return resp, nil
}

// логика для /stats/reviewers
func (s *UserService) GetReviewerStats(ctx context.Context) (dto.ReviewerStatsResponse, error) {
	rows, err := s.Users.GetReviewerStats(ctx)
	if err != nil {
		return dto.ReviewerStatsResponse{}, err
	}

	stats := make([]dto.ReviewerStat, 0, len(rows))
	for _, r := range rows {
		stats = append(stats, dto.ReviewerStat{
			Username:     r.Username,
			ReviewsCount: r.ReviewsCount,
		})
	}

	return dto.ReviewerStatsResponse{Stats: stats}, nil
}
