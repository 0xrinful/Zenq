package parse

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var NumberRe = regexp.MustCompile(`\d+(\.\d+)?`)

func ParseChapterNumber(raw string) (float64, error) {
	match := NumberRe.FindString(raw)
	if match == "" {
		return 0, fmt.Errorf("no number found in %q", raw)
	}
	return strconv.ParseFloat(match, 64)
}

var arabicMonths = map[string]time.Month{
	"يناير":  time.January,
	"فبراير": time.February,
	"مارس":   time.March,
	"أبريل":  time.April,
	"ابريل":  time.April,
	"مايو":   time.May,
	"يونيو":  time.June,
	"يوليو":  time.July,
	"أغسطس":  time.August,
	"اغسطس":  time.August,
	"سبتمبر": time.September,
	"أكتوبر": time.October,
	"اكتوبر": time.October,
	"نوفمبر": time.November,
	"ديسمبر": time.December,
}

func ParseDate(raw string) (time.Time, error) {
	raw = strings.TrimSpace(strings.ToLower(raw))
	now := time.Now()

	// Handle Arabic explicit singular/dual phrases first
	switch {
	case strings.Contains(raw, "دقيقة واحدة"):
		return now.Add(-1 * time.Minute), nil
	case strings.Contains(raw, "دقيقتين") || strings.Contains(raw, "دقيقتان"):
		return now.Add(-2 * time.Minute), nil

	case strings.Contains(raw, "ساعة واحدة"):
		return now.Add(-1 * time.Hour), nil
	case strings.Contains(raw, "ساعتين") || strings.Contains(raw, "ساعتان"):
		return now.Add(-2 * time.Hour), nil

	case strings.Contains(raw, "يوم واحد"):
		return now.AddDate(0, 0, -1), nil
	case strings.Contains(raw, "يومين") || strings.Contains(raw, "يومان"):
		return now.AddDate(0, 0, -2), nil
	}

	// Relative dates with digits
	if n := NumberRe.FindString(raw); n != "" {
		num, _ := strconv.Atoi(n)

		switch {
		case strings.Contains(raw, "minute"),
			strings.Contains(raw, "minutes"),
			strings.Contains(raw, "دقيقة"),
			strings.Contains(raw, "دقائق"):
			return now.Add(-time.Duration(num) * time.Minute), nil

		case strings.Contains(raw, "hour"),
			strings.Contains(raw, "hours"),
			strings.Contains(raw, "ساعة"),
			strings.Contains(raw, "ساعات"):
			return now.Add(-time.Duration(num) * time.Hour), nil

		case strings.Contains(raw, "day"),
			strings.Contains(raw, "days"),
			strings.Contains(raw, "يوم"),
			strings.Contains(raw, "أيام"):
			return now.AddDate(0, 0, -num), nil

		case strings.Contains(raw, "week"),
			strings.Contains(raw, "weeks"),
			strings.Contains(raw, "أسبوع"),
			strings.Contains(raw, "أسابيع"):
			return now.AddDate(0, 0, -(num * 7)), nil

		case strings.Contains(raw, "month"),
			strings.Contains(raw, "months"),
			strings.Contains(raw, "شهر"),
			strings.Contains(raw, "أشهر"):
			return now.AddDate(0, -num, 0), nil

		case strings.Contains(raw, "year"),
			strings.Contains(raw, "years"),
			strings.Contains(raw, "سنة"),
			strings.Contains(raw, "سنوات"):
			return now.AddDate(-num, 0, 0), nil
		}
	}

	// yyyy-mm-dd
	if t, err := time.Parse("2006-01-02", raw); err == nil {
		return t, nil
	}

	// يناير 01, 2025
	parts := strings.Fields(raw)

	if len(parts) == 3 {
		month, ok := arabicMonths[parts[0]]
		if ok {
			day, err := strconv.Atoi(strings.TrimSuffix(parts[1], ","))
			if err != nil {
				return time.Time{}, err
			}

			year, err := strconv.Atoi(parts[2])
			if err != nil {
				return time.Time{}, err
			}

			return time.Date(year, month, day, 0, 0, 0, 0, time.UTC), nil
		}
	}

	return time.Time{}, fmt.Errorf("unknown date format: %q", raw)
}
