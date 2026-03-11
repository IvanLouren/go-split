package groups

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

type CreateGroupRequest struct {
	Name string `json:"name"`
}

type AddMemberRequest struct {
	UserID string `json:"user_id"`
}

// CreateGroup godoc
// @Summary      Create a new group
// @Tags         groups
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      CreateGroupRequest  true  "Group name"
// @Success      201   {object}  models.Group
// @Failure      400   {string}  string  "invalid request body"
// @Failure      401   {string}  string  "unauthorized"
// @Failure      500   {string}  string  "internal error"
// @Router       /api/groups [post]
func (h *Handler) CreateGroup(w http.ResponseWriter, r *http.Request) {

	userID := middleware.GetUserID(r)
	parsedID, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "invalid user ID in token", http.StatusUnauthorized)
		return
	}

	var req CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	group, err := h.service.CreateGroup(req.Name, parsedID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(group)
}

// GetGroups godoc
// @Summary      List all groups for the current user
// @Tags         groups
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   models.Group
// @Failure      401  {string}  string  "unauthorized"
// @Failure      500  {string}  string  "internal error"
// @Router       /api/groups [get]
func (h *Handler) GetGroups(w http.ResponseWriter, r *http.Request) {

	userID := middleware.GetUserID(r)
	parsedID, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "invalid user ID in token", http.StatusUnauthorized)
		return
	}

	groups, err := h.service.GetGroups(parsedID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if groups == nil {
		groups = []models.Group{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(groups)
}

// GetGroup godoc
// @Summary      Get a group by ID
// @Tags         groups
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Group ID"
// @Success      200  {object}  models.Group
// @Failure      400  {string}  string  "invalid group ID"
// @Failure      401  {string}  string  "unauthorized"
// @Failure      404  {string}  string  "group not found"
// @Failure      500  {string}  string  "internal error"
// @Router       /api/groups/{id} [get]
func (h *Handler) GetGroup(w http.ResponseWriter, r *http.Request) {

	groupIDStr := r.PathValue("id")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		http.Error(w, "invalid group ID", http.StatusBadRequest)
		return
	}

	group, err := h.service.GetGroup(groupID)

	if err == sql.ErrNoRows {
		http.Error(w, "group not found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(group)
}

// UpdateGroup godoc
// @Summary      Update a group name
// @Tags         groups
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string              true  "Group ID"
// @Param        body  body      CreateGroupRequest  true  "New group name"
// @Success      200   {object}  models.Group
// @Failure      400   {string}  string  "invalid request"
// @Failure      401   {string}  string  "unauthorized"
// @Failure      404   {string}  string  "group not found"
// @Failure      500   {string}  string  "internal error"
// @Router       /api/groups/{id} [put]
func (h *Handler) UpdateGroup(w http.ResponseWriter, r *http.Request) {

	groupIDStr := r.PathValue("id")
	groupID, err := uuid.Parse(groupIDStr)

	if err != nil {
		http.Error(w, "invalid group ID", http.StatusBadRequest)
		return
	}

	var req CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	updatedGroup, err := h.service.UpdateGroup(groupID, req.Name)

	if err == sql.ErrNoRows {
		http.Error(w, "group not found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedGroup)
}

// DeleteGroup godoc
// @Summary      Delete a group
// @Tags         groups
// @Security     BearerAuth
// @Param        id   path      string  true  "Group ID"
// @Success      204
// @Failure      400  {string}  string  "invalid group ID"
// @Failure      401  {string}  string  "unauthorized"
// @Failure      500  {string}  string  "internal error"
// @Router       /api/groups/{id} [delete]
func (h *Handler) DeleteGroup(w http.ResponseWriter, r *http.Request) {

	groupIDStr := r.PathValue("id")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		http.Error(w, "invalid group ID", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteGroup(groupID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// AddMember godoc
// @Summary      Add a member to a group
// @Tags         groups
// @Accept       json
// @Security     BearerAuth
// @Param        id    path      string            true  "Group ID"
// @Param        body  body      AddMemberRequest  true  "User to add"
// @Success      204
// @Failure      400  {string}  string  "invalid request"
// @Failure      401  {string}  string  "unauthorized"
// @Failure      500  {string}  string  "internal error"
// @Router       /api/groups/{id}/members [post]
func (h *Handler) AddMember(w http.ResponseWriter, r *http.Request) {

	groupIDStr := r.PathValue("id")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		http.Error(w, "invalid group ID", http.StatusBadRequest)
		return
	}

	var req AddMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}
	err = h.service.AddMember(groupID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RemoveMember godoc
// @Summary      Remove a member from a group
// @Tags         groups
// @Security     BearerAuth
// @Param        id       path      string  true  "Group ID"
// @Param        user_id  path      string  true  "User ID"
// @Success      204
// @Failure      400  {string}  string  "invalid ID"
// @Failure      401  {string}  string  "unauthorized"
// @Failure      500  {string}  string  "internal error"
// @Router       /api/groups/{id}/members/{user_id} [delete]
func (h *Handler) RemoveMember(w http.ResponseWriter, r *http.Request) {

	groupIDStr := r.PathValue("id")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		http.Error(w, "invalid group ID", http.StatusBadRequest)
		return
	}

	userIDStr := r.PathValue("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}

	err = h.service.RemoveMember(groupID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
