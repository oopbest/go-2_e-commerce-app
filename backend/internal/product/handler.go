package product

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

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
	mux.HandleFunc("GET /api/brands", h.handleGetBrands)
	// Protected Endpoints (ต้องมี Token และต้องเป็น Admin เท่านั้น)
	mux.HandleFunc("POST /api/products", auth(middleware.RequireRole(domain.RoleAdmin, h.handleCreateProduct)))
	mux.HandleFunc("PUT /api/products/{id}", auth(middleware.RequireRole(domain.RoleAdmin, h.handleUpdateProduct)))
	mux.HandleFunc("DELETE /api/products/{id}", auth(middleware.RequireRole(domain.RoleAdmin, h.handleDeleteProduct)))
}

// handleGetProducts godoc
// @Summary      Get products with search, filter, sort and pagination
// @Description  Retrieve products list. If no query params are provided, returns all products. Supports search, category, brand, price range, stock, sort and pagination.
// @Tags         Products
// @Produce      json
// @Param        search        query     string  false  "Keyword to search in name or description"
// @Param        category_id   query     int     false  "Filter by Category ID"
// @Param        brand_id      query     int     false  "Filter by Brand ID"
// @Param        min_price     query     number  false  "Minimum price"
// @Param        max_price     query     number  false  "Maximum price"
// @Param        in_stock      query     bool    false  "Filter in-stock products only (true)"
// @Param        sort_by       query     string  false  "Sort by: price_asc, price_desc, rating, newest"
// @Param        page          query     int     false  "Page number (default: 1)"
// @Param        limit         query     int     false  "Items per page (default: 20, max: 100)"
// @Success      200           {object}  domain.ProductListResponse
// @Failure      500           {object}  map[string]string
// @Router       /api/products [get]
func (h *Handler) handleGetProducts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	// 1. ถ้าไม่มี Query Params ใดๆ เลย -> คืนค่าตามเดิม (Backward Compatibility)
	if len(q) == 0 {
		products, err := h.service.GetAllProducts()
		if err != nil {
			h.sendError(w, http.StatusInternalServerError, err.Error())
			return
		}
		h.sendJSON(w, http.StatusOK, products)
		return
	}

	// 2. ถ้ามี Query Params -> แปลงค่าเป็น ProductFilter DTO
	filter := domain.ProductFilter{
		Search: strings.TrimSpace(q.Get("search")),
		SortBy: q.Get("sort_by"),
		Page:   1,
		Limit:  20,
	}

	if catStr := q.Get("category_id"); catStr != "" {
		if catID, err := strconv.Atoi(catStr); err == nil && catID > 0 {
			filter.CategoryID = &catID
		}
	}

	if brandStr := q.Get("brand_id"); brandStr != "" {
		if bID, err := strconv.Atoi(brandStr); err == nil && bID > 0 {
			filter.BrandID = &bID
		}
	}

	if minStr := q.Get("min_price"); minStr != "" {
		if minP, err := strconv.ParseFloat(minStr, 64); err == nil && minP >= 0 {
			filter.MinPrice = &minP
		}
	}

	if maxStr := q.Get("max_price"); maxStr != "" {
		if maxP, err := strconv.ParseFloat(maxStr, 64); err == nil && maxP > 0 {
			filter.MaxPrice = &maxP
		}
	}

	if q.Get("in_stock") == "true" {
		filter.InStockOnly = true
	}

	if pageStr := q.Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			filter.Page = p
		}
	}

	if limitStr := q.Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			filter.Limit = l
		}
	}

	// 3. เรียก Service พร้อม Context
	resp, err := h.service.GetProductsWithFilter(r.Context(), filter)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.sendJSON(w, http.StatusOK, resp)
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

// handleGetBrands godoc
// @Summary      Get all brands
// @Description  Retrieve a list of all brands (Cached in Redis)
// @Tags         Brands
// @Produce      json
// @Success      200  {array}   domain.Brand
// @Failure      500  {object}  map[string]string
// @Router       /api/brands [get]
func (h *Handler) handleGetBrands(w http.ResponseWriter, r *http.Request) {
	brands, err := h.service.GetAllBrands()
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.sendJSON(w, http.StatusOK, brands)
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
