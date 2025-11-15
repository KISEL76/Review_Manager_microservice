package httpapi

import (
	"errors"
	"net/http"

	"review-manager/internal/models/dto"
	"review-manager/internal/repository"
	"review-manager/internal/validation"
)

/* POST /users/setIsActive */

func (h *Handler) UserSetActive(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}

	var req dto.SetUserIsActiveRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	if err := validation.ValidateUserSetActive(req); err != nil {
		writeError(w, http.StatusBadRequest, dto.ErrorCodeNotFound, err.Error())
		return
	}

	user, err := h.UserSvc.SetIsActive(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrUserNotFound):
			writeError(w, http.StatusNotFound, dto.ErrorCodeNotFound, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, dto.ErrorCodeNotFound, internalError)
		}
		return
	}

	writeJSON(w, http.StatusOK, dto.SetUserIsActiveResponse{User: user})
}

/* GET /users/getReview?user_id=... */

func (h *Handler) UserGetReview(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		writeError(w, http.StatusBadRequest, dto.ErrorCodeNotFound, rscNotFound)
		return
	}

	resp, err := h.UserSvc.GetReviewPRs(r.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrUserNotFound):
			writeError(w, http.StatusNotFound, dto.ErrorCodeNotFound, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, dto.ErrorCodeNotFound, internalError)
		}
		return
	}

	writeJSON(w, http.StatusOK, resp)
}
