package localexec

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/puppetlabs/regulator/localfile"
	"github.com/puppetlabs/regulator/rgerror"
	"github.com/puppetlabs/regulator/sanitize"
)

func ExecReadOutput(executable string, args []string) (string, string, *rgerror.RGerror) {
	shell_command := exec.Command(executable, args...)
	shell_command.Env = os.Environ()
	var stdout, stderr bytes.Buffer
	shell_command.Stdout = &stdout
	shell_command.Stderr = &stderr
	err := shell_command.Run()
	output := sanitize.ReplaceAllNewlines(stdout.String())
	logs := sanitize.ReplaceAllNewlines(stderr.String())
	if err != nil {
		return output, logs, &rgerror.RGerror{
			Kind:    rgerror.ShellError,
			Message: fmt.Sprintf("Command '%s' failed:\n%s\nstderr:\n%s", shell_command, err, logs),
			Origin:  err,
		}
	}
	return output, logs, nil
}

func ExecScriptReadOutput(executable string, script string, args []string) (string, string, *rgerror.RGerror) {
	f, err := os.CreateTemp("", "regulator_script")
	if err != nil {
		return "", "", &rgerror.RGerror{
			Kind:    rgerror.ShellError,
			Message: "Could not create tmp file!",
			Origin:  err,
		}
	}
	filename := f.Name()
	defer os.Remove(filename) // clean up
	localfile.OverwriteFile(filename, []byte(script))
	final_args := append([]string{filename}, args...)
	return ExecReadOutput(executable, final_args)
}

func BuildAndRunCommand(executable string, file string, script string, args []string) (string, string, *rgerror.RGerror) {
	var output, logs string
	var airr *rgerror.RGerror
	if file == "" {
		output, logs, airr = ExecScriptReadOutput(executable, script, args)
	} else {
		final_args := append([]string{file}, args...)
		output, logs, airr = ExecReadOutput(executable, final_args)
	}
	if airr != nil {
		return output, logs, airr
	}

	return output, logs, nil
}
