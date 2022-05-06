package remote

import (
	"fmt"
	"strings"

	"github.com/puppetlabs/regulator/connection"
	. "github.com/puppetlabs/regulator/rgerror"
	"github.com/puppetlabs/regulator/utils"
	"github.com/puppetlabs/regulator/version"
)

func Setup(username string, target string, port string) (string, string, *RGerror) {

	command := fmt.Sprintf(
		`#!/usr/bin/env bash

		mkdir -p $HOME/.regulator/bin 1>&2
		curl -L https://github.com/puppetlabs/regulator/releases/download/%s/regulator > $HOME/.regulator/bin/regulator
		chmod 755 $HOME/.regulator/bin/regulator 1>&2`,
		version.VERSION,
	)
	sout, serr, ec, arr := connection.RunSSHCommand(command, "", username, target, port)
	if arr != nil {
		return "", "", &RGerror{
			Kind: RemoteExecError,
			Message: fmt.Sprintf("regulator client on remote target returned non-zero exit code %d\n\nStdout:\n%s\nStderr:\n%s\n",
				ec,
				sout,
				serr),
			Origin: arr.Origin,
		}
	}
	return sout, serr, nil
}

func CLISetup(username string, target string, port string) *RGerror {
	_, serr, airr := Setup(username, target, port)
	if airr != nil {
		return airr
	}
	output := make(map[string]interface{})
	output["ok"] = true
	output["logs"] = strings.TrimSpace(serr)
	final_result, json_rgerr := utils.RenderJson(output)
	if json_rgerr != nil {
		return json_rgerr
	}
	fmt.Printf(final_result)
	return nil
}
