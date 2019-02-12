package monitor

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"online_judge/talcity/scaffold/criteria/route"
)

var defaultHistogramBuckets = []float64{1, 5, 10, 20, 50, 100, 200, 500, 1000, 2000, 5000, 10000}

var labels = []string{"caller", "api", "code"}

type monitor struct {
	nameSpace string
	subSystem string

	counter *prometheus.CounterVec
	timer   *prometheus.HistogramVec
}

var Monitor *monitor

func init() {
}

// Init 只应该被执行一次
func Init(nameSpace, subSystem string) {
	m := &monitor{
		nameSpace: nameSpace,
		subSystem: subSystem,
	}

	// register api counter
	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: nameSpace,
			Subsystem: subSystem,
			Name:      "count",
			Help:      fmt.Sprintf("api counter for %s system in %s", m.subSystem, m.nameSpace),
		},
		labels,
	)
	prometheus.MustRegister(counter)
	m.counter = counter

	// register api timer
	timer := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: m.nameSpace,
			Subsystem: m.subSystem,
			Name:      "timer",
			Help:      fmt.Sprintf("api timer for %s system in %s", m.subSystem, m.nameSpace),
			Buckets:   defaultHistogramBuckets,
		},
		labels,
	)
	prometheus.MustRegister(timer)
	m.timer = timer

	Monitor = m
}

func (m *monitor) Counter(caller, api, code string) (prometheus.Counter, error) {
	if m.counter == nil {
		return nil, errors.New("no counter registered")
	}

	return m.counter.GetMetricWithLabelValues(caller, api, code)
}

func (m *monitor) Timer(caller, api, code string) (prometheus.Observer, error) {
	if m.timer == nil {
		return nil, errors.New("no timer registered")
	}

	return m.timer.GetMetricWithLabelValues(caller, api, code)
}

// SetVersion 设置版本号
func SetVersion(v Version) {
	initVersion(v)
}

// GetVersionHandler /internal/version
func GetVersionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(versionJsonCache)
}

// HealthCheckHandler  使用Head 方法注册
// /internal/health_check
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func NewModuleRoute() *route.ModuleRoute {
	var (
		routes         []*route.Route
		ignorePrevious bool
		middles        []route.Middleware
	)
	routes = []*route.Route{
		route.NewRoute(
			"/internal/health_check",
			"HEAD",
			HealthCheckHandler,
		),
		route.NewRoute(
			"/internal/version",
			"GET",
			GetVersionHandler,
		),
		route.NewRoute(
			"/internal/metrics",
			"GET",
			promhttp.Handler().ServeHTTP,
		),
	}

	ignorePrevious = true
	return &route.ModuleRoute{
		Routes: routes,
		IgnorePreviousMiddleware: ignorePrevious,
		Middlewares:              middles,
	}
}
