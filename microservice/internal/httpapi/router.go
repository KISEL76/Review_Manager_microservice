package httpapi

import (
	"net/http"

	"review-manager/internal/service"
)

type Handler struct {
	TeamSvc    *service.TeamService
	UserSvc    *service.UserService
	PrSvc      *service.PRService
	AdminToken string
}

func NewHandler(team *service.TeamService, user *service.UserService, pr *service.PRService, admToken string) *Handler {
	return &Handler{
		TeamSvc:    team,
		UserSvc:    user,
		PrSvc:      pr,
		AdminToken: admToken,
	}
}

func NewMux(h *Handler) http.Handler {
	mux := http.NewServeMux()

	// health
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Teams
	mux.HandleFunc("/team/add", h.TeamAdd)
	mux.HandleFunc("/team/get", h.TeamGet)

	// Users
	mux.HandleFunc("/users/setIsActive", h.UserSetActive)
	mux.HandleFunc("/users/getReview", h.UserGetReview)

	// Pull Requests
	mux.HandleFunc("/pullRequest/create", h.PrCreate)
	mux.HandleFunc("/pullRequest/merge", h.MergePR)
	mux.HandleFunc("/pullRequest/reassign", h.PrReassign)

	// Stats
	mux.HandleFunc("/stats/reviewers", h.StatsReviewerAssignments)

	return mux
}
