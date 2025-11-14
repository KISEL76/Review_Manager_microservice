package validation

import (
	"errors"
	"strings"

	"review-manager/internal/models/dto"
)

/* /pullRequest/create */
func ValidateCreatePR(req dto.CreatePrRequest) error {
	if strings.TrimSpace(req.PullRequestID) == "" {
		return errors.New("pull_request_id is required")
	}
	if strings.TrimSpace(req.PullRequestName) == "" {
		return errors.New("pull_request_name is required")
	}
	if strings.TrimSpace(req.AuthorID) == "" {
		return errors.New("author_id is required")
	}
	return nil
}

/* /pullRequest/merge */
func ValidateMergePR(req dto.MergePrRequest) error {
	if strings.TrimSpace(req.PullRequestID) == "" {
		return errors.New("pull_request_id is required")
	}
	return nil
}

/* /pullRequest/reassign */
func ValidateReassignPR(req dto.ReassignPrRequest) error {
	if strings.TrimSpace(req.PullRequestID) == "" {
		return errors.New("pull_request_id is required")
	}
	if strings.TrimSpace(req.OldUserID) == "" {
		return errors.New("old_user_id is required")
	}
	return nil
}
