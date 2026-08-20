package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// ==========================================
// 1. Data Model (เปรียบเหมือน TypeScript Interface + Class)
// ==========================================

// Product โครงสร้างข้อมูลสินค้า
// - ชื่อ Field ขึ้นต้นด้วย "ตัวพิมพ์ใหญ่" = Public (Exported) แพ็กเกจอื่นและ JSON encoder มองเห็น
// - `json:"..."` คือ Struct Tag (เปรียบเสมือนการ Mapping Key ใน JSON)
type Product struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	CreatedAt   time.Time `json:"created_at"`
}

// CreateProductRequest DTO สำหรับรับข้อมูลตอนสร้างสินค้า (เหมือน Req Body Type ใน TS)
type CreateProductRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
}

// ==========================================
// 2. In-Memory Store & Concurrency Safety
// ==========================================

// ใน Node.js Single-Threaded เราประกาศ `let products = []` แล้ว push ได้เลย
// แต่ใน Go ทุกๆ HTTP Request จะทำงานใน "Goroutine" (Thread ย่อย) แยกกันโดยอัตโนมัติ
// ดังนั้นเมื่อเขียนข้อมูลพร้อมกัน ต้องใช้ Mutex ป้องกัน Race Condition
var (
	products = []Product{
		{ID: 1, Name: "Mechanical Keyboard", Description: "RGB Hot-swappable", Price: 2590.00, Stock: 15, CreatedAt: time.Now()},
		{ID: 2, Name: "Wireless Mouse", Description: "Ergonomic 2.4GHz", Price: 1290.00, Stock: 30, CreatedAt: time.Now()},
	}
	nextID = 3
	mu     sync.RWMutex // Mutex สำหรับจัดการ Concurrency (อ่านหลายคนพร้อมกันได้ แต่เขียนทีละคน)
)

// ==========================================
// 3. Helper Functions (เหมือน res.json() และ res.status().json() ใน Express)
// ==========================================

func sendJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON: %v", err)
	}
}

func sendError(w http.ResponseWriter, status int, message string) {
	sendJSON(w, status, map[string]string{"error": message})
}

// ==========================================
// 4. HTTP Handlers (เหมือน Controller ใน Express/Laravel)
// ==========================================

// GET /api/products - ดึงรายการสินค้าทั้งหมด
func getProductsHandler(w http.ResponseWriter, r *http.Request) {
	mu.RLock()         // RLock: ล็อกสำหรับการอ่าน (คนอื่นอ่านพร้อมกันได้ แต่ห้ามเขียน)
	defer mu.RUnlock() // defer: จะทำงานเมื่อฟังก์ชันนี้จบการทำงาน (คล้าย finally ใน JS/PHP)

	sendJSON(w, http.StatusOK, products)
}

// GET /api/products/{id} - ดึงสินค้าตาม ID
func getProductByIDHandler(w http.ResponseWriter, r *http.Request) {
	// ใน Go 1.22+ สามารถดึง Path Parameter ได้ด้วย r.PathValue("id")
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr) // strconv.Atoi = แปลง string เป็น int (เหมือน parseInt ใน JS)
	if err != nil {
		sendError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	mu.RLock()
	defer mu.RUnlock()

	for _, p := range products {
		if p.ID == id {
			sendJSON(w, http.StatusOK, p)
			return
		}
	}

	sendError(w, http.StatusNotFound, "Product not found")
}

// POST /api/products - สร้างสินค้าใหม่
func createProductHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateProductRequest

	// แกะ JSON body (เหมือน req.body ใน Express แต่ใน Go ต้อง Decode ลง Struct ด้วย Pointer &req)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	// Validation พื้นฐาน
	if req.Name == "" {
		sendError(w, http.StatusBadRequest, "Product name is required")
		return
	}
	if req.Price <= 0 {
		sendError(w, http.StatusBadRequest, "Price must be greater than 0")
		return
	}
	if req.Stock < 0 {
		sendError(w, http.StatusBadRequest, "Stock cannot be negative")
		return
	}

	mu.Lock() // Lock: ล็อกสำหรับการเขียน (คนอื่นต้องรอจนกว่าจะเขียนเสร็จ)
	defer mu.Unlock()

	newProduct := Product{
		ID:          nextID,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		CreatedAt:   time.Now(),
	}
	nextID++
	products = append(products, newProduct) // append = เสมือน array.push() ใน JS หรือ array_push() ใน PHP

	sendJSON(w, http.StatusCreated, newProduct)
}

// PUT /api/products/{id} - อัปเดตสินค้า
func updateProductHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")     // r.PathValue() เป็น method ของ Go 1.22+ ใช้สำหรับดึง Path Parameter
	id, err := strconv.Atoi(idStr) // strconv.Atoi() เป็น function ของ package strconv ใช้สำหรับแปลง string เป็น int
	if err != nil {
		sendError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	var req CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.Name == "" {
		sendError(w, http.StatusBadRequest, "Product name is required")
		return
	}
	if req.Price <= 0 {
		sendError(w, http.StatusBadRequest, "Price must be greater than 0")
		return
	}
	if req.Stock < 0 {
		sendError(w, http.StatusBadRequest, "Stock cannot be negative")
		return
	}

	mu.Lock()
	defer mu.Unlock()

	for i, p := range products {
		if p.ID == id {
			products[i].Name = req.Name
			products[i].Description = req.Description
			products[i].Price = req.Price
			products[i].Stock = req.Stock
			sendJSON(w, http.StatusOK, products[i])
			return
		}
	}

	sendError(w, http.StatusNotFound, "Product not found")
}

// DELETE /api/products/{id} - ลบสินค้า
func deleteProductHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		sendError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	mu.Lock()
	defer mu.Unlock()

	for i, p := range products {
		if p.ID == id {
			// 1. เลื่อน Element ถัดไปทั้งหมดมาทับตำแหน่ง i
			copy(products[i:], products[i+1:])

			// 2. เคลียร์ช่องสุดท้ายให้เป็น nil เพื่อให้ GC กวาด Memory ได้
			products[len(products)-1] = Product{}

			// 3. หดความยาว Slice ลง 1 ตัว
			products = products[:len(products)-1]
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}

	sendError(w, http.StatusNotFound, "Product not found")
}

// ==========================================
// 5. Main Application Entry Point
// ==========================================

func main() {
	// สร้าง ServeMux (Router ของ Go มาตรฐาน)
	mux := http.NewServeMux()

	// ลงทะเบียน Routes (Go 1.22+ รองรับระบุ Method นำหน้า Path ได้โดยตรง)
	mux.HandleFunc("GET /api/products", getProductsHandler)
	mux.HandleFunc("GET /api/products/{id}", getProductByIDHandler)
	mux.HandleFunc("POST /api/products", createProductHandler)
	mux.HandleFunc("PUT /api/products/{id}", updateProductHandler)
	mux.HandleFunc("DELETE /api/products/{id}", deleteProductHandler)

	// Route ตรวจสอบสถานะเซิร์ฟเวอร์
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		sendJSON(w, http.StatusOK, map[string]string{
			"status":  "healthy",
			"message": "E-Commerce API is running",
		})
	})

	port := ":8080"
	fmt.Printf("🚀 E-Commerce API server started on http://localhost%s\n", port)

	// สั่งรันเซิร์ฟเวอร์ (Blocking Call คล้าย app.listen ใน Express)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
