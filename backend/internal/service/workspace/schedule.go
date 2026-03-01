package workspace

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// NormalizeTimezone 标准化时区，空值回退到默认时区。
func NormalizeTimezone(tz string) string {
	value := strings.TrimSpace(tz)
	if value == "" {
		return "Asia/Shanghai"
	}
	return value
}

func parseTimeInLocation(raw string, location *time.Location) (time.Time, error) {
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04",
		"2006-01-02 15:04:05",
	}
	var lastErr error
	for _, layout := range layouts {
		t, err := time.ParseInLocation(layout, raw, location)
		if err == nil {
			return t, nil
		}
		lastErr = err
	}
	return time.Time{}, lastErr
}

// ComputeNextRunAt 根据触发配置计算下一次运行时间。
func ComputeNextRunAt(triggerType, triggerValue, timezone string, now time.Time) (*time.Time, error) {
	mode := strings.TrimSpace(triggerType)
	value := strings.TrimSpace(triggerValue)
	location, err := time.LoadLocation(NormalizeTimezone(timezone))
	if err != nil {
		return nil, fmt.Errorf("invalid timezone: %w", err)
	}

	switch mode {
	case "", "manual":
		return nil, nil
	case "once":
		if value == "" {
			return nil, fmt.Errorf("triggerValue required when triggerType=once")
		}
		parsed, err := parseTimeInLocation(value, location)
		if err != nil {
			return nil, fmt.Errorf("invalid once triggerValue: %w", err)
		}
		return &parsed, nil
	case "daily":
		if value == "" {
			return nil, fmt.Errorf("triggerValue required when triggerType=daily (HH:MM)")
		}
		parts := strings.Split(value, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid daily triggerValue, expected HH:MM")
		}
		hour, err := strconv.Atoi(parts[0])
		if err != nil || hour < 0 || hour > 23 {
			return nil, fmt.Errorf("invalid daily hour")
		}
		minute, err := strconv.Atoi(parts[1])
		if err != nil || minute < 0 || minute > 59 {
			return nil, fmt.Errorf("invalid daily minute")
		}
		base := now.In(location)
		next := time.Date(base.Year(), base.Month(), base.Day(), hour, minute, 0, 0, location)
		if !next.After(base) {
			next = next.Add(24 * time.Hour)
		}
		return &next, nil
	case "interval_hours":
		if value == "" {
			return nil, fmt.Errorf("triggerValue required when triggerType=interval_hours")
		}
		hours, err := strconv.Atoi(value)
		if err != nil || hours <= 0 || hours > 720 {
			return nil, fmt.Errorf("invalid interval hours, expected 1~720")
		}
		next := now.In(location).Add(time.Duration(hours) * time.Hour)
		return &next, nil
	default:
		return nil, fmt.Errorf("unsupported triggerType: %s", mode)
	}
}

// DefaultAsyncStatus 触发类型对应的默认异步状态。
func DefaultAsyncStatus(triggerType string) string {
	switch strings.TrimSpace(triggerType) {
	case "", "manual":
		return "idle"
	default:
		return "scheduled"
	}
}
