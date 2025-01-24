package logger

type UI interface {
	DisplayLog(format string, args ...interface{})
}
