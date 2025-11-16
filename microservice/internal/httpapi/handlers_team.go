package httpapi

import (
	"errors"
	"net/http"

	"review-manager/internal/dto"
	"review-manager/internal/repository"
	"review-manager/internal/validation"
)

/* POST /team/add */

func (h *Handler) TeamAdd(w http.ResponseWriter, r *http.Request) {
	var req dto.TeamAddRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	if err := validation.ValidateTeamAdd(req); err != nil {
		writeError(w, http.StatusBadRequest, dto.ErrorCodeNotFound, err.Error())
		return
	}

	resp, err := h.TeamSvc.TeamAdd(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrTeamExists):
			writeError(w, http.StatusBadRequest, dto.ErrorCodeTeamExists, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, dto.ErrorCodeNotFound, internalError)
		}
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

/* GET /team/get?team_name=... */

func (h *Handler) TeamGet(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		writeError(w, http.StatusBadRequest, dto.ErrorCodeNotFound, rscNotFound)
		return
	}

	team, err := h.TeamSvc.TeamGet(r.Context(), teamName)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrTeamNotFound):
			writeError(w, http.StatusNotFound, dto.ErrorCodeNotFound, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, dto.ErrorCodeNotFound, internalError)
		}
		return
	}

	writeJSON(w, http.StatusOK, team)
}
