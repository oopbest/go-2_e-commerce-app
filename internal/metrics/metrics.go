package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// 1. ตัวนับจำนวน HTTP Requests ทั้งหมด แยกตาม Method, Path, และ Status Code
	HttpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ecommerce_http_requests_total",
			Help: "Total number of HTTP requests processed by the API",
		},
		[]string{"method", "path", "status"},
	)

	// 2. Histogram วัดค่า Response Latency (วินาที) สำหรับคำนวณ P90, P95, P99
	HttpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ecommerce_http_request_duration_seconds",
			Help:    "Histogram of HTTP request latencies in seconds",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"method", "path", "status"},
	)

	// 3. Gauge วัดจำนวน Request ที่กำลังประมวลผลอยู่ ณ ขณะนั้น (In-Flight)
	HttpRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "ecommerce_http_requests_in_flight",
			Help: "Current number of HTTP requests being served simultaneously",
		},
	)

	// 4. Business Metrics: ตัวนับจำนวน Order ที่เกิดขึ้นจริง
	OrdersTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ecommerce_orders_total",
			Help: "Total number of checkout orders placed",
		},
		[]string{"status"},
	)

	// 5. Business Metrics: ตัวนับยอดขายรวม (บาท)
	RevenueTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "ecommerce_revenue_thb_total",
			Help: "Total revenue generated in THB",
		},
	)
)
