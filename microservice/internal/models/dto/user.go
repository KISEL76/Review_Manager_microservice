package dto

type User struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

/* /users/setIsActive */
type SetUserIsActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

type SetUserIsActiveResponse struct {
	User User `json:"user"`
}

/* /users/getReview */
type UserGetReviewResponse struct {
	UserID       string             `json:"user_id"`
	PullRequests []PullRequestShort `json:"pull_requests"`
}
