package logs

type Level int

const (
	Trace Level = iota + 1
	Debug
	Info
	Notice
	Warning
	Error
)

func (lv Level) String() string {
	switch lv {
	case Trace:
		return "Trace"
	case Debug:
		return "Debug"
	case Info:
		return "Info"
	case Notice:
		return "Notice"
	case Warning:
		return "Warning"
	case Error:
		return "Error"
	default:
		return "Unknown"
	}
}
