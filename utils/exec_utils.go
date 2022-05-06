package utils

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	. "github.com/puppetlabs/regulator/rgerror"
	"github.com/puppetlabs/regulator/sanitize"
)

func ExecReadOutput(command_string string, params ...string) (string, string, *RGerror) {
	shell_command := exec.Command(command_string, params...)
	shell_command.Env = os.Environ()
	var stdout, stderr bytes.Buffer
	shell_command.Stdout = &stdout
	shell_command.Stderr = &stderr
	err := shell_command.Run()
	output := sanitize.ReplaceAllNewlines(string(stdout.Bytes()))
	logs := sanitize.ReplaceAllNewlines(string(stderr.Bytes()))
	if err != nil {
		return output, logs, &RGerror{
			ShellError,
			fmt.Sprintf("Command '%s' failed:\n%s\nstderr:\n%s", shell_command, err, logs),
			err,
		}
	}
	return output, logs, nil
}

func BuildAndRunCommand(executable string, script string, args []string) (string, string, *RGerror) {
	var output, logs string
	var airr *RGerror
	if executable == "" {
		// Script is directly executable
		output, logs, airr = ExecReadOutput(script, args...)
	} else {
		args = append([]string{script}, args...)
		output, logs, airr = ExecReadOutput(executable, args...)
	}
	if airr != nil {
		return output, logs, airr
	}

	return output, logs, nil
}
