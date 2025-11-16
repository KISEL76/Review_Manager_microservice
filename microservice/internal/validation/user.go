package validation

import (
	"errors"
	"strings"

	"review-manager/internal/dto"
)

func ValidateUserSetActive(req dto.SetUserIsActiveRequest) error {
	if strings.TrimSpace(req.UserID) == "" {
		return errors.New("user_id is required")
	}
	return nil
}

func ValidateUserID(userID string) error {
	if strings.TrimSpace(userID) == "" {
		return errors.New("user_id is required")
	}
	return nil
}
