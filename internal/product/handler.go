package product

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/oopbest/ecommerce-app/internal/domain"
	"github.com/oopbest/ecommerce-app/internal/middleware"
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

// RegisterRoutes ลงทะเบียน Endpoint เข้ากับ ServeMux พร้อมป้องกันด้วย Auth Middleware
func (h *Handler) RegisterRoutes(mux *http.ServeMux, auth func(http.HandlerFunc) http.HandlerFunc) {
	// Public Endpoints (ไม่ต้องใช้ Token)
	mux.HandleFunc("GET /api/products", h.handleGetProducts)
	mux.HandleFunc("GET /api/products/{id}", h.handleGetProductByID)
	// Protected Endpoints (ต้องมี Token และต้องเป็น Admin เท่านั้น)
	mux.HandleFunc("POST /api/products", auth(middleware.RequireRole("admin", h.handleCreateProduct)))
	mux.HandleFunc("PUT /api/products/{id}", auth(middleware.RequireRole("admin", h.handleUpdateProduct)))
	mux.HandleFunc("DELETE /api/products/{id}", auth(middleware.RequireRole("admin", h.handleDeleteProduct)))
}

// handleGetProducts godoc
// @Summary      Get all products
// @Description  Retrieve a list of all products (Cached in Redis)
// @Tags         Products
// @Produce      json
// @Success      200  {array}   domain.Product
// @Failure      500  {object}  map[string]string
// @Router       /api/products [get]
func (h *Handler) handleGetProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.service.GetAllProducts()
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.sendJSON(w, http.StatusOK, products)
}

// handleGetProductByID godoc
// @Summary      Get product by ID
// @Description  Retrieve single product details by ID
// @Tags         Products
// @Produce      json
// @Param        id   path      int  true  "Product ID"
// @Success      200  {object}  domain.Product
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/products/{id} [get]
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

// handleCreateProduct godoc
// @Summary      Create a new product
// @Description  Create a new product (Admin Only)
// @Tags         Products
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        input body      domain.CreateProductInput true "Create Product Data"
// @Success      201   {object}  domain.Product
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Router       /api/products [post]
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

// handleUpdateProduct godoc
// @Summary      Update a product
// @Description  Update product details by ID (Admin Only)
// @Tags         Products
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int                       true "Product ID"
// @Param        input body      domain.UpdateProductInput true "Update Product Data"
// @Success      200   {object}  domain.Product
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Router       /api/products/{id} [put]
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

// handleDeleteProduct godoc
// @Summary      Delete a product
// @Description  Delete product by ID (Admin Only)
// @Tags         Products
// @Security     BearerAuth
// @Param        id   path     int  true "Product ID"
// @Success      204  "No Content"
// @Failure      400  {object} map[string]string
// @Failure      401  {object} map[string]string
// @Failure      403  {object} map[string]string
// @Failure      404  {object} map[string]string
// @Router       /api/products/{id} [delete]
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
