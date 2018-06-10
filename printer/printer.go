package printer

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/gosuri/uilive"
)

// Printer ...
type Printer interface {
	Start()
	Next()
	Finish(error)
}

type printer struct {
	writer   *uilive.Writer
	task     string
	subtasks []string
	active   int
}

// NewPrinter ...
func NewPrinter(task string, subtasks []string) Printer {
	return &printer{
		writer:   uilive.New(),
		task:     task,
		subtasks: subtasks,
		active:   0,
	}
}

func (prtr printer) printSubtask() {
	if prtr.active < len(prtr.subtasks) {
		progress := fmt.Sprintf("[%d/%d]", prtr.active+1, len(prtr.subtasks))
		fmt.Printf("%s %s\n", color.HiBlackString(progress), color.BlueString(prtr.subtasks[prtr.active]))
	}
}

func (prtr *printer) Start() {
	prtr.printSubtask()
}

func (prtr *printer) Next() {
	prtr.active++
	prtr.printSubtask()
}

func (prtr *printer) Finish(err error) {
	var coloredTasks []string
	for i := range prtr.subtasks {
		coloredTasks = append(coloredTasks, color.BlueString(prtr.subtasks[i]))
	}

	tasksDone := strings.Join(coloredTasks, ", ")

	if err != nil {
		fmt.Printf("%s: %v\n", color.RedString("error"), err)
	} else {
		fmt.Printf("%s ensure finished: %s\n", color.GreenString("success"), tasksDone)
	}
}
