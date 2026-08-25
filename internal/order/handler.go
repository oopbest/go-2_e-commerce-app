package order

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/oopbest/ecommerce-app/internal/domain"
	"github.com/oopbest/ecommerce-app/internal/middleware"
)

type Handler struct {
	service domain.OrderService
}

// NewHandler Constructor สำหรับสร้าง Order Handler
func NewHandler(service domain.OrderService) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes ลงทะเบียน Endpoint ของ Order (ทุก Endpoint ต้องผ่าน AuthMiddleware)
func (h *Handler) RegisterRoutes(mux *http.ServeMux, auth func(http.HandlerFunc) http.HandlerFunc) {
	mux.HandleFunc("POST /api/orders/checkout", auth(h.handleCheckout))
	mux.HandleFunc("GET /api/orders", auth(h.handleGetOrders))
	mux.HandleFunc("GET /api/orders/{id}", auth(h.handleGetOrderByID))
}

// handleCheckout godoc
// @Summary      Checkout order
// @Description  Create an order with atomic stock deduction in database transaction
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        input body     domain.CheckoutInput true "Cart Items to Checkout"
// @Success      201   {object} domain.Order
// @Failure      400   {object} map[string]string
// @Failure      401   {object} map[string]string
// @Router       /api/orders/checkout [post]
func (h *Handler) handleCheckout(w http.ResponseWriter, r *http.Request) {
	// 1. ดึง Claims จาก Context ที่ AuthMiddleware ตรวจสอบไว้
	claims, ok := middleware.GetUserClaims(r)
	if !ok {
		h.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// 2. แกะข้อมูล JSON Body
	var input domain.CheckoutInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	// 3. สั่ง Checkout ผ่าน Service
	order, err := h.service.Checkout(r.Context(), claims.UserID, input)
	if err != nil {
		if errors.Is(err, domain.ErrInsufficientStock) || errors.Is(err, domain.ErrEmptyCart) || errors.Is(err, domain.ErrInvalidQuantity) {
			h.sendError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.sendJSON(w, http.StatusCreated, order)
}

// handleGetOrders godoc
// @Summary      Get user orders
// @Description  Get list of orders (Customer: own orders, Admin: all orders)
// @Tags         Orders
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}  domain.Order
// @Failure      401  {object} map[string]string
// @Router       /api/orders [get]
func (h *Handler) handleGetOrders(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaims(r)
	if !ok {
		h.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var orders []domain.Order
	var err error

	// ถ้าเป็น Admin ให้ดึงคำสั่งซื้อทั้งหมดในระบบ ถ้าเป็น Customer ให้ดึงเฉพาะของตนเอง
	if claims.Role == domain.RoleAdmin {
		orders, err = h.service.GetAllOrders(r.Context())
	} else {
		orders, err = h.service.GetUserOrders(r.Context(), claims.UserID)
	}

	if err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if orders == nil {
		orders = []domain.Order{} // ป้องกันไม่ให้ส่ง null กลับไปใน JSON ให้ส่ง [] แทน
	}

	h.sendJSON(w, http.StatusOK, orders)
}

// handleGetOrderByID godoc
// @Summary      Get order by ID
// @Description  Get detailed order information by ID
// @Tags         Orders
// @Produce      json
// @Security     BearerAuth
// @Param        id   path     int  true "Order ID"
// @Success      200  {object} domain.Order
// @Failure      400  {object} map[string]string
// @Failure      401  {object} map[string]string
// @Failure      404  {object} map[string]string
// @Router       /api/orders/{id} [get]
func (h *Handler) handleGetOrderByID(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaims(r)
	if !ok {
		h.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	orderID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid order ID")
		return
	}

	order, err := h.service.GetOrderByID(r.Context(), orderID, claims.UserID, claims.Role)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			h.sendError(w, http.StatusNotFound, "Order not found")
			return
		}
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.sendJSON(w, http.StatusOK, order)
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
