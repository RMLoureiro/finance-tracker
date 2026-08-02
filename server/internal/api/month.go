package api

import (
	"net/http"
	"time"
)

// parseMonth interpreta "2026-08" como o primeiro dia desse mês (UTC).
func parseMonth(s string) (time.Time, bool) {
	t, err := time.Parse("2006-01", s)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

// monthParam lê o parâmetro ?month= (ou o path value "month"), com o mês
// atual como omissão.
func monthParam(r *http.Request) (time.Time, bool) {
	s := r.PathValue("month")
	if s == "" {
		s = r.URL.Query().Get("month")
	}
	if s == "" {
		now := time.Now()
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC), true
	}
	return parseMonth(s)
}
