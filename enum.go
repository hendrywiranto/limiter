package limiter

type Duration uint8

const (
	DurationUnknown Duration = iota
	DurationSecond
	DurationMinute
	DurationHour
	DurationDay
)

func (d Duration) Seconds() int64 {
	switch d {
	case DurationSecond:
		return 1
	case DurationMinute:
		return 60
	case DurationHour:
		return 3600
	case DurationDay:
		return 86400
	default:
		return 0
	}
}
