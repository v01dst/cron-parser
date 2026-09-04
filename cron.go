package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var ErrBadExpr = errors.New("invalid cron expression")

type field struct {
	min, max int
	values   map[int]bool
}

func (f field) matches(v int) bool { return f.values[v] }

func parseField(expr string, min, max int, names map[string]int) (field, error) {
	f := field{min: min, max: max, values: map[int]bool{}}
	if expr == "*" {
		for i := min; i <= max; i++ {
			f.values[i] = true
		}
		return f, nil
	}

	for _, part := range strings.Split(expr, ",") {
		var base string
		var step int = 1
		if idx := strings.Index(part, "/"); idx >= 0 {
			base = part[:idx]
			stepStr := part[idx+1:]
			s, err := strconv.Atoi(stepStr)
			if err != nil || s <= 0 {
				return f, fmt.Errorf("%w: bad step %q", ErrBadExpr, stepStr)
			}
			step = s
		} else {
			base = part
		}

		lo, hi, hasRange, err := resolveBounds(base, min, max, names)
		if err != nil {
			return f, err
		}

		if !hasRange && step > 1 && base != "*" {
			// "5/10" means start at 5, step 10 (to max).
			for i := lo; i <= max; i += step {
				f.values[i] = true
			}
			continue
		}

		if step > max-min+1 {
			return f, fmt.Errorf("%w: step %d too large", ErrBadExpr, step)
		}
		for i := lo; i <= hi; i += step {
			f.values[i] = true
		}
	}

	if len(f.values) == 0 {
		return f, fmt.Errorf("%w: field %q selects nothing", ErrBadExpr, expr)
	}
	return f, nil
}

func resolveBounds(base string, min, max int, names map[string]int) (lo, hi int, hasRange bool, err error) {
	if base == "*" {
		return min, max, true, nil
	}
	if strings.Contains(base, "-") {
		parts := strings.SplitN(base, "-", 2)
		lo, err = resolveValue(parts[0], names)
		if err != nil {
			return 0, 0, false, err
		}
		hi, err = resolveValue(parts[1], names)
		if err != nil {
			return 0, 0, false, err
		}
		if lo > hi || lo < min || hi > max {
			return 0, 0, false, fmt.Errorf("%w: range %q out of bounds [%d,%d]", ErrBadExpr, base, min, max)
		}
		return lo, hi, true, nil
	}
	lo, err = resolveValue(base, names)
	if err != nil {
		return 0, 0, false, err
	}
	if lo < min || lo > max {
		return 0, 0, false, fmt.Errorf("%w: value %d out of bounds [%d,%d]", ErrBadExpr, lo, min, max)
	}
	return lo, lo, false, nil
}

func resolveValue(s string, names map[string]int) (int, error) {
	if names != nil {
		if v, ok := names[strings.ToLower(s)]; ok {
			return v, nil
		}
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("%w: bad value %q", ErrBadExpr, s)
	}
	return v, nil
}

var dayNames = map[string]int{
	"sun": 0, "mon": 1, "tue": 2, "wed": 3, "thu": 4, "fri": 5, "sat": 6,
}

var monNames = map[string]int{
	"jan": 1, "feb": 2, "mar": 3, "apr": 4, "may": 5, "jun": 6,
	"jul": 7, "aug": 8, "sep": 9, "oct": 10, "nov": 11, "dec": 12,
}

// Cron is a parsed 5-field cron expression (minute hour day month weekday).
type Cron struct {
	minute, hour, day, month, weekday field
	Original                          string
}

func Parse(expr string) (*Cron, error) {
	fields := strings.Fields(strings.TrimSpace(expr))
	if len(fields) != 5 {
		return nil, fmt.Errorf("%w: need 5 fields, got %d", ErrBadExpr, len(fields))
	}

	var c Cron
	var err error
	if c.minute, err = parseField(fields[0], 0, 59, nil); err != nil {
		return nil, err
	}
	if c.hour, err = parseField(fields[1], 0, 23, nil); err != nil {
		return nil, err
	}
	if c.day, err = parseField(fields[2], 1, 31, nil); err != nil {
		return nil, err
	}
	if c.month, err = parseField(fields[3], 1, 12, monNames); err != nil {
		return nil, err
	}
	if c.weekday, err = parseField(fields[4], 0, 6, dayNames); err != nil {
		return nil, err
	}
	c.Original = expr
	return &c, nil
}

// Matches reports whether t satisfies the expression.
func (c *Cron) Matches(t time.Time) bool {
	return c.minute.matches(t.Minute()) &&
		c.hour.matches(t.Hour()) &&
		c.month.matches(int(t.Month())) &&
		(c.day.matches(t.Day()) || c.weekday.matches(int(t.Weekday())))
}

// Next returns the next matching time strictly after `from`.
// Day-of-month and weekday are OR-ed when both are restricted (vixie cron).
func (c *Cron) Next(from time.Time) time.Time {
	dayRestricted := c.Original != "" && dayFieldRestricted(c)
	t := from.Truncate(time.Minute).Add(time.Minute)

	limit := from.AddDate(5, 0, 0)
	for t.Before(limit) {
		if !c.month.matches(int(t.Month())) {
			// jump to first day of next month
			t = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location()).AddDate(0, 1, 0)
			continue
		}

		dayOK := c.day.matches(t.Day())
		wdOK := c.weekday.matches(int(t.Weekday()))
		if dayRestricted {
			dayOK = dayOK || wdOK
		} else {
			dayOK = dayOK && wdOK
		}
		if !dayOK {
			t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).AddDate(0, 0, 1)
			continue
		}
		if !c.hour.matches(t.Hour()) {
			t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location()).Add(time.Hour)
			continue
		}
		if !c.minute.matches(t.Minute()) {
			t = t.Add(time.Minute)
			continue
		}
		return t
	}
	return time.Time{}
}

// NextN returns up to n upcoming runs.
func (c *Cron) NextN(from time.Time, n int) []time.Time {
	out := make([]time.Time, 0, n)
	cur := from
	for i := 0; i < n; i++ {
		next := c.Next(cur)
		if next.IsZero() {
			break
		}
		out = append(out, next)
		cur = next
	}
	return out
}

func dayFieldRestricted(c *Cron) bool {
	// In vixie cron, when both day and weekday are restricted, they OR.
	// We approximate: both restricted if neither is "*".
	fields := strings.Fields(c.Original)
	if len(fields) != 5 {
		return false
	}
	return fields[2] != "*" && fields[4] != "*"
}
