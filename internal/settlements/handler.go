package settlements

import (
	"encoding/json"
	"net/http"

	"github.com/IvanLouren/GoSplit/pkg/middleware"
	"github.com/IvanLouren/GoSplit/pkg/models"
	"github.com/google/uuid"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type CreateSettlementRequest struct {
	PaidTo string  `json:"paid_to"`
	Amount float64 `json:"amount"`
}

// CreateSettlement godoc
// @Summary      Record a settlement between two users
// @Tags         settlements
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string                   true  "Group ID"
// @Param        body  body      CreateSettlementRequest  true  "Settlement data"
// @Success      201   {object}  models.Settlement
// @Failure      400   {string}  string  "invalid request"
// @Failure      401   {string}  string  "unauthorized"
// @Failure      500   {string}  string  "internal error"
// @Router       /api/groups/{id}/settlements [post]
func (h *Handler) CreateSettlement(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	parsedID, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "invalid user ID in token", http.StatusUnauthorized)
		return
	}

	groupIDStr := r.PathValue("id")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		http.Error(w, "invalid group ID", http.StatusBadRequest)
		return
	}

	var req CreateSettlementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Amount <= 0 {
		http.Error(w, "Amount must not be zero", http.StatusBadRequest)
		return
	}

	if req.PaidTo == "" {
		http.Error(w, "Paid To must not be empty(someone has to be assigned to receive the certain amount)", http.StatusBadRequest)
		return
	}

	parsedPaidTo, err := uuid.Parse(req.PaidTo)
	if err != nil {
		http.Error(w, "invalid user ID in Paid To", http.StatusBadRequest)
		return
	}

	settlement, err := h.service.CreateSettlement(groupID, parsedID, parsedPaidTo, req.Amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(settlement)
}

// GetSettlements godoc
// @Summary      List all settlements in a group
// @Tags         settlements
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Group ID"
// @Success      200  {array}   models.Settlement
// @Failure      400  {string}  string  "invalid group ID"
// @Failure      401  {string}  string  "unauthorized"
// @Failure      500  {string}  string  "internal error"
// @Router       /api/groups/{id}/settlements [get]
func (h *Handler) GetSettlements(w http.ResponseWriter, r *http.Request) {
	groupIDStr := r.PathValue("id")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		http.Error(w, "invalid group ID", http.StatusBadRequest)
		return
	}

	settlements, err := h.service.GetSettlements(groupID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if settlements == nil {
		settlements = []models.Settlement{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(settlements)
}
