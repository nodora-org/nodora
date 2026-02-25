package time

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	gotime "time"
)

var durationRe = regexp.MustCompile(`(\d+)(ms|[smhdw])`)

func parseDuration(s string) (gotime.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty duration string")
	}

	matches := durationRe.FindAllStringSubmatch(s, -1)
	if len(matches) == 0 {
		return 0, fmt.Errorf("invalid duration format: %q", s)
	}

	// verify the entire string is consumed by matches
	var combined strings.Builder
	for _, m := range matches {
		combined.WriteString(m[0])
	}
	if combined.String() != s {
		return 0, fmt.Errorf("invalid duration format: %q", s)
	}

	var total gotime.Duration
	for _, m := range matches {
		n, _ := strconv.ParseInt(m[1], 10, 64)
		d, err := unitToDuration(m[2])
		if err != nil {
			return 0, err
		}
		total += gotime.Duration(n) * d
	}
	return total, nil
}

func unitToDuration(unit string) (gotime.Duration, error) {
	switch unit {
	case "ms":
		return gotime.Millisecond, nil
	case "s":
		return gotime.Second, nil
	case "m":
		return gotime.Minute, nil
	case "h":
		return gotime.Hour, nil
	case "d":
		return 24 * gotime.Hour, nil
	case "w":
		return 7 * 24 * gotime.Hour, nil
	default:
		return 0, fmt.Errorf("unsupported duration unit: %q", unit)
	}
}

func msToTime(ms float64) gotime.Time {
	sec := int64(ms / 1000)
	nsec := int64((ms - float64(sec)*1000) * 1e6)
	return gotime.Unix(sec, nsec).UTC()
}

func timeToMs(t gotime.Time) float64 {
	return float64(t.UnixMilli())
}
