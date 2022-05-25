package remote

import (
	"fmt"

	"github.com/puppetlabs/regulator/connection"
	"github.com/puppetlabs/regulator/localfile"
	"github.com/puppetlabs/regulator/rgerror"
	"github.com/puppetlabs/regulator/validator"
)

func Run(raw_data []byte, actn_name string, username string, target string, port string) (string, *rgerror.RGerror) {
	rgerr := validator.ValidateParams(fmt.Sprintf(
		`[
			{"name":"action name","value":"%s","validate":["NotEmpty"]},
			{"name":"username","value":"%s","validate":["NotEmpty"]},
			{"name":"target","value":"%s","validate":["NotEmpty"]},
			{"name":"port","value":"%s","validate":["NotEmpty","IsNumber"]}
		 ]`,
		actn_name,
		username,
		target,
		port,
	))
	if rgerr != nil {
		return "", rgerr
	}
	command := fmt.Sprintf("$HOME/.regulator/bin/regulator run local \"%s\" --stdin", actn_name)
	sout, serr, ec, rgerr := connection.RunSSHCommand(command, string(raw_data), username, target, port)
	if rgerr != nil {
		return sout, &rgerror.RGerror{
			Kind: rgerror.RemoteExecError,
			Message: fmt.Sprintf("regulator client on remote target returned non-zero exit code %d\n\nStdout:\n%s\nStderr:\n%s\n",
				ec,
				sout,
				serr),
			Origin: rgerr.Origin,
		}
	}
	return sout, nil
}

func CLIRun(maybe_file string, actn_name string, username string, target string, port string) *rgerror.RGerror {
	raw_data, rgerr := localfile.ReadFileOrStdin(maybe_file)
	if rgerr != nil {
		return rgerr
	}
	sout, rgerr := Run(raw_data, actn_name, username, target, port)
	if rgerr != nil {
		return rgerr
	}
	fmt.Printf("%s", sout)
	return nil
}
