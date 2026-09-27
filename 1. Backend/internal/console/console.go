package console

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	dim    = "\033[2m"
	cyan   = "\033[36m"
	green  = "\033[32m"
	yellow = "\033[33m"
	red    = "\033[31m"
	blue   = "\033[34m"
	white  = "\033[37m"
)

type Logger struct {
	out io.Writer
}

func New() *Logger {
	enableVirtualTerminal()
	return &Logger{out: os.Stdout}
}

func (l *Logger) Banner(title, env, addr, dbHost, dbPort, dbName string) {
	line := strings.Repeat("─", 46)
	fmt.Fprintf(l.out, "\n%s%s┌%s┐%s\n", bold, cyan, line, reset)
	fmt.Fprintf(l.out, "%s%s│  %-44s│%s\n", bold, cyan, title, reset)
	fmt.Fprintf(l.out, "%s%s├%s┤%s\n", bold, cyan, line, reset)
	l.row("Env", env)
	l.row("Listen", "http://localhost"+addr)
	l.row("Database", fmt.Sprintf("%s @ %s:%s", dbName, dbHost, dbPort))
	l.row("Time", time.Now().Format("15:04:05 02.01.2006"))
	fmt.Fprintf(l.out, "%s%s└%s┘%s\n\n", bold, cyan, line, reset)
}

func (l *Logger) row(key, value string) {
	fmt.Fprintf(l.out, "%s%s│%s  %s%-10s%s %s%-33s%s%s│%s\n",
		bold, cyan, reset, dim, key, reset, bold, truncate(value, 33), reset, cyan, reset)
}

func (l *Logger) OK(step, msg string) {
	fmt.Fprintf(l.out, "  %s%s✓%s  %s%-12s%s %s\n", bold, green, reset, dim, step, reset, msg)
}

func (l *Logger) Skip(step, msg string) {
	fmt.Fprintf(l.out, "  %s%s•%s  %s%-12s%s %s%s%s\n", bold, yellow, reset, dim, step, reset, dim, msg, reset)
}

func (l *Logger) Info(step, msg string) {
	fmt.Fprintf(l.out, "  %s%s›%s  %s%-12s%s %s\n", bold, blue, reset, dim, step, reset, msg)
}

func (l *Logger) Fail(step string, err error) {
	fmt.Fprintf(os.Stderr, "  %s%s✗%s  %s%-12s%s %s\n", bold, red, reset, dim, step, reset, err)
}

func (l *Logger) Ready(addr string) {
	fmt.Fprintf(l.out, "\n  %s%s●  Server tayyor%s  %s%s%s\n\n", bold, green, reset, cyan, addr, reset)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
