package validation

import (
	"errors"
	"strings"

	"review-manager/internal/dto"
)

func ValidateTeamAdd(req dto.TeamAddRequest) error {
	if strings.TrimSpace(req.TeamName) == "" {
		return errors.New("team_name is required")
	}

	if len(req.Members) == 0 {
		return errors.New("team must have at least one member")
	}

	for i, m := range req.Members {
		if strings.TrimSpace(m.UserID) == "" {
			return errors.New("members[" + itoa(i) + "].user_id is required")
		}
		if strings.TrimSpace(m.Username) == "" {
			return errors.New("members[" + itoa(i) + "].username is required")
		}
	}

	return nil
}
