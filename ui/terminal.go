package ui

import (
	"fmt"
	ui "github.com/gizak/termui"
)

type Terminal struct {
	Logs              string
	RequestsPerSecond string
	RequestsCompleted string
	FailedRequests    string
	Widgets           *Widgets
}

type Widgets struct {
	logs              *ui.Par
	requestsPerSecond *ui.Par
	requestsCompleted *ui.Par
	failedRequests    *ui.Par
}

func NewTerminal(pauseChan chan int) (*Terminal, error) {
	err := ui.Init()
	if err != nil {
		return nil, err
	}

	terminal := &Terminal{
		Logs:              "",
		RequestsPerSecond: "0 r/s",
		RequestsCompleted: "0/0 (0%)",
		FailedRequests:    "0",
		Widgets:           NewWidgets(),
	}

	ui.Body.AddRows(
		ui.NewRow(
			ui.NewCol(4, 0, terminal.Widgets.requestsPerSecond),
			ui.NewCol(4, 0, terminal.Widgets.requestsCompleted),
			ui.NewCol(4, 0, terminal.Widgets.failedRequests),
		),
		ui.NewRow(
			ui.NewCol(12, 0, terminal.Widgets.logs),
		),
	)

	ui.Handle("/sys/kbd/q", func(ui.Event) {
		pauseChan <- 0
	})

	ui.Handle("/timer/1s", func(e ui.Event) {
		terminal.Render()
	})

	return terminal, nil
}

func (terminal *Terminal) Loop() {
	ui.Loop()
}

func (terminal *Terminal) StopLoop() {
	ui.StopLoop()
}

func (terminal *Terminal) AddLog(log []byte) {
	str := fmt.Sprintf("%s\n", log)
	terminal.Logs = str + terminal.Logs
	terminal.Render()
}

func (terminal *Terminal) SetRequestsPerSecond(rps int) {
	terminal.RequestsPerSecond = fmt.Sprintf("%v r/s", rps)
	terminal.Render()
}

func (terminal *Terminal) SetCompletedRequests(completed int, total int) {
	if total == 0 {
		return
	}

	percent := float64(completed) / float64(total) * 100
	terminal.RequestsCompleted = fmt.Sprintf("%v/%v (%.2f%%)", completed, total, percent)
	terminal.Render()
}

func (terminal *Terminal) SetFailedRequests(failed int) {
	terminal.FailedRequests = fmt.Sprintf("%v", failed)
	terminal.Render()
}

func (terminal *Terminal) Render() {
	ui.Body.Align()
	terminal.Widgets.logs.Height = ui.TermHeight() - 3
	terminal.Widgets.logs.Text = terminal.Logs
	terminal.Widgets.requestsPerSecond.Text = terminal.RequestsPerSecond
	terminal.Widgets.requestsCompleted.Text = terminal.RequestsCompleted
	terminal.Widgets.failedRequests.Text = terminal.FailedRequests
	ui.Render(ui.Body)
}

func NewWidgets() *Widgets {
	widgets := &Widgets{
		logs:              ui.NewPar(""),
		requestsPerSecond: ui.NewPar("0 r/s"),
		requestsCompleted: ui.NewPar("0/0 (0%)"),
		failedRequests:    ui.NewPar("0"),
	}
	widgets.logs.Height = ui.TermHeight() - 3
	widgets.logs.BorderLabel = "Logs"
	widgets.requestsPerSecond.Height = 3
	widgets.requestsPerSecond.BorderLabel = "Requests per second"
	widgets.requestsCompleted.Height = 3
	widgets.requestsCompleted.BorderLabel = "Requests completed"
	widgets.failedRequests.Height = 3
	widgets.failedRequests.BorderLabel = "Failed requests"
	return widgets
}
