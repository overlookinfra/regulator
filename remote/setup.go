package remote

import (
	"fmt"
	"strings"

	"github.com/puppetlabs/regulator/connection"
	"github.com/puppetlabs/regulator/render"
	"github.com/puppetlabs/regulator/rgerror"
	"github.com/puppetlabs/regulator/validator"
	"github.com/puppetlabs/regulator/version"
)

func Setup(username string, target string, port string) (string, string, *rgerror.RGerror) {
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
		return "", "", rgerr
	}
	command := fmt.Sprintf(
		`#!/usr/bin/env bash

		mkdir -p $HOME/.regulator/bin 1>&2
		curl -L https://github.com/puppetlabs/regulator/releases/download/%s/regulator > $HOME/.regulator/bin/regulator
		chmod 755 $HOME/.regulator/bin/regulator 1>&2`,
		version.VERSION,
	)
	sout, serr, ec, rgerr := connection.RunSSHCommand(command, "", username, target, port)
	if rgerr != nil {
		return "", "", &rgerror.RGerror{
			Kind: rgerror.RemoteExecError,
			Message: fmt.Sprintf("regulator client on remote target returned non-zero exit code %d\n\nStdout:\n%s\nStderr:\n%s\n",
				ec,
				sout,
				serr),
			Origin: rgerr.Origin,
		}
	}
	return sout, serr, nil
}

func CLISetup(username string, target string, port string) *rgerror.RGerror {
	_, serr, rgerr := Setup(username, target, port)
	if rgerr != nil {
		return rgerr
	}
	output := make(map[string]interface{})
	output["ok"] = true
	output["logs"] = strings.TrimSpace(serr)
	final_result, json_rgerr := render.RenderJson(output)
	if json_rgerr != nil {
		return json_rgerr
	}
	fmt.Printf(final_result)
	return nil
}
