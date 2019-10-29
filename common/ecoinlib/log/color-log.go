package log

import (
	"fmt"
	"time"
)

const (
	RED = uint8(iota + 91)
	GREEN	// 92
	YELLOW	// 93
	BLUE	// 94
	MAGENTA		// 95

	INFO = "[INFO]"
	TRAC = "[TRAC]"
	ERRO = "[ERRO]"
	WARN = "[WARN]"
	SUCC = "[SUCC]"
)

func Trace(format string, a ...interface{}) {
	prefix := yellow(TRAC)
	fmt.Println(formatLog(prefix), fmt.Sprintf(format, a...))
}

func Info(format string, a ...interface{}) {
	prefix := blue(INFO)
	fmt.Println(formatLog(prefix), fmt.Sprintf(format, a...))
}

func Success(format string, a ...interface{}) {
	prefix := green(SUCC)
	fmt.Println(formatLog(prefix), fmt.Sprintf(format, a...))
}

func Warn(format string, a ...interface{}) {
	prefix := megenta(WARN)
	fmt.Println(formatLog(prefix), fmt.Sprintf(format, a...))
	// TODO: 增加退出函数体操作
}

func Error(format string, a ...interface{}) {
	prefix := red(ERRO)
	fmt.Println(formatLog(prefix), fmt.Sprintf(format, a...))
	// TODO: 增加退出进程操作
}

func red(s string) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", RED, s)
}
func green(s string) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", GREEN, s)
}
func yellow(s string) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", YELLOW, s)
}
func blue(s string) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", BLUE, s)
}
func megenta(s string) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", MAGENTA, s)
}

func formatLog(prefix string) string {
	return time.Now().Format("2006/01/02 15:04:05") + " " + prefix + " "
}
