// TODOHP:  review
package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "tee_auction",
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests.",
		},
		[]string{"method", "route", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "tee_auction",
			Name:      "http_request_duration_seconds",
			Help:      "Duration of HTTP requests in seconds.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "route", "status"},
	)

	httpInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "tee_auction",
			Name:      "http_in_flight_requests",
			Help:      "Current number of in-flight HTTP requests.",
		},
	)

	auctionRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "tee_auction",
			Name:      "auction_requests_total",
			Help:      "Total number of auction requests by result.",
		},
		[]string{"result"},
	)

	auctionRunDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "tee_auction",
			Name:      "auction_run_duration_seconds",
			Help:      "Duration of TEE auction execution in seconds.",
			Buckets:   prometheus.DefBuckets,
		},
	)

	reportDeserializeTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "tee_auction",
			Name:      "report_deserialize_total",
			Help:      "Total number of report deserialization attempts by result.",
		},
		[]string{"result"},
	)
)

func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		httpInFlight.Inc()
		defer httpInFlight.Dec()

		c.Next()

		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}
		status := strconv.Itoa(c.Writer.Status())

		httpRequestsTotal.WithLabelValues(c.Request.Method, route, status).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, route, status).Observe(time.Since(start).Seconds())
	}
}

func Handler() gin.HandlerFunc {
	return gin.WrapH(promhttp.Handler())
}

func ObserveAuctionRequest(result string) {
	auctionRequestsTotal.WithLabelValues(result).Inc()
}

func ObserveAuctionRunDuration(d time.Duration) {
	auctionRunDuration.Observe(d.Seconds())
}

func ObserveReportDeserialize(result string) {
	reportDeserializeTotal.WithLabelValues(result).Inc()
}
