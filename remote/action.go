package remote

import (
	"fmt"

	"github.com/puppetlabs/regulator/connection"
	. "github.com/puppetlabs/regulator/rgerror"
	"github.com/puppetlabs/regulator/utils"
	. "github.com/puppetlabs/regulator/validator"
)

func Run(raw_data []byte, actn_name string, username string, target string, port string) (string, *RGerror) {
	arr := ValidateParams(
		[]Validator{
			{Name: "action_name", Value: actn_name, Validate: []ValidateType{NotEmpty}},
			{Name: "username", Value: username, Validate: []ValidateType{NotEmpty}},
			{Name: "target", Value: target, Validate: []ValidateType{NotEmpty}},
			{Name: "port", Value: port, Validate: []ValidateType{NotEmpty, IsNumber}},
		})
	if arr != nil {
		return "", arr
	}
	command := fmt.Sprintf("$HOME/.regulator/bin/regulator run local \"%s\" --stdin", actn_name)
	sout, serr, ec, arr := connection.RunSSHCommand(command, string(raw_data), username, target, port)
	if arr != nil {
		return sout, &RGerror{
			Kind: RemoteExecError,
			Message: fmt.Sprintf("regulator client on remote target returned non-zero exit code %d\n\nStdout:\n%s\nStderr:\n%s\n",
				ec,
				sout,
				serr),
			Origin: arr.Origin,
		}
	}
	return sout, nil
}

func CLIRun(maybe_file string, actn_name string, username string, target string, port string) *RGerror {
	raw_data, arr := utils.ReadFileOrStdin(maybe_file)
	if arr != nil {
		return arr
	}
	sout, airr := Run(raw_data, actn_name, username, target, port)
	if airr != nil {
		return airr
	}
	fmt.Printf("%s", sout)
	return nil
}
