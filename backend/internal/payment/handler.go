package payment

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/oopbest/ecommerce-app/internal/domain"
	"github.com/oopbest/ecommerce-app/internal/middleware"
	"github.com/oopbest/ecommerce-app/pkg/security"
)

type Handler struct {
	service domain.PaymentService
}

// NewHandler Constructor สำหรับสร้าง Payment Handler
func NewHandler(service domain.PaymentService) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes ผูก URL Routes เข้ากับ Handler Methods
func (h *Handler) RegisterRoutes(mux *http.ServeMux, auth func(http.HandlerFunc) http.HandlerFunc) {
	mux.HandleFunc("POST /api/payments/intent", auth(h.handleCreatePaymentIntent))
	mux.HandleFunc("POST /api/payments/confirm", h.handleConfirmPayment) // จำลอง Webhook / Client confirm
	mux.HandleFunc("GET /api/payments/orders/{id}", auth(h.handleGetPaymentByOrderID))
}

// handleCreatePaymentIntent godoc
// @Summary      Create Payment Intent (ขอรับช่องทางชำระเงิน)
// @Description  สร้าง Payment Intent (เช่น PromptPay QR Code หรือ URL สำหรับบัตรเครดิต) ตาม Order ID
// @Tags         Payments
// @Accept       json
// @Produce      json
// @Param        payload body domain.CreatePaymentIntentInput true "Order ID and Payment Method (credit_card, promptpay, truemoney)"
// @Success      201  {object}  domain.PaymentIntentResponse
// @Failure      400  {object}  map[string]string "Invalid request body or order status"
// @Failure      401  {object}  map[string]string "Unauthorized"
// @Failure      404  {object}  map[string]string "Order not found"
// @Security     BearerAuth
// @Router       /api/payments/intent [post]
func (h *Handler) handleCreatePaymentIntent(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(*security.CustomClaims)
	if !ok || claims == nil {
		h.writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var input domain.CreatePaymentIntentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	intent, err := h.service.CreatePaymentIntent(r.Context(), claims.UserID, input)
	if errors.Is(err, domain.ErrOrderNotFound) {
		h.writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	if errors.Is(err, domain.ErrOrderAlreadyPaid) || errors.Is(err, domain.ErrOrderCancelled) || errors.Is(err, domain.ErrInvalidPaymentMethod) {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err != nil {
		h.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	h.writeJSON(w, http.StatusCreated, intent)
}

// handleConfirmPayment godoc
// @Summary      Confirm Payment / Webhook (ยืนยันการชำระเงิน)
// @Description  จำลองการยืนยันการจ่ายเงินสำเร็จ (เปลี่ยน orders.status เป็น 'paid' และบันทึกใบเสร็จ)
// @Tags         Payments
// @Accept       json
// @Produce      json
// @Param        payload body domain.ConfirmPaymentInput true "Transaction confirmation data from Gateway"
// @Success      200  {object}  domain.Payment
// @Failure      400  {object}  map[string]string "Invalid payload, status, or amount mismatch"
// @Failure      404  {object}  map[string]string "Order not found"
// @Router       /api/payments/confirm [post]
func (h *Handler) handleConfirmPayment(w http.ResponseWriter, r *http.Request) {
	var input domain.ConfirmPaymentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	payment, err := h.service.ConfirmPayment(r.Context(), input)
	if errors.Is(err, domain.ErrOrderNotFound) {
		h.writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	if errors.Is(err, domain.ErrOrderAlreadyPaid) || errors.Is(err, domain.ErrOrderCancelled) || errors.Is(err, domain.ErrInvalidPaymentAmount) || errors.Is(err, domain.ErrPaymentFailed) {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err != nil {
		h.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	h.writeJSON(w, http.StatusOK, payment)
}

// handleGetPaymentByOrderID godoc
// @Summary      Get Payment Receipt (ดูใบเสร็จการชำระเงิน)
// @Description  ดึงข้อมูลประวัติการชำระเงินตาม Order ID
// @Tags         Payments
// @Produce      json
// @Param        id   path      int  true  "Order ID"
// @Success      200  {object}  domain.Payment
// @Failure      401  {object}  map[string]string "Unauthorized"
// @Failure      404  {object}  map[string]string "Payment record or order not found"
// @Security     BearerAuth
// @Router       /api/payments/orders/{id} [get]
func (h *Handler) handleGetPaymentByOrderID(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(*security.CustomClaims)
	if !ok || claims == nil {
		h.writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	idStr := r.PathValue("id")
	orderID, err := strconv.Atoi(idStr)
	if err != nil || orderID <= 0 {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid order ID format"})
		return
	}

	payment, err := h.service.GetPaymentByOrderID(r.Context(), claims.UserID, orderID, claims.Role)
	if errors.Is(err, domain.ErrOrderNotFound) || errors.Is(err, domain.ErrPaymentNotFound) {
		h.writeJSON(w, http.StatusNotFound, map[string]string{"error": "payment receipt not found"})
		return
	}
	if err != nil {
		h.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	h.writeJSON(w, http.StatusOK, payment)
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
