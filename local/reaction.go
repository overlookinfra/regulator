package local

import (
	"fmt"

	"github.com/puppetlabs/regulator/localfile"
	"github.com/puppetlabs/regulator/operation"
	"github.com/puppetlabs/regulator/operparse"
	"github.com/puppetlabs/regulator/render"
	"github.com/puppetlabs/regulator/rgerror"
)

func runReaction(check_result bool, rctn operation.Reaction, actn_name string, actn *operation.Action, skipped_message string) operation.ReactionResult {
	if check_result {
		action_result := RunAction(*actn)
		if !action_result.Succeeded {
			return operation.ReactionResult{
				Succeeded: false,
				Skipped:   false,
				Output:    action_result.Output,
				Logs:      action_result.Logs,
				Message:   "Error running '" + actn_name + "'",
				Reaction:  rctn,
			}
		} else {
			return operation.ReactionResult{
				Succeeded: true,
				Skipped:   false,
				Output:    action_result.Output,
				Logs:      action_result.Logs,
				Message:   "Successfully ran '" + actn_name + "'",
				Reaction:  rctn,
			}
		}
	} else {
		return operation.ReactionResult{
			Succeeded: true,
			Skipped:   true,
			Output:    "",
			Logs:      "",
			Message:   skipped_message,
			Reaction:  rctn,
		}
	}
}

func maybeRunReaction(reaction operation.Reaction, obsv *operation.Observation, obsv_result *operation.ObservationResult, rgln *operation.Operations) operation.ReactionResult {
	if obsv == nil {
		return operation.ReactionResult{
			Succeeded: false,
			Skipped:   true,
			Output:    "",
			Logs:      "",
			Message:   "Cannot react, '" + reaction.Observation + "' observation not found",
			Reaction:  reaction,
		}
	}
	if obsv_result.Succeeded == false {
		return operation.ReactionResult{
			Succeeded: false,
			Skipped:   true,
			Output:    obsv_result.Result,
			Logs:      obsv_result.Logs,
			Message:   "Cannot react, error running observation",
			Reaction:  reaction,
		}
	} else {
		var actn *operation.Action = nil
		if reaction.Action == "correction" {
			actn_name, actn := operparse.SelectImplementActionForCorrection(*obsv, *obsv_result, rgln.Implements)
			if actn == nil && obsv_result.Expected == false {
				return operation.ReactionResult{
					Succeeded: false,
					Skipped:   true,
					Output:    "",
					Logs:      "",
					Message: fmt.Sprintf(
						"Could not react, no correction found for Entity %s Query %s with result %s that can correct to expected result %s",
						obsv.Entity,
						obsv.Query,
						obsv_result.Result,
						obsv.Expect,
					),
					Reaction: reaction,
				}
			} else {
				if actn != nil {
					actn.Args = operparse.ComputeArgs(actn.Args, *obsv)
				}
				return runReaction(
					obsv_result.Expected == false,
					reaction,
					actn_name,
					actn,
					"Skipped reaction: observation was the expected result",
				)
			}
		} else {
			actn = operparse.SelectAction(reaction.Action, rgln.Actions)
			if actn == nil {
				actn = operparse.SelectImplementActionByName(reaction.Action, rgln.Implements)
				if actn != nil {
					actn.Args = operparse.ComputeArgs(actn.Args, *obsv)
				}
			}
			if actn == nil {
				return operation.ReactionResult{
					Succeeded: false,
					Skipped:   true,
					Output:    "",
					Logs:      "",
					Message:   "Could not react, '" + reaction.Action + "' action not found",
					Reaction:  reaction,
				}
			} else {
				switch reaction.Condition.Check {
				case "matches":
					return runReaction(
						obsv_result.Result == reaction.Condition.Value,
						reaction,
						reaction.Action,
						actn,
						"Skipped reaction: observation output did not match",
					)
				case "expected":
					skip_msg := ""
					if reaction.Condition.Value == true {
						skip_msg = "Skipped reaction: observation was the expected result"
					} else {
						skip_msg = "Skipped reaction: observation was not the expected result"
					}
					return runReaction(
						reaction.Condition.Value == obsv_result.Expected,
						reaction,
						reaction.Action,
						actn,
						skip_msg,
					)
				default:
					return operation.ReactionResult{
						Succeeded: false,
						Output:    "",
						Message:   "Error checking condition, unknown Check type '" + reaction.Condition.Check + "'",
						Reaction:  reaction,
					}
				}
			}
		}
	}
}

func ReactTo(rgln *operation.Operations, all_obsv_results operation.ObservationResults) (*operation.ReactionResults, *rgerror.RGerror) {
	obsv_results := all_obsv_results.Observations
	results := operation.ReactionResults{
		Reactions:               make(map[string]operation.ReactionResult),
		Observations:            obsv_results,
		Total_Observations:      all_obsv_results.Total_Observations,
		Failed_Observations:     all_obsv_results.Failed_Observations,
		Unexpected_Observations: all_obsv_results.Unexpected_Observations,
	}
	for rctn_name, reaction := range rgln.Reactions {
		obsv_name := reaction.Observation
		obsv := operparse.SelectObservation(obsv_name, rgln.Observations)
		obsv_result := operparse.SelectObservationResult(obsv_name, obsv_results)
		this_result := maybeRunReaction(reaction, obsv, obsv_result, rgln)
		results.Reactions[rctn_name] = this_result
		results.Total_Reactions++
		if this_result.Succeeded == false {
			results.Failed_Reactions++
		}
		if this_result.Skipped == true {
			results.Skipped_Reactions++
		}
	}
	return &results, nil
}

func React(raw_data []byte) (string, *rgerror.RGerror) {
	var data operation.Operations
	parse_rgerr := operparse.ParseOperations(raw_data, &data)
	if parse_rgerr != nil {
		return "", parse_rgerr
	}

	obsv_results := RunAllObservations(data.Observations, data.Implements)
	results, rgerr := ReactTo(&data, obsv_results)
	if rgerr != nil {
		return "", rgerr
	}
	final_result, parse_rgerr := render.RenderJson(results)
	if parse_rgerr != nil {
		return "", parse_rgerr
	}

	return final_result, nil
}

func CLIReact(maybe_file string) *rgerror.RGerror {
	// ReadFileOrStdin performs validation on maybe_file
	raw_data, rgerr := localfile.ReadFileOrStdin(maybe_file)
	if rgerr != nil {
		return rgerr
	}
	result, rgerr := React(raw_data)
	if rgerr != nil {
		return rgerr
	}
	fmt.Printf(result)
	return nil
}
