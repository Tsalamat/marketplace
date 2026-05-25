package metrics

import (
	"fmt"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gofiber/fiber/v2"
)

type routeStats struct {
	requests     atomic.Uint64
	errors       atomic.Uint64
	durationNano atomic.Uint64
}

var (
	startedAt = time.Now()
	inFlight  atomic.Int64
	routes    sync.Map
)

func Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		inFlight.Add(1)
		err := c.Next()
		inFlight.Add(-1)

		route := c.Route().Path
		if route == "" {
			route = c.Path()
		}

		stats := getRouteStats(c.Method(), route)
		stats.requests.Add(1)
		stats.durationNano.Add(uint64(time.Since(start).Nanoseconds()))
		if c.Response().StatusCode() >= 500 || err != nil {
			stats.errors.Add(1)
		}

		return err
	}
}

func Handler(c *fiber.Ctx) error {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	var b strings.Builder
	writeHeader(&b, "student_marketplace_uptime_seconds", "Process uptime in seconds.", "gauge")
	writeSample(&b, "student_marketplace_uptime_seconds", time.Since(startedAt).Seconds(), nil)
	writeHeader(&b, "student_marketplace_inflight_requests", "Requests currently being handled.", "gauge")
	writeSample(&b, "student_marketplace_inflight_requests", float64(inFlight.Load()), nil)
	writeHeader(&b, "student_marketplace_go_goroutines", "Current goroutine count.", "gauge")
	writeSample(&b, "student_marketplace_go_goroutines", float64(runtime.NumGoroutine()), nil)
	writeHeader(&b, "student_marketplace_go_alloc_bytes", "Currently allocated heap bytes.", "gauge")
	writeSample(&b, "student_marketplace_go_alloc_bytes", float64(mem.Alloc), nil)

	var keys []string
	routes.Range(func(key, _ any) bool {
		keys = append(keys, key.(string))
		return true
	})
	sort.Strings(keys)

	writeHeader(&b, "student_marketplace_http_requests_total", "Total HTTP requests.", "counter")
	writeHeader(&b, "student_marketplace_http_errors_total", "Total HTTP requests that returned 5xx or handler errors.", "counter")
	writeHeader(&b, "student_marketplace_http_request_duration_seconds_sum", "Total HTTP request duration in seconds.", "counter")
	writeHeader(&b, "student_marketplace_http_request_duration_seconds_count", "Total HTTP request duration observations.", "counter")
	for _, key := range keys {
		value, _ := routes.Load(key)
		stats := value.(*routeStats)
		labels := parseKey(key)
		writeSample(&b, "student_marketplace_http_requests_total", float64(stats.requests.Load()), labels)
		writeSample(&b, "student_marketplace_http_errors_total", float64(stats.errors.Load()), labels)
		writeSample(&b, "student_marketplace_http_request_duration_seconds_sum", float64(stats.durationNano.Load())/float64(time.Second), labels)
		writeSample(&b, "student_marketplace_http_request_duration_seconds_count", float64(stats.requests.Load()), labels)
	}

	c.Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	return c.SendString(b.String())
}

func getRouteStats(method, route string) *routeStats {
	key := method + "\xff" + route
	value, _ := routes.LoadOrStore(key, &routeStats{})
	return value.(*routeStats)
}

func parseKey(key string) map[string]string {
	parts := strings.SplitN(key, "\xff", 2)
	labels := map[string]string{"method": parts[0], "route": ""}
	if len(parts) == 2 {
		labels["route"] = parts[1]
	}
	return labels
}

func writeHeader(b *strings.Builder, name, help, metricType string) {
	fmt.Fprintf(b, "# HELP %s %s\n# TYPE %s %s\n", name, help, name, metricType)
}

func writeSample(b *strings.Builder, name string, value float64, labels map[string]string) {
	fmt.Fprintf(b, "%s%s %g\n", name, formatLabels(labels), value)
}

func formatLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}

	keys := make([]string, 0, len(labels))
	for key := range labels {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		value := strings.ReplaceAll(labels[key], `\`, `\\`)
		value = strings.ReplaceAll(value, `"`, `\"`)
		value = strings.ReplaceAll(value, "\n", `\n`)
		parts = append(parts, fmt.Sprintf(`%s="%s"`, key, value))
	}
	return "{" + strings.Join(parts, ",") + "}"
}
