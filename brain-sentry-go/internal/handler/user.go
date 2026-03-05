package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/integraltech/brainsentry/internal/dto"
	"github.com/integraltech/brainsentry/internal/repository/postgres"
	"github.com/integraltech/brainsentry/internal/service"
)

// UserHandler handles user management endpoints.
type UserHandler struct {
	userRepo    *postgres.UserRepository
	authService *service.AuthService
	bcryptCost  int
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(userRepo *postgres.UserRepository, authService *service.AuthService, bcryptCost int) *UserHandler {
	return &UserHandler{
		userRepo:    userRepo,
		authService: authService,
		bcryptCost:  bcryptCost,
	}
}

// List handles GET /v1/users
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	users, err := h.userRepo.ListByTenant(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list users")
		return
	}

	var resp []dto.UserResponse
	for _, u := range users {
		resp = append(resp, dto.UserResponse{
			ID:    u.ID,
			Email: u.Email,
			Name:  u.Name,
			Roles: u.Roles,
		})
	}
	writeJSON(w, http.StatusOK, resp)
}

// GetByID handles GET /v1/users/{id}
func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	user, err := h.userRepo.FindByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	writeJSON(w, http.StatusOK, dto.UserResponse{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
		Roles: user.Roles,
	})
}

// Create handles POST /v1/users
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" || req.TenantID == "" {
		writeError(w, http.StatusBadRequest, "email, password, and tenantId are required")
		return
	}

	user, err := h.authService.CreateUser(r.Context(), req, h.bcryptCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	writeJSON(w, http.StatusCreated, dto.UserResponse{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
		Roles: user.Roles,
	})
}
