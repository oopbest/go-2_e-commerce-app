package user

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/oopbest/ecommerce-app/internal/domain"
)

type Handler struct {
	service domain.UserService
}

// NewHandler Constructor สำหรับสร้าง User Handler
func NewHandler(service domain.UserService) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes ลงทะเบียน Auth Routes เข้ากับ ServeMux
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/auth/register", h.handleRegister)
	mux.HandleFunc("POST /api/auth/login", h.handleLogin)
}

// handleRegister godoc
// @Summary      Register a new user
// @Description  Register a new customer or admin user and return JWT token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        input body domain.RegisterInput true "User Registration Data"
// @Success      201  {object} domain.AuthResponse
// @Failure      400  {object} map[string]string
// @Failure      409  {object} map[string]string
// @Router       /api/auth/register [post]
func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var input domain.RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	res, err := h.service.Register(input)
	if err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			h.sendError(w, http.StatusConflict, err.Error()) // 409 Conflict (Email ซ้ำ)
			return
		}
		h.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.sendJSON(w, http.StatusCreated, res)
}

// handleLogin godoc
// @Summary      User Login
// @Description  Authenticate user with email and password to receive JWT token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        input body domain.LoginInput true "User Login Credentials"
// @Success      200  {object} domain.AuthResponse
// @Failure      400  {object} map[string]string
// @Failure      401  {object} map[string]string
// @Router       /api/auth/login [post]
func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var input domain.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	res, err := h.service.Login(input)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			h.sendError(w, http.StatusUnauthorized, "Invalid email or password") // 401 Unauthorized
			return
		}
		h.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.sendJSON(w, http.StatusOK, res)
}

// ==========================================
// Helper Methods
// ==========================================

func (h *Handler) sendJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func (h *Handler) sendError(w http.ResponseWriter, status int, message string) {
	h.sendJSON(w, status, map[string]string{"error": message})
}
