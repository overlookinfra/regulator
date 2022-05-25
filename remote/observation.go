package remote

import (
	"fmt"

	"github.com/puppetlabs/regulator/connection"
	"github.com/puppetlabs/regulator/localfile"
	"github.com/puppetlabs/regulator/rgerror"
	"github.com/puppetlabs/regulator/validator"
)

func Observe(raw_data []byte, username string, target string, port string) (string, *rgerror.RGerror) {
	rgerr := validator.ValidateParams(fmt.Sprintf(
		`[
			{"name":"username","value":"%s","validate":["NotEmpty"]},
			{"name":"target","value":"%s","validate":["NotEmpty"]},
			{"name":"port","value":"%s","validate":["NotEmpty","IsNumber"]}
		 ]`,
		username,
		target,
		port,
	))
	if rgerr != nil {
		return "", rgerr
	}
	sout, serr, ec, rgerr := connection.RunSSHCommand("$HOME/.regulator/bin/regulator observe local --stdin", string(raw_data), username, target, port)
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

func CLIObserve(maybe_file string, username string, target string, port string) *rgerror.RGerror {
	raw_data, rgerr := localfile.ReadFileOrStdin(maybe_file)
	if rgerr != nil {
		return rgerr
	}
	sout, rgerr := Observe(raw_data, username, target, port)
	if rgerr != nil {
		return rgerr
	}
	fmt.Printf("%s", sout)
	return nil
}
