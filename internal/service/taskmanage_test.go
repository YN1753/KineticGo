package service

import (
	"testing"
	"time"
)

func TestLastDueBefore(t *testing.T) {
	sched, err := secondParser.Parse("0 0 8 * * *")
	if err != nil {
		t.Fatal(err)
	}

	// 2026-01-15 10:00:00 → 最近应触发点为当天 08:00（窗口 3h 内）
	now := time.Date(2026, 1, 15, 10, 0, 0, 0, time.Local)
	last := lastDueBefore(sched, now, 3*time.Hour)
	want := time.Date(2026, 1, 15, 8, 0, 0, 0, time.Local)
	if !last.Equal(want) {
		t.Fatalf("lastDueBefore = %v, want %v", last, want)
	}

	// 07:00 时今日 08:00 尚未到点；昨日 08:00 已超出 3h 窗口 → 不应补跑
	now2 := time.Date(2026, 1, 15, 7, 0, 0, 0, time.Local)
	last2 := lastDueBefore(sched, now2, 3*time.Hour)
	if !last2.IsZero() {
		t.Fatalf("lastDueBefore at 07:00 should be zero, got %v", last2)
	}

	// 12:00 时距 08:00 已 4h，超出窗口 → 零值
	now3 := time.Date(2026, 1, 15, 12, 0, 0, 0, time.Local)
	if last3 := lastDueBefore(sched, now3, 3*time.Hour); !last3.IsZero() {
		t.Fatalf("lastDueBefore at 12:00 should be zero (out of window), got %v", last3)
	}
}

func TestCompareVersions(t *testing.T) {
	cases := []struct {
		latest, current string
		want            int
	}{
		{"v1.2.3", "1.2.3", 0},
		{"v1.2.4", "1.2.3", 1},
		{"v1.2.2", "1.2.3", -1},
		{"v2.0.0", "1.9.9", 1},
		{"1.10.0", "1.9.0", 1},
	}
	for _, c := range cases {
		if got := compareVersions(c.latest, c.current); got != c.want {
			t.Errorf("compareVersions(%q, %q) = %d, want %d", c.latest, c.current, got, c.want)
		}
	}
}

func TestShouldRetryTrigger(t *testing.T) {
	for _, trig := range []string{"schedule", "catchup", "retry"} {
		if !shouldRetryTrigger(trig) {
			t.Errorf("shouldRetryTrigger(%q) = false, want true", trig)
		}
	}
	if shouldRetryTrigger("manual") || shouldRetryTrigger("immediate") {
		t.Error("manual/immediate should not auto-retry")
	}
}

func TestSecondParserAcceptsPresets(t *testing.T) {
	for _, expr := range []string{
		"*/10 * * * * *",
		"0 * * * * *",
		"0 */5 * * * *",
		"0 0 * * * *",
	} {
		if _, err := secondParser.Parse(expr); err != nil {
			t.Errorf("preset %q: %v", expr, err)
		}
	}
}
