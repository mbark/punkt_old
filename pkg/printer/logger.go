package printer

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/julienroland/usg"
	"github.com/reconquest/loreley"
	"github.com/sirupsen/logrus"
)

// Log ...
var Log = New()

type logLevel struct {
	figure string
	color  int
}

var logLevels = map[string]logLevel{
	"success":  {figure: usg.Get.Tick, color: 2},
	"warning":  {figure: usg.Get.Warning, color: 3},
	"error":    {figure: usg.Get.CrossThin, color: 1},
	"note":     {figure: usg.Get.Bullet, color: 5},
	"start":    {figure: usg.Get.Play, color: 2},
	"progress": {figure: usg.Get.CheckboxOn, color: 2},
	"done":     {figure: usg.Get.Tick, color: 2},
}

// Logger ...
type Logger struct {
	Out         io.Writer
	logPrefixes map[string]string
	timers      map[string]time.Time
}

// New ...
func New() *Logger {
	logPrefixes := make(map[string]string)
	for label, lvl := range logLevels {
		logPrefixes[label] = prefix(label, lvl)

	}

	return &Logger{
		Out:         os.Stdout,
		logPrefixes: logPrefixes,
		timers:      make(map[string]time.Time),
	}
}

func prefix(label string, lvl logLevel) string {
	var padding = 10 - len(label)
	if padding < 0 {
		padding = 0
	}

	var msg = "{fg 0}[punkt]  {fg %d}{.figure}  {underline}{.label}{reset}%" + strconv.Itoa(padding) + "s"

	prefix, _ := loreley.CompileAndExecuteToString(
		fmt.Sprintf(msg, lvl.color, ""),
		nil,
		map[string]interface{}{"figure": lvl.figure, "label": label},
	)

	return prefix
}

func (logger Logger) log(label, msg string, args ...interface{}) {
	text, err := loreley.CompileAndExecuteToString(
		fmt.Sprintf(` %s{reset}`, msg), nil, nil)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"label": label,
			"msg":   msg,
			"args":  args,
		}).WithError(err).Error("Unable to compile string!")
		return
	}

	logText := fmt.Sprintf(text, args...)

	fmt.Fprintln(logger.Out, logger.logPrefixes[label]+logText)
}

// Success ...
func (logger Logger) Success(msg string, args ...interface{}) {
	logger.log("success", msg, args...)
}

// Warning ...
func (logger Logger) Warning(msg string, args ...interface{}) {
	logger.log("warning", msg, args...)
}

// Error ...
func (logger Logger) Error(msg string, args ...interface{}) {
	logger.log("error", msg, args...)
}

// Note ...
func (logger Logger) Note(msg string, args ...interface{}) {
	logger.log("note", msg, args...)
}

// Start ...
func (logger Logger) Start(timer, msg string, args ...interface{}) {
	logger.log("start", msg, args...)
	logger.timers[timer] = time.Now()
}

// Progress ...
func (logger Logger) Progress(current, total int, msg string, args ...interface{}) {
	args = append(args, current+1, total)
	logger.log("progress", msg+" {reset}{fg 0}[%d/%d]", args...)
}

// Done ...
func (logger Logger) Done(timer, msg string, args ...interface{}) {
	timeTaken := time.Since(logger.timers[timer])

	args = append(args, timeTaken)
	logger.log("done", msg+" {reset}{fg 0}(%s){reset}", args...)
}
