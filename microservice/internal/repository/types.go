package repository

import (
	"fmt"
	"time"
)

const (
	uniqueViolationErr = "23505"
	StatusOpen         = 1
	StatusMerged       = 2
)

// Ошибки для работы с СУБД
var (
	ErrTeamExists     = fmt.Errorf("team_name already exists")
	ErrTeamNotFound   = fmt.Errorf("resource not found")
	ErrUserNotFound   = fmt.Errorf("resource not found")
	ErrPRNotFound     = fmt.Errorf("resource not found")
	ErrNoCandidate    = fmt.Errorf("no active candidate in team")
	ErrReviewerNotSet = fmt.Errorf("reviewer is not assigned to this PR")
	ErrPRMerged       = fmt.Errorf("cannot reassign on merged PR")
	ErrPRExists       = fmt.Errorf("PR id already exists")
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

type ReviewerStatRow struct {
	UserID       string
	Username     string
	ReviewsCount int
}
