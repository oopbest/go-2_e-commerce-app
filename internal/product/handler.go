package product

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/oopbest/ecommerce-app/internal/domain"
)

// Handler จัดการ HTTP Request สำหรับ Product
type Handler struct {
	service domain.ProductService
}

// NewHandler Constructor สำหรับสร้าง Handler
func NewHandler(service domain.ProductService) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes ลงทะเบียน Endpoint เข้ากับ ServeMux
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/products", h.handleGetProducts)
	mux.HandleFunc("GET /api/products/{id}", h.handleGetProductByID)
	mux.HandleFunc("POST /api/products", h.handleCreateProduct)
	mux.HandleFunc("PUT /api/products/{id}", h.handleUpdateProduct)
	mux.HandleFunc("DELETE /api/products/{id}", h.handleDeleteProduct)
}

// GET /api/products
func (h *Handler) handleGetProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.service.GetAllProducts()
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.sendJSON(w, http.StatusOK, products)
}

// GET /api/products/{id}
func (h *Handler) handleGetProductByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	product, err := h.service.GetProductByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrProductNotFound) {
			h.sendError(w, http.StatusNotFound, "Product not found")
			return
		}
		h.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.sendJSON(w, http.StatusOK, product)
}

// POST /api/products
func (h *Handler) handleCreateProduct(w http.ResponseWriter, r *http.Request) {
	var input domain.CreateProductInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	product, err := h.service.CreateProduct(input)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.sendJSON(w, http.StatusCreated, product)
}

// PUT /api/products/{id}
func (h *Handler) handleUpdateProduct(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	var input domain.UpdateProductInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	product, err := h.service.UpdateProduct(id, input)
	if err != nil {
		if errors.Is(err, domain.ErrProductNotFound) {
			h.sendError(w, http.StatusNotFound, "Product not found")
			return
		}
		h.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.sendJSON(w, http.StatusOK, product)
}

// DELETE /api/products/{id}
func (h *Handler) handleDeleteProduct(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	if err := h.service.DeleteProduct(id); err != nil {
		if errors.Is(err, domain.ErrProductNotFound) {
			h.sendError(w, http.StatusNotFound, "Product not found")
			return
		}
		h.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
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
