// doctor.go — hcflow doctor
package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/hickstein/hcflow/internal/doctor"
)

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose the hcflow configuration and environment",
		Long: `Run a comprehensive health check of the repository's hcflow setup.

Checks:
  • git installation
  • gh installation and authentication
  • GitHub remote
  • .hcflow.yml presence and schema
  • workflow files
  • merge strategy
  • PR template

Note: doctor reads but does not modify any state.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := os.Getwd()
			if err != nil {
				return err
			}

			report := doctor.Run(dir)
			printDoctorReport(report)

			if !report.Healthy {
				return fmt.Errorf("doctor found issues — see above")
			}
			return nil
		},
	}
}

func printDoctorReport(report *doctor.Report) {
	currentCategory := ""

	for _, f := range report.Findings {
		if f.Category != currentCategory {
			section(f.Category)
			currentCategory = f.Category
		}

		var mark string
		var msgColor func(string, ...interface{}) string
		switch f.Severity {
		case doctor.SeverityOK:
			mark = checkMark()
			msgColor = func(s string, a ...interface{}) string {
				if len(a) > 0 {
					return fmt.Sprintf(s, a...)
				}
				return s
			}
		case doctor.SeverityWarning:
			mark = warnMark()
			msgColor = func(s string, a ...interface{}) string {
				if len(a) > 0 {
					return yellow.Sprintf(s, a...)
				}
				return yellow.Sprint(s)
			}
		case doctor.SeverityError:
			mark = crossMark()
			msgColor = func(s string, a ...interface{}) string {
				if len(a) > 0 {
					return red.Sprintf(s, a...)
				}
				return red.Sprint(s)
			}
		case doctor.SeverityInfo:
			mark = "·"
			msgColor = func(s string, a ...interface{}) string {
				if len(a) > 0 {
					return fmt.Sprintf(s, a...)
				}
				return s
			}
		}

		fmt.Printf("  %s %-26s %s\n", mark, dimFmt.Sprint(f.Name), msgColor(f.Message))
	}

	fmt.Println()
	if report.Healthy {
		fmt.Printf("%s Result: %s\n\n", checkMark(), green.Sprint("healthy"))
	} else {
		fmt.Printf("%s Result: %s\n\n", crossMark(), red.Sprint(report.Summary()))
	}
}
