/*
	NOTE: The implementation in this file is loosely based on
	https://github.com/verdverm/httopd/
*/

package main

import (
	"fmt"
	"sort"
	"time"

	"github.com/nsf/termbox-go"

	"github.com/ekarlso/gomdns/stats"
)

const coldef = termbox.ColorDefault
const DATEPRINT = "Jan 02, 2006 15:04:05"

const FORMAT = "%-20s %-20s %-10s %-10s %-10s %-10s"

var startTime time.Time

func init() {
	startTime = time.Now()
}

var w, h int

var selectedRow = 0
var colHeaderRow = 7
var minSelectedRow = colHeaderRow + 1
var maxSelectedRow = colHeaderRow + 1

func redraw() {
	termbox.Clear(coldef, coldef)
	w, h = termbox.Size()

	drawCurrentTime(1, 0)
	drawHeaders(1, colHeaderRow)

	y := colHeaderRow + 1
	drawStats(1, y, meters)
	maxSelectedRow = y - 1

	drawFooter()
	termbox.HideCursor()

	tbprint(w-6, h-1, coldef, termbox.ColorBlue, "ʕ◔ϖ◔ʔ")
	termbox.Flush()
}

func fill(x, y, w, h int, cell termbox.Cell) {
	for ly := 0; ly < h; ly++ {
		for lx := 0; lx < w; lx++ {
			termbox.SetCell(x+lx, y+ly, cell.Ch, cell.Fg, cell.Bg)
		}
	}
}

func tbprint(x, y int, fg, bg termbox.Attribute, msg string) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, fg, bg)
		x++
	}
}

func drawCurrentTime(x, y int) {
	now := time.Now()
	since := now.Sub(startTime)
	h := int(since.Hours())
	m := int(since.Minutes()) % 60
	s := int(since.Seconds()) % 60
	timeStr := fmt.Sprintf("Now:  %-24s  Watching:  %3d:%02d:%02d", now.Format(DATEPRINT), h, m, s)
	for i, c := range timeStr {
		termbox.SetCell(x+i, y, c, coldef, coldef)
	}
}

func prettyFloat(fl float64) string {
	return fmt.Sprintf("%.1f", fl)
}

func drawStats(x, y int, meters map[string]stats.Meter) {
	if selectedRow < minSelectedRow {
		selectedRow = minSelectedRow
	}

	fg_col, bg_col := coldef, coldef
	if y == selectedRow {
		fg_col = termbox.ColorBlack
		bg_col = termbox.ColorYellow
	}

	var keys []string
	for k, v := range meters {
		if v.IsValid() {
			keys = append(keys, k)
		}
	}

	sort.Strings(keys)

	xcnt := x
	for _, k := range keys {
		m := meters[k]

		/*k = fmt.Sprintf("%-20s", k)
		for _, c := range k {
			termbox.SetCell(xcnt, y, c, fg_col, bg_col)
			xcnt++
		}*/

		str := fmt.Sprintf(
			FORMAT,
			k,
			prettyFloat(m.Rate1),
			prettyFloat(m.Rate5),
			prettyFloat(m.Rate15),
			prettyFloat(m.RateMean),
			fmt.Sprintf("%d", m.Count))

		for _, c := range str {
			termbox.SetCell(xcnt, y, c, fg_col, bg_col)
			xcnt++
		}
		xcnt = x
		y++
	}
	y++
}

func drawHeaders(x, y int) {
	columnHeaders := fmt.Sprintf(
		FORMAT,
		"Type", "1m", "5m", "15m", "Mean", "Count",
	)

	for i := 0; i < w; i++ {
		termbox.SetCell(i, y, ' ', coldef, termbox.ColorBlue)
	}
	for i, c := range columnHeaders {
		termbox.SetCell(x+i, y, c, coldef, termbox.ColorBlue)
	}
}

func drawFooter() {
	footerText := " Esc: Quit  Ctrl-Q:Quit " //"<_sort_> "

	for i := 0; i < w; i++ {
		termbox.SetCell(i, h-1, ' ', coldef, termbox.ColorBlue)
	}
	for i, c := range footerText {
		termbox.SetCell(i, h-1, c, coldef, termbox.ColorBlue)
	}

}
