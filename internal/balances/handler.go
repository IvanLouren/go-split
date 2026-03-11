package balances

import (
	"encoding/json"
	"net/http"

	"github.com/IvanLouren/GoSplit/pkg/models"
	"github.com/google/uuid"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetBalances godoc
// @Summary      Get net balances for all users in a group
// @Tags         balances
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Group ID"
// @Success      200  {array}   models.Balance
// @Failure      400  {string}  string  "invalid group ID"
// @Failure      401  {string}  string  "unauthorized"
// @Failure      500  {string}  string  "internal error"
// @Router       /api/groups/{id}/balances [get]
func (h *Handler) GetBalances(w http.ResponseWriter, r *http.Request) {
	groupIDStr := r.PathValue("id")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		http.Error(w, "invalid group ID", http.StatusBadRequest)
		return
	}

	balances, err := h.service.GetBalances(groupID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if balances == nil {
		balances = []models.Balance{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(balances)
}
