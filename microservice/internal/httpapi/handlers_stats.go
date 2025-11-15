package httpapi

import (
	"net/http"
	"review-manager/internal/models/dto"
)

// GET /stats/reviewers
func (h *Handler) StatsReviewerAssignments(w http.ResponseWriter, r *http.Request) {
	resp, err := h.UserSvc.GetReviewerStats(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, dto.ErrorCodeNotFound, internalError)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}
