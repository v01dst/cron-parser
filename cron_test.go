package main

import (
	"strings"
	"testing"
	"time"
)

func fixedTime() time.Time {
	return time.Date(2026, 9, 4, 10, 30, 0, 0, time.UTC) // Friday
}

func TestParseValid(t *testing.T) {
	for _, expr := range []string{
		"* * * * *",
		"*/5 * * * *",
		"0 0 * * 0",
		"30 4 1,15 * 5",
		"0 9-17 * * mon-fri",
		"15 14 1 * *",
		"0 12 * jan *",
	} {
		if _, err := Parse(expr); err != nil {
			t.Errorf("Parse(%q) failed: %v", expr, err)
		}
	}
}

func TestParseInvalid(t *testing.T) {
	for _, expr := range []string{
		"",
		"* * * *",
		"* * * * * *",
		"60 * * * *",
		"* 24 * * *",
		"* * 32 * *",
		"* * * 13 *",
		"* * * * 7",
		"a * * * *",
	} {
		if _, err := Parse(expr); err == nil {
			t.Errorf("Parse(%q) should fail", expr)
		}
	}
}

func TestNextEvery5Minutes(t *testing.T) {
	c, _ := Parse("*/5 * * * *")
	next := c.Next(fixedTime())
	if next.Minute() != 35 || next.Hour() != 10 {
		t.Fatalf("next = %v, want 10:35", next)
	}
}

func TestNextSpecificTime(t *testing.T) {
	c, _ := Parse("30 14 * * *")
	next := c.Next(fixedTime())
	if next.Hour() != 14 || next.Minute() != 30 || next.Day() != 4 {
		t.Fatalf("next = %v, want Sep 4 14:30", next)
	}
}

func TestNextSkipsToNextDay(t *testing.T) {
	c, _ := Parse("0 9 * * *")
	next := c.Next(fixedTime())
	if next.Day() != 5 || next.Hour() != 9 {
		t.Fatalf("next = %v, want Sep 5 09:00", next)
	}
}

func TestNextSunday(t *testing.T) {
	c, _ := Parse("0 12 * * 0")
	next := c.Next(fixedTime()) // Sep 4 is Friday
	if next.Weekday() != time.Sunday || next.Day() != 6 {
		t.Fatalf("next = %v, want Sunday Sep 6", next)
	}
}

func TestNextMonthBoundary(t *testing.T) {
	c, _ := Parse("0 0 1 * *")
	next := c.Next(fixedTime())
	if next.Day() != 1 || next.Month() != time.October {
		t.Fatalf("next = %v, want Oct 1", next)
	}
}

func TestNextN(t *testing.T) {
	c, _ := Parse("*/10 * * * *")
	runs := c.NextN(fixedTime(), 3)
	if len(runs) != 3 {
		t.Fatalf("want 3 runs, got %d", len(runs))
	}
	if !runs[0].Before(runs[1]) || !runs[1].Before(runs[2]) {
		t.Fatal("runs should be ascending")
	}
	for _, r := range runs {
		if r.Minute()%10 != 0 {
			t.Fatalf("run %v not aligned to 10 minutes", r)
		}
	}
}

func TestMatches(t *testing.T) {
	c, _ := Parse("*/15 * * * *")
	if !c.Matches(time.Date(2026, 9, 4, 10, 45, 0, 0, time.UTC)) {
		t.Error("10:45 should match */15")
	}
	if c.Matches(time.Date(2026, 9, 4, 10, 31, 0, 0, time.UTC)) {
		t.Error("10:31 should not match */15")
	}
}

func TestDayNamesAndRanges(t *testing.T) {
	c, _ := Parse("0 9 * * mon-fri")
	next := c.Next(fixedTime())
	if next.Weekday() != time.Monday {
		t.Errorf("weekday = %v, want Monday", next.Weekday())
	}
	if next.Day() != 7 { // next Monday is Sep 7
		t.Fatalf("next = %v, want Monday Sep 7", next)
	}
}

func TestDayOrWeekdayRestriction(t *testing.T) {
	// Both restricted: OR semantics (vixie cron)
	c, _ := Parse("0 0 13 * 5") // 13th OR Friday
	next := c.Next(fixedTime()) // Fri Sep 4 10:30 → next is Fri Sep 4? No — must be after.
	// Sep 4 is a Friday; the run for Sep 4 00:00 already passed → next Friday Sep 11.
	if next.Day() != 11 {
		t.Fatalf("next = %v, want Sep 11 (OR semantics)", next)
	}
}

func TestValidateCrontabLines(t *testing.T) {
	good := []string{"*/5 * * * *", "0 9 * * mon-fri", "# comment", "", "@reboot x"}
	for _, line := range good {
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		if _, err := Parse(strings.Join(fields[:5], " ")); err != nil {
			t.Errorf("Parse(%q): %v", line, err)
		}
	}
	bad := []string{"* * * *", "60 * * * *", "not a cron"}
	for _, line := range bad {
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		if _, err := Parse(strings.Join(fields[:5], " ")); err == nil {
			t.Errorf("Parse(%q) should fail", line)
		}
	}
}
