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

func ReactTo(rgln *operation.Regulation, obsv_results map[string]operation.ObservationResult) (*operation.ReactionResults, *rgerror.RGerror) {
	results := operation.ReactionResults{Reactions: make(map[string]operation.ReactionResult), Observations: obsv_results}
	for rctn_name, reaction := range rgln.Reactions {
		obsv_name := reaction.Observation
		obsv := operparse.SelectObservation(obsv_name, rgln.Observations)
		obsv_result := operparse.SelectObservationResult(obsv_name, obsv_results)
		if obsv == nil {
			results.Reactions[rctn_name] = operation.ReactionResult{
				Succeeded: false,
				Skipped:   true,
				Output:    "",
				Logs:      "",
				Message:   "Cannot react, '" + reaction.Observation + "' observation not found",
				Reaction:  reaction,
			}
			continue
		}
		if obsv_result.Succeeded == false {
			results.Reactions[rctn_name] = operation.ReactionResult{
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
					results.Reactions[rctn_name] = operation.ReactionResult{
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
					reaction_result := runReaction(
						obsv_result.Expected == false,
						reaction,
						actn_name,
						actn,
						"Skipped reaction: observation was the expected result",
					)
					results.Reactions[rctn_name] = reaction_result
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
					results.Reactions[rctn_name] = operation.ReactionResult{
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
						reaction_result := runReaction(
							obsv_result.Result == reaction.Condition.Value,
							reaction,
							reaction.Action,
							actn,
							"Skipped reaction: observation output did not match",
						)
						results.Reactions[rctn_name] = reaction_result
					case "expected":
						skip_msg := ""
						if reaction.Condition.Value == true {
							skip_msg = "Skipped reaction: observation was the expected result"
						} else {
							skip_msg = "Skipped reaction: observation was not the expected result"
						}
						reaction_result := runReaction(
							reaction.Condition.Value == obsv_result.Expected,
							reaction,
							reaction.Action,
							actn,
							skip_msg,
						)
						results.Reactions[rctn_name] = reaction_result
					default:
						results.Reactions[rctn_name] = operation.ReactionResult{
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
	return &results, nil
}

func React(raw_data []byte) (string, *rgerror.RGerror) {
	var data operation.Regulation
	parse_arr := operparse.ParseRegulation(raw_data, &data)
	if parse_arr != nil {
		return "", parse_arr
	}

	obsv_results := RunAllObservations(data.Observations, data.Implements).Observations
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
