package expenses

import (
	"database/sql"
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

type SplitRequest struct {
	UserID string  `json:"user_id"`
	Amount float64 `json:"amount"`
}

type CreateExpenseRequest struct {
	Description string         `json:"description"`
	Amount      float64        `json:"amount"`
	Splits      []SplitRequest `json:"splits"`
}

// CreateExpense godoc
// @Summary      Create an expense in a group
// @Tags         expenses
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string                true  "Group ID"
// @Param        body  body      CreateExpenseRequest  true  "Expense data"
// @Success      201   {object}  models.Expense
// @Failure      400   {string}  string  "invalid request"
// @Failure      401   {string}  string  "unauthorized"
// @Failure      500   {string}  string  "internal error"
// @Router       /api/groups/{id}/expenses [post]
func (h *Handler) CreateExpense(w http.ResponseWriter, r *http.Request) {
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

	var req CreateExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Amount <= 0 {
		http.Error(w, "Amount must not be zero", http.StatusBadRequest)
		return
	}

	if req.Description == "" {
		http.Error(w, "Description must not be empty/null", http.StatusBadRequest)
		return
	}

	var splits []SplitInput
	for _, s := range req.Splits {
		splitUserID, err := uuid.Parse(s.UserID)
		if err != nil {
			http.Error(w, "invalid user ID in splits", http.StatusBadRequest)
			return
		}
		splits = append(splits, SplitInput{UserID: splitUserID, Amount: s.Amount})
	}

	const epsilon = 0.01 // to validate splits add up to total
	var total float64
	for _, s := range splits {
		total += s.Amount
	}
	diff := total - req.Amount
	if diff < -epsilon || diff > epsilon {
		http.Error(w, "splits must add up to total amount", http.StatusBadRequest)
		return
	}

	expense, err := h.service.CreateExpense(groupID, parsedID, req.Description, req.Amount, splits)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(expense)
}

// GetExpenses godoc
// @Summary      List all expenses in a group
// @Tags         expenses
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Group ID"
// @Success      200  {array}   models.Expense
// @Failure      400  {string}  string  "invalid group ID"
// @Failure      401  {string}  string  "unauthorized"
// @Failure      500  {string}  string  "internal error"
// @Router       /api/groups/{id}/expenses [get]
func (h *Handler) GetExpenses(w http.ResponseWriter, r *http.Request) {
	groupIDStr := r.PathValue("id")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		http.Error(w, "invalid group ID", http.StatusBadRequest)
		return
	}

	expenses, err := h.service.GetExpenses(groupID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if expenses == nil {
		expenses = []models.Expense{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(expenses)
}

// GetExpense godoc
// @Summary      Get a single expense
// @Tags         expenses
// @Produce      json
// @Security     BearerAuth
// @Param        id         path      string  true  "Group ID"
// @Param        expenseId  path      string  true  "Expense ID"
// @Success      200  {object}  models.Expense
// @Failure      400  {string}  string  "invalid ID"
// @Failure      401  {string}  string  "unauthorized"
// @Failure      404  {string}  string  "expense not found"
// @Failure      500  {string}  string  "internal error"
// @Router       /api/groups/{id}/expenses/{expenseId} [get]
func (h *Handler) GetExpense(w http.ResponseWriter, r *http.Request) {
	expenseIDStr := r.PathValue("expenseId")
	expenseID, err := uuid.Parse(expenseIDStr)
	if err != nil {
		http.Error(w, "Invalid expense ID", http.StatusBadRequest)
		return
	}

	expense, err := h.service.GetExpense(expenseID)
	if err == sql.ErrNoRows {
		http.Error(w, "expense not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(expense)
}

// DeleteExpense godoc
// @Summary      Delete an expense
// @Tags         expenses
// @Security     BearerAuth
// @Param        id         path      string  true  "Group ID"
// @Param        expenseId  path      string  true  "Expense ID"
// @Success      204
// @Failure      400  {string}  string  "invalid ID"
// @Failure      401  {string}  string  "unauthorized"
// @Failure      500  {string}  string  "internal error"
// @Router       /api/groups/{id}/expenses/{expenseId} [delete]
func (h *Handler) DeleteExpense(w http.ResponseWriter, r *http.Request) {
	expenseIDStr := r.PathValue("expenseId")
	expenseID, err := uuid.Parse(expenseIDStr)
	if err != nil {
		http.Error(w, "Invalid expense ID", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteExpense(expenseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
