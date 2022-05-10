package local

import (
	"fmt"

	"github.com/puppetlabs/regulator/language"
	"github.com/puppetlabs/regulator/localexec"
	"github.com/puppetlabs/regulator/localfile"
	"github.com/puppetlabs/regulator/render"
	"github.com/puppetlabs/regulator/rgerror"
	"github.com/puppetlabs/regulator/validator"
)

func RunAction(actn language.Action) language.ActionResult {
	result := language.ActionResult{
		Action: actn,
	}
	output, logs, cmd_rgerr := localexec.BuildAndRunCommand(actn.Exe, actn.Path, actn.Script, actn.Args)
	if cmd_rgerr != nil {
		result.Succeeded = false
		result.Output = output
		result.Logs = fmt.Sprintf("Error: %s, Logs: %s", cmd_rgerr.Message, logs)
	} else {
		result.Succeeded = true
		result.Output = output
		result.Logs = logs
	}
	return result
}

func Run(raw_data []byte, actn_name string) (string, *rgerror.RGerror) {
	rgerr := validator.ValidateParams(fmt.Sprintf(
		`[{"name":"action name","value":"%s","validate":["NotEmpty"]}]`,
		actn_name,
	))
	if rgerr != nil {
		return "", rgerr
	}
	var data language.Regulation
	parse_rgerr := language.ParseRegulation(raw_data, &data)
	if parse_rgerr != nil {
		return "", parse_rgerr
	}
	actn := language.SelectAction(actn_name, data.Actions)
	if actn == nil {
		return "", &rgerror.RGerror{
			Kind:    rgerror.InvalidInput,
			Message: fmt.Sprintf("Name \"%s\" does not match any existing action names", actn_name),
			Origin:  nil,
		}
	}
	result := RunAction(*actn)
	raw_final_result := language.ActionResults{Actions: make(map[string]language.ActionResult)}
	raw_final_result.Actions[actn_name] = result
	// The result for actions (for now) is an actionresults set with one action
	// result in the actions field.
	final_result, parse_rgerr := render.RenderJson(raw_final_result)
	if parse_rgerr != nil {
		return "", parse_rgerr
	}
	return final_result, nil
}

func CLIRun(maybe_file string, actn_name string) *rgerror.RGerror {
	// ReadFileOrStdin performs validation on maybe_file
	raw_data, rgerr := localfile.ReadFileOrStdin(maybe_file)
	if rgerr != nil {
		return rgerr
	}
	result, rgerr := Run(raw_data, actn_name)
	if rgerr != nil {
		return rgerr
	}
	fmt.Print(result)
	return nil
}
