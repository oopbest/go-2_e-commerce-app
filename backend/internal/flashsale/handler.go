package flashsale

import (
	"encoding/json"
	"net/http"

	"github.com/oopbest/ecommerce-app/internal/domain"
)

type Handler struct {
	service domain.FlashSaleService
}

// NewHandler Constructor สำหรับสร้าง Flash Sale Handler
func NewHandler(service domain.FlashSaleService) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes ผูก URL Routes
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/flash-sales/active", h.handleGetActiveFlashSale)
}

// handleGetActiveFlashSale godoc
// @Summary      Get current active flash sale (ดึงแคมเปญ Flash Sale ปัจจุบัน)
// @Description  ดึงข้อมูลแคมเปญ Flash Sale ที่กำลังทำงาน พร้อมรายการสินค้า ราคาพิเศษ และโควตาสต็อก
// @Tags         FlashSales
// @Produce      json
// @Success      200  {object}  domain.FlashSale
// @Failure      500  {object}  map[string]string
// @Router       /api/flash-sales/active [get]
func (h *Handler) handleGetActiveFlashSale(w http.ResponseWriter, r *http.Request) {
	sale, err := h.service.GetCurrentActive(r.Context())
	if err != nil {
		h.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if sale == nil {
		h.writeJSON(w, http.StatusOK, map[string]any{
			"message": "no active flash sale currently",
			"items":   []any{},
		})
		return
	}

	h.writeJSON(w, http.StatusOK, sale)
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
