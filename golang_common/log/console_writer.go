package log

import (
	"fmt"
	"os"
)

type colorRecord Record

func (r *colorRecord) String() string {
	switch r.level {
	case TRACE:
		return fmt.Sprintf("\033[36m%s\033[0m [\033[34m%s\033[0m] \033[47;30m%s\033[0m %s\n",
			r.time, LEVEL_FLAGS[r.level], r.code, r.info)
	case DEBUG:
		return fmt.Sprintf("\033[36m%s\033[0m [\033[34m%s\033[0m] \033[47;30m%s\033[0m %s\n",
			r.time, LEVEL_FLAGS[r.level], r.code, r.info)

	case INFO:
		return fmt.Sprintf("\033[36m%s\033[0m [\033[32m%s\033[0m] \033[47;30m%s\033[0m %s\n",
			r.time, LEVEL_FLAGS[r.level], r.code, r.info)

	case WARNING:
		return fmt.Sprintf("\033[36m%s\033[0m [\033[33m%s\033[0m] \033[47;30m%s\033[0m %s\n",
			r.time, LEVEL_FLAGS[r.level], r.code, r.info)

	case ERROR:
		return fmt.Sprintf("\033[36m%s\033[0m [\033[31m%s\033[0m] \033[47;30m%s\033[0m %s\n",
			r.time, LEVEL_FLAGS[r.level], r.code, r.info)

	case FATAL:
		return fmt.Sprintf("\033[36m%s\033[0m [\033[35m%s\033[0m] \033[47;30m%s\033[0m %s\n",
			r.time, LEVEL_FLAGS[r.level], r.code, r.info)
	}

	return ""
}

type ConsoleWriter struct {
	color bool
}

func NewConsoleWriter() *ConsoleWriter {
	return &ConsoleWriter{}
}

func (w *ConsoleWriter) Write(r *Record) error {
	if w.color {
		fmt.Fprint(os.Stdout, ((*colorRecord)(r)).String())
	} else {
		fmt.Fprint(os.Stdout, r.String())
	}
	return nil
}

func (w *ConsoleWriter) Init() error {
	return nil
}

func (w *ConsoleWriter) SetColor(c bool) {
	w.color = c
}
