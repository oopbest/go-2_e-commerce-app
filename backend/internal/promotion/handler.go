package promotion

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/oopbest/ecommerce-app/internal/domain"
)

type Handler struct {
	service domain.PromotionService
}

// NewHandler Constructor สำหรับสร้าง Promotion Handler
func NewHandler(service domain.PromotionService) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes ผูก URL Routes เข้ากับ Handler Methods
func (h *Handler) RegisterRoutes(mux *http.ServeMux, auth func(http.HandlerFunc) http.HandlerFunc) {
	mux.HandleFunc("GET /api/promotions", h.handleGetActivePromotions)
	mux.HandleFunc("POST /api/promotions/validate", h.handleValidateCoupon)
}

// ValidateCouponRequest DTO สำหรับรับข้อมูลตรวจสอบคูปอง
type ValidateCouponRequest struct {
	Code     string  `json:"code"`
	Subtotal float64 `json:"subtotal"`
}

// handleGetActivePromotions godoc
// @Summary      Get active promotions (คูปองส่วนลดที่ใช้งานได้)
// @Description  ดึงรายการคูปองส่วนลดทั้งหมดที่ยังเปิดใช้งานและยังไม่หมดอายุ
// @Tags         Promotions
// @Produce      json
// @Success      200  {array}   domain.Promotion
// @Failure      500  {object}  map[string]string
// @Router       /api/promotions [get]
func (h *Handler) handleGetActivePromotions(w http.ResponseWriter, r *http.Request) {
	promos, err := h.service.GetActivePromotions(r.Context())
	if err != nil {
		h.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	h.writeJSON(w, http.StatusOK, promos)
}

// handleValidateCoupon godoc
// @Summary      Validate coupon code (ตรวจสอบและคำนวณส่วนลด)
// @Description  ตรวจสอบรหัสคูปองกับยอดรวมคำสั่งซื้อ พร้อมคำนวณยอดส่วนลดสุทธิ
// @Tags         Promotions
// @Accept       json
// @Produce      json
// @Param        payload body ValidateCouponRequest true "Coupon Code and Subtotal"
// @Success      200  {object}  domain.ValidationResult
// @Failure      400  {object}  map[string]string "Invalid coupon or conditions not met"
// @Failure      404  {object}  map[string]string "Coupon not found"
// @Router       /api/promotions/validate [post]
func (h *Handler) handleValidateCoupon(w http.ResponseWriter, r *http.Request) {
	var req ValidateCouponRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	result, err := h.service.ValidateCoupon(r.Context(), req.Code, req.Subtotal)
	if err != nil {
		if errors.Is(err, domain.ErrPromotionNotFound) {
			h.writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	h.writeJSON(w, http.StatusOK, result)
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
