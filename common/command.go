package common

import (
	"fmt"
	"os"
	"strings"

	"github.com/ryanuber/columnize"
	"github.com/spf13/cobra"
)

type ICommandResult interface {
	GetOutput() string
}

// OutputFormatter is the standardized interface all output formatters
// should use
type OutputFormatter interface {
	// SetError sets the encountered error
	SetError(err error)

	// SetCommandResult sets the result of the command execution
	SetCommandResult(result ICommandResult)

	// WriteOutput writes the previously set result / error output
	WriteOutput()

	// WriteCommandResult immediately writes the given command result without waiting for WriteOutput func call.
	WriteCommandResult(result ICommandResult)

	// Write extends io.Writer interface
	Write(p []byte) (n int, err error)
}

type CmdResults []ICommandResult

func (r CmdResults) GetOutput() string {
	outputs := make([]string, len(r))
	for i, o := range r {
		outputs[i] = o.GetOutput()
	}

	return strings.Join(outputs, "\n")
}

func FormatKV(in []string) string {
	columnConf := columnize.DefaultConfig()
	columnConf.Empty = "<none>"
	columnConf.Glue = " = "

	return columnize.Format(in, columnConf)
}

type commonOutputFormatter struct {
	errorOutput   error
	commandOutput ICommandResult
}

func (c *commonOutputFormatter) SetError(err error) {
	c.errorOutput = err
}

func (c *commonOutputFormatter) SetCommandResult(result ICommandResult) {
	c.commandOutput = result
}

func InitializeOutputter(cmd *cobra.Command) OutputFormatter {
	return &cliOutput{}
}

type cliOutput struct {
	commonOutputFormatter
}

// WriteOutput implements OutputFormatter interface
func (cli *cliOutput) WriteOutput() {
	if cli.errorOutput != nil {
		_, _ = fmt.Fprintln(os.Stderr, cli.getErrorOutput())

		// return proper error exit code for cli error output
		os.Exit(1)
	}

	_, _ = fmt.Fprintln(os.Stdout, cli.getCommandOutput())
}

// WriteCommandResult implements OutputFormatter interface
func (cli *cliOutput) WriteCommandResult(result ICommandResult) {
	_, _ = fmt.Fprintln(os.Stdout, result.GetOutput())
}

// WriteOutput implements OutputFormatter plus io.Writer interfaces
func (cli *cliOutput) Write(p []byte) (n int, err error) {
	return os.Stdout.Write(p)
}

func (cli *cliOutput) getErrorOutput() string {
	if cli.errorOutput == nil {
		return ""
	}

	return cli.errorOutput.Error()
}

func (cli *cliOutput) getCommandOutput() string {
	if cli.commandOutput == nil {
		return ""
	}

	return cli.commandOutput.GetOutput()
}
