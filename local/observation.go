package local

import (
	"fmt"
	"strings"

	"github.com/mcdonaldseanp/regulator/language"
	. "github.com/mcdonaldseanp/regulator/rgerror"
	"github.com/mcdonaldseanp/regulator/utils"
)

func RunObservation(name string, obsv language.Observation, impls map[string]language.Implement) language.ObservationResult {
	entity := obsv.Entity
	query := obsv.Query

	for _, impl := range impls {
		if impl.Observes.Query == query && impl.Observes.Entity == entity {
			script := impl.Path
			executable := impl.Exe
			args := language.ComputeArgs(impl.Observes.Args, obsv)
			output, logs, cmd_arr := utils.BuildAndRunCommand(executable, script, args)
			if cmd_arr != nil {
				return language.ObservationResult{
					Succeeded:   true,
					Result:      "Error: " + strings.TrimSpace(cmd_arr.Message),
					Expected:    false,
					Logs:        logs,
					Observation: obsv,
				}
			} else {
				result := language.ObservationResult{
					Succeeded:   true,
					Result:      output,
					Logs:        logs,
					Observation: obsv,
				}
				if obsv.Expect == output || obsv.Expect == "" {
					result.Expected = true
				} else {
					result.Expected = false
				}
				return result
			}
		}
	}
	return language.ObservationResult{
		Succeeded:   false,
		Result:      "Error: No implement found for observation '" + name + "'",
		Observation: obsv,
	}
}

func RunAllObservations(obsvs map[string]language.Observation, impls map[string]language.Implement) language.ObservationResults {
	results := language.ObservationResults{Observations: make(map[string]language.ObservationResult)}
	for obsv_name, obsv := range obsvs {
		results.Observations[obsv_name] = RunObservation(obsv_name, obsv, impls)
	}
	return results
}

func Observe(raw_data []byte) (string, *RGerror) {
	// No validators are required to run here because ParseRegulation
	// will use utils.ReadFileOrStdin which performs validation on
	// maybe_file
	var data language.Regulation
	parse_rgerr := language.ParseRegulation(raw_data, &data)
	if parse_rgerr != nil {
		return "", parse_rgerr
	}
	results := RunAllObservations(data.Observations, data.Implements)
	final_result, parse_rgerr := utils.RenderJson(results)
	if parse_rgerr != nil {
		return "", parse_rgerr
	}

	return final_result, nil
}

func CLIObserve(maybe_file string) *RGerror {
	// utils.ReadFileOrStdin performs validation on maybe_file
	raw_data, rgerr := utils.ReadFileOrStdin(maybe_file)
	if rgerr != nil {
		return rgerr
	}
	result, rgerr := Observe(raw_data)
	if rgerr != nil {
		return rgerr
	}
	fmt.Printf(result)
	return nil
}
