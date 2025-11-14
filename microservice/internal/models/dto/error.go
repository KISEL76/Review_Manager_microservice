package dto

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

const (
	ErrorCodeTeamExists  = "TEAM_EXISTS"
	ErrorCodePRExists    = "PR_EXISTS"
	ErrorCodePRMerged    = "PR_MERGED"
	ErrorCodeNotAssigned = "NOT_ASSIGNED"
	ErrorCodeNoCandidate = "NO_CANDIDATE"
	ErrorCodeNotFound    = "NOT_FOUND"
)
