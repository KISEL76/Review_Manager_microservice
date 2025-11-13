package repository

import (
	"fmt"
	"time"
)

const (
	UNIQUE_VIOLATION = "23505"
)

var (
	ErrTeamExists     = fmt.Errorf("team already exists")
	ErrTeamNotFound   = fmt.Errorf("team not found")
	ErrUserNotFound   = fmt.Errorf("user not found")
	ErrPRNotFound     = fmt.Errorf("pull request not found")
	ErrNoCandidate    = fmt.Errorf("no active candidate in team")
	ErrReviewerNotSet = fmt.Errorf("reviewer is not assigned to this PR")
	ErrPRMerged       = fmt.Errorf("pull request already merged")
)

type Team struct {
	ID   int
	Name string
}

type User struct {
	ID       string
	Username string
	TeamID   int
	TeamName string
	IsActive bool
}

type TeamWithMembers struct {
	Team    Team
	Members []User
}

type PullRequest struct {
	ID         string
	Name       string
	AuthorID   string
	StatusID   int
	StatusName string
	CreatedAt  time.Time
	MergedAt   *time.Time
}

type PullRequestShort struct {
	ID         string
	Name       string
	AuthorID   string
	StatusName string
}
