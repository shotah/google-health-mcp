package ghealth

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// CivilDate is YYYY-MM-DD in the caller's local timezone semantics.
type CivilDate struct {
	Year  int
	Month time.Month
	Day   int
}

func ParseCivilDate(s string) (CivilDate, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return CivilDate{}, errors.New("empty civil date")
	}
	t, err := time.ParseInLocation("2006-01-02", s, time.Local)
	if err != nil {
		return CivilDate{}, fmt.Errorf("civil date %q: %w", s, err)
	}
	return CivilDate{Year: t.Year(), Month: t.Month(), Day: t.Day()}, nil
}

func TodayCivil(now time.Time) CivilDate {
	t := now.In(time.Local)
	return CivilDate{Year: t.Year(), Month: t.Month(), Day: t.Day()}
}

func (d CivilDate) String() string {
	return fmt.Sprintf("%04d-%02d-%02d", d.Year, int(d.Month), d.Day)
}

func (d CivilDate) AddDays(n int) CivilDate {
	t := time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, time.Local).AddDate(0, 0, n)
	return CivilDate{Year: t.Year(), Month: t.Month(), Day: t.Day()}
}

func resolveDay(date string, now time.Time) (CivilDate, error) {
	if strings.TrimSpace(date) == "" {
		return TodayCivil(now), nil
	}
	return ParseCivilDate(date)
}

// filterDaySession builds a civil day window on session end time (sleep / exercise).
// Uses civil_* fields so DST transitions do not shift the wake-up day.
func filterDaySession(fieldPrefix, date string, now time.Time) (string, CivilDate, error) {
	day, err := resolveDay(date, now)
	if err != nil {
		return "", CivilDate{}, err
	}
	next := day.AddDays(1)
	// e.g. sleep.interval.civil_end_time
	f := fmt.Sprintf(`%s.interval.civil_end_time >= %q AND %s.interval.civil_end_time < %q`,
		fieldPrefix, day.String(), fieldPrefix, next.String())
	return f, day, nil
}

// filterSessionRange bounds sessions by civil start time (activities list).
func filterSessionRange(fieldPrefix, after, before string, now time.Time) (string, error) {
	var parts []string
	if strings.TrimSpace(after) != "" {
		d, err := ParseCivilDate(after)
		if err != nil {
			return "", err
		}
		parts = append(parts, fmt.Sprintf(`%s.interval.civil_start_time >= %q`, fieldPrefix, d.String()))
	}
	if strings.TrimSpace(before) != "" {
		d, err := ParseCivilDate(before)
		if err != nil {
			return "", err
		}
		parts = append(parts, fmt.Sprintf(`%s.interval.civil_start_time < %q`, fieldPrefix, d.String()))
	}
	if len(parts) == 0 {
		// Default: last 7 civil days through tomorrow (exclusive upper).
		end := TodayCivil(now).AddDays(1)
		start := end.AddDays(-7)
		parts = append(parts,
			fmt.Sprintf(`%s.interval.civil_start_time >= %q`, fieldPrefix, start.String()),
			fmt.Sprintf(`%s.interval.civil_start_time < %q`, fieldPrefix, end.String()),
		)
	}
	return strings.Join(parts, " AND "), nil
}

// filterDailyDate filters daily rollup types (resting HR, HRV) for one civil day.
func filterDailyDate(fieldPrefix, date string, now time.Time) (string, CivilDate, error) {
	day, err := resolveDay(date, now)
	if err != nil {
		return "", CivilDate{}, err
	}
	f := fmt.Sprintf(`%s.date = date(%d, %d, %d)`, fieldPrefix, day.Year, int(day.Month), day.Day)
	return f, day, nil
}

// filterWeightPeriod maps garmin-style period (1m/3m) onto civil sample dates.
func filterWeightPeriod(baseDate, period string, now time.Time) (string, error) {
	end := TodayCivil(now)
	if strings.TrimSpace(baseDate) != "" {
		d, err := ParseCivilDate(baseDate)
		if err != nil {
			return "", err
		}
		end = d
	}
	var days int
	switch strings.ToLower(strings.TrimSpace(period)) {
	case "", "1m", "30d":
		days = 30
	case "3m", "90d":
		days = 90
	case "1w", "7d":
		days = 7
	case "6m":
		days = 180
	case "1y", "12m":
		days = 365
	default:
		// Allow Nd (e.g. 14d)
		if nStr, ok := strings.CutSuffix(period, "d"); ok {
			n, err := strconv.Atoi(nStr)
			if err != nil || n <= 0 {
				return "", fmt.Errorf("unsupported weight period %q", period)
			}
			days = n
		} else {
			return "", fmt.Errorf("unsupported weight period %q (use 1m, 3m, 7d, …)", period)
		}
	}
	start := end.AddDays(-days)
	next := end.AddDays(1)
	return fmt.Sprintf(
		`weight.sample_time.civil_time.date >= date(%d, %d, %d) AND weight.sample_time.civil_time.date < date(%d, %d, %d)`,
		start.Year, int(start.Month), start.Day,
		next.Year, int(next.Month), next.Day,
	), nil
}

func parseDurationSeconds(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	s = strings.TrimSuffix(s, "s")
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}

func parseFloatish(v any) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case float32:
		return float64(x)
	case int:
		return float64(x)
	case int64:
		return float64(x)
	case json.Number:
		f, _ := x.Float64()
		return f
	case string:
		f, _ := strconv.ParseFloat(x, 64)
		return f
	default:
		return 0
	}
}