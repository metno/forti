package radar

type Coverage int

const (
	OK                     Coverage = 0
	TemporarilyUnavailable Coverage = 1
	NoCoverage             Coverage = 2
	UnknownCoverage        Coverage = -1
)

func (c Coverage) String() string {
	switch c {
	case OK:
		return "ok"
	case TemporarilyUnavailable:
		return "temporarily unavailable"
	case NoCoverage:
		return "no coverage"
	}
	return "unknown"
}
