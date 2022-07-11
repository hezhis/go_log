package logger

import "testing"

func TestLogger(t *testing.T) {
	InitLogger("testlogger")
	Debug("hello %s", "world")
	Info("hello %s", "world")
	Warn("hello %s", "world")
	Error("hello %s", "world")
}
