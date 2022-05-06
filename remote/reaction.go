package remote

import (
	"fmt"

	"github.com/mcdonaldseanp/regulator/connection"
	. "github.com/mcdonaldseanp/regulator/rgerror"
	"github.com/mcdonaldseanp/regulator/utils"
	. "github.com/mcdonaldseanp/regulator/validator"
)

func React(raw_data []byte, username string, target string, port string) (string, *RGerror) {
	arr := ValidateParams(
		[]Validator{
			{Name: "username", Value: username, Validate: []ValidateType{NotEmpty}},
			{Name: "target", Value: target, Validate: []ValidateType{NotEmpty}},
			{Name: "port", Value: port, Validate: []ValidateType{NotEmpty, IsNumber}},
		})
	if arr != nil {
		return "", arr
	}
	sout, serr, ec, arr := connection.RunSSHCommand("$HOME/.regulator/bin/regulator react local --stdin", string(raw_data), username, target, port)
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

func CLIReact(maybe_file string, username string, target string, port string) *RGerror {
	raw_data, arr := utils.ReadFileOrStdin(maybe_file)
	if arr != nil {
		return arr
	}
	sout, airr := React(raw_data, username, target, port)
	if airr != nil {
		return airr
	}
	fmt.Printf("%s", sout)
	return nil
}
