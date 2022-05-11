package local

import (
	"fmt"
	"strings"

	"github.com/puppetlabs/regulator/localexec"
	"github.com/puppetlabs/regulator/localfile"
	"github.com/puppetlabs/regulator/operdefs"
	"github.com/puppetlabs/regulator/operparse"
	"github.com/puppetlabs/regulator/render"
	"github.com/puppetlabs/regulator/rgerror"
)

func RunObservation(name string, obsv operdefs.Observation, impls map[string]operdefs.Implement) operdefs.ObservationResult {
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
				return operdefs.ObservationResult{
					Succeeded:   true,
					Result:      "Error: " + strings.TrimSpace(cmd_arr.Message),
					Expected:    false,
					Logs:        logs,
					Observation: obsv,
				}
			} else {
				result := operdefs.ObservationResult{
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
	return operdefs.ObservationResult{
		Succeeded:   false,
		Result:      "Error: No implement found for observation '" + name + "'",
		Observation: obsv,
	}
}

func RunAllObservations(obsvs map[string]operdefs.Observation, impls map[string]operdefs.Implement) operdefs.ObservationResults {
	results := operdefs.ObservationResults{Observations: make(map[string]operdefs.ObservationResult)}
	for obsv_name, obsv := range obsvs {
		results.Observations[obsv_name] = RunObservation(obsv_name, obsv, impls)
	}
	return results
}

func Observe(raw_data []byte) (string, *rgerror.RGerror) {
	// No validators are required to run here because ParseRegulation
	// will use ReadFileOrStdin which performs validation on
	// maybe_file
	var data operdefs.Regulation
	parse_rgerr := operparse.ParseRegulation(raw_data, &data)
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
