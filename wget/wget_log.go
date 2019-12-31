package wget

import (
	"fmt"
	"os"
)

type log struct {

	info *os.File
	err *os.File
}

var (
	gLog = newLog()
)

func newLog() *log {

	return &log{
		info:os.Stdout,
		err:os.Stderr,
	}
}

func (l *log) logInfo(format string, a ...interface{})  {

	fmt.Fprintf(l.info, format, a...)
}

func (l *log) logErr(format string, a ...interface{})  {

	fmt.Fprintf(l.err, format, a...)
}