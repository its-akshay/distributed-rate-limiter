package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	RequestsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "rate_limiter_requests_total",
			Help: "Total rate limit check requests",
		},
	)

	AllowedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "rate_limiter_allowed_total",
			Help: "Total allowed requests",
		},
	)

	RejectedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "rate_limiter_rejected_total",
			Help: "Total rejected requests",
		},
	)

	ErrorsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "rate_limiter_errors_total",
			Help: "Total errors",
		},
	)
)

func Init() {
	prometheus.MustRegister(
		RequestsTotal,
		AllowedTotal,
		RejectedTotal,
		ErrorsTotal,
	)
}