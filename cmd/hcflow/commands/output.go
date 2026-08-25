// output.go — shared formatting helpers for hcflow commands.
package commands

import (
	"fmt"

	"github.com/fatih/color"
)

var (
	green   = color.New(color.FgGreen)
	red     = color.New(color.FgRed)
	yellow  = color.New(color.FgYellow)
	cyan    = color.New(color.FgCyan)
	boldFmt = color.New(color.Bold)
	dimFmt  = color.New(color.Faint)
)

func checkMark() string { return green.Sprint("✓") }
func crossMark() string  { return red.Sprint("✗") }
func warnMark() string   { return yellow.Sprint("!") }

func success(format string, a ...any) {
	fmt.Printf("%s %s\n", checkMark(), fmt.Sprintf(format, a...))
}

func failure(format string, a ...any) {
	fmt.Printf("%s %s\n", crossMark(), fmt.Sprintf(format, a...))
}

func warn(format string, a ...any) {
	fmt.Printf("%s %s\n", warnMark(), fmt.Sprintf(format, a...))
}

func info(format string, a ...any) {
	fmt.Printf("  %s\n", fmt.Sprintf(format, a...))
}

func header(s string) string { return boldFmt.Sprint(s) }
func dim(s string) string    { return dimFmt.Sprint(s) }

func errorStyle(s string) string { return red.Sprint(s) }

func section(name string) {
	fmt.Printf("\n%s\n", header(name))
}

func kv(key, value string) {
	fmt.Printf("  %-14s %s\n", dimFmt.Sprint(key), value)
}

func kvColored(key string, value string, c *color.Color) {
	fmt.Printf("  %-14s %s\n", dimFmt.Sprint(key), c.Sprint(value))
}
