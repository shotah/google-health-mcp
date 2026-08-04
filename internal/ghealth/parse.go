package ghealth

import (
	"path"
	"strings"
	"time"
)

func mapSleep(m map[string]any, civilDay CivilDate) *SleepResult {
	sleep := asMap(m["sleep"])
	if sleep == nil {
		return &SleepResult{Date: civilDay.String(), Note: "no sleep payload", Raw: m}
	}
	interval := asMap(sleep["interval"])
	summary := asMap(sleep["summary"])
	start := strField(interval, "startTime")
	end := strField(interval, "endTime")
	minsAsleep := parseFloatish(summary["minutesAsleep"])
	minsInBed := parseFloatish(summary["minutesInSleepPeriod"])
	efficiency := 0.0
	if minsInBed > 0 && minsAsleep > 0 {
		efficiency = (minsAsleep / minsInBed) * 100
	}
	date := civilDay.String()
	if end != "" {
		if t, err := time.Parse(time.RFC3339, end); err == nil {
			// Wake-up civil day in local TZ when caller omitted a date.
			date = t.In(time.Local).Format("2006-01-02")
		}
	}
	return &SleepResult{
		Date:            date,
		DurationMinutes: minsAsleep,
		Efficiency:      efficiency,
		StartTime:       start,
		EndTime:         end,
		DataPointName:   dataPointName(m),
		Raw:             m,
	}
}

func pickMainSleep(points []map[string]any, day CivilDate) *SleepResult {
	if len(points) == 0 {
		return &SleepResult{Date: day.String(), Note: "no sleep data points for day"}
	}
	var best map[string]any
	var bestMins float64
	for _, p := range points {
		sleep := asMap(p["sleep"])
		meta := asMap(sleep["metadata"])
		summary := asMap(sleep["summary"])
		mins := parseFloatish(summary["minutesAsleep"])
		if meta != nil {
			if main, ok := meta["main"].(bool); ok && main {
				return mapSleep(p, day)
			}
		}
		if best == nil || mins > bestMins {
			best = p
			bestMins = mins
		}
	}
	return mapSleep(best, day)
}

func mapExercise(m map[string]any) Activity {
	ex := asMap(m["exercise"])
	interval := asMap(ex["interval"])
	metrics := asMap(ex["metricsSummary"])
	name := dataPointName(m)
	id := name
	if id != "" {
		id = path.Base(id)
	}
	activeDur := parseDurationSeconds(strField(ex, "activeDuration"))
	durationMin := activeDur / 60
	if durationMin == 0 {
		start := strField(interval, "startTime")
		end := strField(interval, "endTime")
		if st, err1 := time.Parse(time.RFC3339, start); err1 == nil {
			if et, err2 := time.Parse(time.RFC3339, end); err2 == nil {
				durationMin = et.Sub(st).Minutes()
			}
		}
	}
	distMM := parseFloatish(metrics["distanceMillimeters"])
	return Activity{
		ID:           firstNonEmpty(name, id),
		Name:         strField(ex, "displayName"),
		ActivityType: strField(ex, "exerciseType"),
		StartTime:    strField(interval, "startTime"),
		DurationMin:  durationMin,
		Calories:     parseFloatish(metrics["caloriesKcal"]),
		Steps:        parseFloatish(metrics["steps"]),
		DistanceKM:   distMM / 1_000_000,
	}
}

func mapRestingHR(m map[string]any, day CivilDate) *HeartRateResult {
	dr := asMap(m["dailyRestingHeartRate"])
	bpm := parseFloatish(dr["beatsPerMinute"])
	date := day.String()
	if d := asMap(dr["date"]); d != nil {
		date = formatAPIDate(d)
	}
	return &HeartRateResult{
		Date:      date,
		RestingHR: bpm,
		Note:      "daily-resting-heart-rate via reconcile",
	}
}

func mapHRV(m map[string]any, day CivilDate) *HRVResult {
	h := asMap(m["dailyHeartRateVariability"])
	rmssd := parseFloatish(h["averageHeartRateVariabilityMilliseconds"])
	if rmssd == 0 {
		rmssd = parseFloatish(h["deepSleepRootMeanSquareOfSuccessiveDifferencesMilliseconds"])
	}
	date := day.String()
	if d := asMap(h["date"]); d != nil {
		date = formatAPIDate(d)
	}
	return &HRVResult{
		Date:  date,
		RMSSD: rmssd,
		Raw:   m,
		Note:  "daily-heart-rate-variability via reconcile",
	}
}

func mapWeightEntries(points []map[string]any) []map[string]any {
	out := make([]map[string]any, 0, len(points))
	for _, p := range points {
		w := asMap(p["weight"])
		grams := parseFloatish(w["weightGrams"])
		sample := asMap(w["sampleTime"])
		entry := map[string]any{
			"data_point_name": dataPointName(p),
			"weight_kg":       grams / 1000,
			"weight_grams":    grams,
			"physical_time":   strField(sample, "physicalTime"),
			"notes":           strField(w, "notes"),
		}
		if civil := asMap(sample["civilTime"]); civil != nil {
			if d := asMap(civil["date"]); d != nil {
				entry["date"] = formatAPIDate(d)
			}
		}
		out = append(out, entry)
	}
	return out
}

func formatAPIDate(d map[string]any) string {
	y := int(parseFloatish(d["year"]))
	m := int(parseFloatish(d["month"]))
	day := int(parseFloatish(d["day"]))
	if y == 0 || m == 0 || day == 0 {
		return ""
	}
	return CivilDate{Year: y, Month: time.Month(m), Day: day}.String()
}

func strField(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	v, ok := m[key].(string)
	if !ok {
		return ""
	}
	return v
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
