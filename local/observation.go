package local

import (
	"fmt"
	"strings"

	"github.com/puppetlabs/regulator/localexec"
	"github.com/puppetlabs/regulator/localfile"
	"github.com/puppetlabs/regulator/operation"
	"github.com/puppetlabs/regulator/operparse"
	"github.com/puppetlabs/regulator/render"
	"github.com/puppetlabs/regulator/rgerror"
)

func RunObservation(name string, obsv operation.Observation, impls map[string]operation.Implement) operation.ObservationResult {
	entity := obsv.Entity
	query := obsv.Query

	for _, impl := range impls {
		if impl.Observes.Query == query && impl.Observes.Entity == entity {
			impl_file := impl.Path
			impl_script := impl.Script
			executable := impl.Exe
			args := operparse.ComputeArgs(impl.Observes.Args, obsv)
			output, logs, cmd_arr := localexec.BuildAndRunCommand(executable, impl_file, impl_script, args)
			if cmd_arr != nil {
				return operation.ObservationResult{
					Succeeded:   false,
					Result:      "Error: " + strings.TrimSpace(cmd_arr.Message),
					Expected:    false,
					Logs:        logs,
					Observation: obsv,
				}
			} else {
				result := operation.ObservationResult{
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
	return operation.ObservationResult{
		Succeeded:   false,
		Result:      "Error: No implement found for observation '" + name + "'",
		Observation: obsv,
	}
}

func RunAllObservations(obsvs map[string]operation.Observation, impls map[string]operation.Implement) operation.ObservationResults {
	results := operation.ObservationResults{Observations: make(map[string]operation.ObservationResult)}
	for obsv_name, obsv := range obsvs {
		this_result := RunObservation(obsv_name, obsv, impls)
		results.Observations[obsv_name] = this_result
		results.Total_Observations++
		if this_result.Succeeded == false {
			results.Failed_Observations++
		}
		if this_result.Expected == false {
			results.Unexpected_Observations++
		}
	}
	return results
}

func Observe(raw_data []byte) (string, *rgerror.RGerror) {
	// No validators are required to run here because ParseOperations
	// will use ReadFileOrStdin which performs validation on
	// maybe_file
	var data operation.Operations
	parse_rgerr := operparse.ParseOperations(raw_data, &data)
	if parse_rgerr != nil {
		return "", parse_rgerr
	}
	results := RunAllObservations(data.Observations, data.Implements)
	final_result, parse_rgerr := render.RenderJson(results)
	if parse_rgerr != nil {
		return "", parse_rgerr
	}

	return final_result, nil
}

func CLIObserve(maybe_file string) *rgerror.RGerror {
	// ReadFileOrStdin performs validation on maybe_file
	raw_data, rgerr := localfile.ReadFileOrStdin(maybe_file)
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
