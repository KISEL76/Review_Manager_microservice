package httpapi

import (
	"errors"
	"net/http"

	"review-manager/internal/dto"
	"review-manager/internal/repository"
	"review-manager/internal/validation"
)

/* POST /pullRequest/create */

func (h *Handler) PrCreate(w http.ResponseWriter, r *http.Request) {
	var req dto.CreatePrRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	if err := validation.ValidateCreatePR(req); err != nil {
		writeError(w, http.StatusBadRequest, dto.ErrorCodeNotFound, err.Error())
		return
	}

	resp, err := h.PrSvc.Create(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrPRExists):
			writeError(w, http.StatusConflict, dto.ErrorCodePRExists, err.Error())
		case errors.Is(err, repository.ErrUserNotFound),
			errors.Is(err, repository.ErrTeamNotFound):
			writeError(w, http.StatusNotFound, dto.ErrorCodeNotFound, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, dto.ErrorCodeNotFound, internalError)
		}
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

/* POST /pullRequest/merge */

func (h *Handler) MergePR(w http.ResponseWriter, r *http.Request) {
	var req dto.MergePrRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	if err := validation.ValidateMergePR(req); err != nil {
		writeError(w, http.StatusBadRequest, dto.ErrorCodeNotFound, err.Error())
		return
	}

	resp, err := h.PrSvc.Merge(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrPRNotFound):
			writeError(w, http.StatusNotFound, dto.ErrorCodeNotFound, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, dto.ErrorCodeNotFound, internalError)
		}
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

/* POST /pullRequest/reassign */

func (h *Handler) PrReassign(w http.ResponseWriter, r *http.Request) {
	var req dto.ReassignPrRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	if err := validation.ValidateReassignPR(req); err != nil {
		writeError(w, http.StatusBadRequest, dto.ErrorCodeNotFound, err.Error())
		return
	}

	resp, err := h.PrSvc.Reassign(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrPRNotFound),
			errors.Is(err, repository.ErrUserNotFound):
			writeError(w, http.StatusNotFound, dto.ErrorCodeNotFound, err.Error())

		case errors.Is(err, repository.ErrPRMerged):
			writeError(w, http.StatusConflict, dto.ErrorCodePRMerged, err.Error())

		case errors.Is(err, repository.ErrReviewerNotSet):
			writeError(w, http.StatusConflict, dto.ErrorCodeNotAssigned, err.Error())

		case errors.Is(err, repository.ErrNoCandidate):
			writeError(w, http.StatusConflict, dto.ErrorCodeNoCandidate, err.Error())

		default:
			writeError(w, http.StatusInternalServerError, dto.ErrorCodeNotFound, internalError)
		}
		return
	}

	writeJSON(w, http.StatusOK, resp)
}
