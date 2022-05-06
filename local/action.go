package local

import (
	"fmt"

	"github.com/puppetlabs/regulator/language"
	. "github.com/puppetlabs/regulator/rgerror"
	"github.com/puppetlabs/regulator/utils"
	. "github.com/puppetlabs/regulator/validator"
)

func RunAction(actn language.Action) language.ActionResult {
	result := language.ActionResult{
		Action: actn,
	}
	output, logs, cmd_rgerr := utils.BuildAndRunCommand(actn.Exe, actn.Path, actn.Args)
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

func Run(raw_data []byte, actn_name string) (string, *RGerror) {
	rgerr := ValidateParams(
		[]Validator{
			{Name: "action_name", Value: actn_name, Validate: []ValidateType{NotEmpty}},
		})
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
		return "", &RGerror{
			Kind:    InvalidInput,
			Message: fmt.Sprintf("Name \"%s\" does not match any existing action names", actn_name),
			Origin:  nil,
		}
	}
	result := RunAction(*actn)
	raw_final_result := language.ActionResults{Actions: make(map[string]language.ActionResult)}
	raw_final_result.Actions[actn_name] = result
	// The result for actions (for now) is an actionresults set with one action
	// result in the actions field.
	final_result, parse_rgerr := utils.RenderJson(raw_final_result)
	if parse_rgerr != nil {
		return "", parse_rgerr
	}
	return final_result, nil
}

func CLIRun(maybe_file string, actn_name string) *RGerror {
	// utils.ReadFileOrStdin performs validation on maybe_file
	raw_data, rgerr := utils.ReadFileOrStdin(maybe_file)
	if rgerr != nil {
		return rgerr
	}
	result, rgerr := Run(raw_data, actn_name)
	if rgerr != nil {
		return rgerr
	}
	fmt.Printf(result)
	return nil
}
