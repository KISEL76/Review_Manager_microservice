package service

import (
	"context"
	"math/rand"
	"review-manager/internal/models/dto"
	"review-manager/internal/repository"
	"time"
)

type PRService struct {
	PRs   repository.PrRepo
	Users repository.UserRepo
	Tx    repository.TxManager
}

func NewPRService(prs repository.PrRepo, users repository.UserRepo, tx repository.TxManager) *PRService {
	return &PRService{
		PRs:   prs,
		Users: users,
		Tx:    tx,
	}
}

func PrToDTO(pr repository.PullRequest, reviewers []string) dto.PullRequest {
	var status dto.PullRequestStatus
	switch pr.StatusID {
	case repository.StatusOpen:
		status = dto.PrStatusOpen
	case repository.StatusMerged:
		status = dto.PrStatusMerged
	default:
		status = dto.PullRequestStatus("UNKNOWN")
	}

	return dto.PullRequest{
		PullRequestID:     pr.ID,
		PullRequestName:   pr.Name,
		AuthorID:          pr.AuthorID,
		Status:            status,
		AssignedReviewers: reviewers,
		CreatedAt:         &pr.CreatedAt,
		MergedAt:          pr.MergedAt,
	}
}

// логика /pullRequest/create
func (s *PRService) Create(ctx context.Context, req dto.CreatePrRequest) (dto.CreatePrResponse, error) {
	exists, err := s.PRs.Exists(ctx, req.PullRequestID)
	if err != nil {
		return dto.CreatePrResponse{}, err
	}
	if exists {
		return dto.CreatePrResponse{}, repository.ErrTeamExists
	}

	author, err := s.Users.GetByID(ctx, req.AuthorID)
	if err != nil {
		return dto.CreatePrResponse{}, err
	}

	candidates, err := s.Users.FindActiveInTeamExcept(ctx, author.TeamID, author.ID, 2)
	if err != nil {
		return dto.CreatePrResponse{}, err
	}
	reviewerIDs := make([]string, 0, len(candidates))
	for _, c := range candidates {
		reviewerIDs = append(reviewerIDs, c.ID)
	}

	now := time.Now().UTC()

	prRow := repository.PullRequest{
		ID:        req.PullRequestID,
		Name:      req.PullRequestName,
		AuthorID:  req.AuthorID,
		StatusID:  repository.StatusOpen,
		CreatedAt: now,
		MergedAt:  nil,
	}

	if err := s.Tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := s.PRs.Insert(ctx, prRow); err != nil {
			return err
		}
		if err := s.PRs.AddReviewers(ctx, prRow.ID, reviewerIDs); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return dto.CreatePrResponse{}, err
	}

	prRow, reviewers, err := s.PRs.GetWithReviewers(ctx, prRow.ID)
	if err != nil {
		return dto.CreatePrResponse{}, err
	}

	return dto.CreatePrResponse{
		PR: PrToDTO(prRow, reviewers),
	}, nil
}

// логика pullRequest/merge
func (s *PRService) Merge(ctx context.Context, req dto.MergePrRequest) (dto.MergePullRequestResponse, error) {
	now := time.Now().UTC()

	prRow, err := s.PRs.SetMerged(ctx, req.PullRequestID, now)
	if err != nil {
		return dto.MergePullRequestResponse{}, err
	}

	_, reviewers, err := s.PRs.GetWithReviewers(ctx, req.PullRequestID)
	if err != nil {
		return dto.MergePullRequestResponse{}, err
	}

	return dto.MergePullRequestResponse{
		PR: PrToDTO(prRow, reviewers),
	}, nil
}

// логика pullRequest/reassign
func (s *PRService) Reassign(ctx context.Context, req dto.ReassignPrRequest) (dto.ReassignPrResponse, error) {
	prRow, reviewers, err := s.PRs.GetWithReviewers(ctx, req.PullRequestID)
	if err != nil {
		return dto.ReassignPrResponse{}, err
	}

	if prRow.StatusID == repository.StatusMerged {
		return dto.ReassignPrResponse{}, repository.ErrPRMerged
	}

	found := false
	for _, r := range reviewers {
		if r == req.OldUserID {
			found = true
			break
		}
	}
	if !found {
		return dto.ReassignPrResponse{}, repository.ErrReviewerNotSet
	}

	oldUser, err := s.Users.GetByID(ctx, req.OldUserID)
	if err != nil {
		return dto.ReassignPrResponse{}, err
	}

	candidates, err := s.Users.FindActiveInTeamExcept(ctx, oldUser.TeamID, oldUser.ID, 20)
	if err != nil {
		return dto.ReassignPrResponse{}, err
	}

	exclude := make(map[string]struct{}, len(reviewers))
	for _, id := range reviewers {
		exclude[id] = struct{}{}
	}

	var possible []string
	for _, c := range candidates {
		if _, used := exclude[c.ID]; used {
			continue
		}
		possible = append(possible, c.ID)
	}

	if len(possible) == 0 {
		return dto.ReassignPrResponse{}, repository.ErrNoCandidate
	}

	newID := possible[rand.Intn(len(possible))]

	if err := s.PRs.ReplaceReviewer(ctx, req.PullRequestID, req.OldUserID, newID); err != nil {
		return dto.ReassignPrResponse{}, err
	}

	prRow, reviewers, err = s.PRs.GetWithReviewers(ctx, req.PullRequestID)
	if err != nil {
		return dto.ReassignPrResponse{}, err
	}

	return dto.ReassignPrResponse{
		PR:         PrToDTO(prRow, reviewers),
		ReplacedBy: newID,
	}, nil
}
