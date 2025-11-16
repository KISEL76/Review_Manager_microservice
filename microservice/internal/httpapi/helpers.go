package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"review-manager/internal/dto"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, msg string) {
	resp := dto.ErrorResponse{
		Error: dto.ErrorBody{
			Code:    code,
			Message: msg,
		},
	}
	writeJSON(w, status, resp)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, dto.ErrorCodeNotFound, invalidJSON)
		return false
	}
	return true
}

func (h *Handler) requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	auth := r.Header.Get("Authorization")
	if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
		writeError(w, http.StatusUnauthorized, dto.ErrorCodeNotFound, rscNotFound)
		return false
	}

	token := strings.TrimPrefix(auth, "Bearer ")
	token = strings.TrimSpace(token)

	if token != h.AdminToken {
		writeError(w, http.StatusUnauthorized, dto.ErrorCodeNotFound, rscNotFound)
		return false
	}

	return true
}
