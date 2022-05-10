package language

import (
	"fmt"
	"sort"
	"strings"

	"github.com/puppetlabs/regulator/rgerror"
	"github.com/puppetlabs/regulator/sanitize"
	"gopkg.in/yaml.v2"
)

// Definitions of the language paradigms.
type Operation interface {
	// Operation conflicts are checked by building
	// a hash and checking if the hashed value of the
	// Operation already exists.
	//
	// Some operations have multiple hash keys, because
	// they can conflict in multiple ways
	HashKeys() []string
	Empty() bool
}

// OAR definitions
// Observations
// ---------------------------------------------------------------
type Observation struct {
	Entity   string `yaml:"entity" json:"entity"`
	Query    string `yaml:"query" json:"query"`
	Instance string `yaml:"instance" json:"instance"`
	Expect   string `yaml:"expect,omitempty" json:"expect,omitempty"`
}

type ObservationResult struct {
	Succeeded   bool        `json:"succeeded"`
	Result      string      `json:"result"`
	Expected    bool        `json:"expected"`
	Logs        string      `json:"logs"`
	Observation Observation `json:"observation"`
}

type ObservationResults struct {
	Observations map[string]ObservationResult `json:"observations"`
}

// Observations can only conflict if
// 1. they expect something
// 2. they share all fields with another observation
//    but expect something different
//
// This makes checking for conflicts a pain because
// in all other cases for the other operations we want
// to check things via hash collision but for observations
// it's okay if two observations collide as long as they
// don't expect two different things.
//
// What we end up having to do is special case the
// collision check for observations so that if we find
// a collision we go the extra step to see if the two
// observations expect different things.
func (obsv Observation) HashKeys() []string {
	result := []string{}
	// Don't even return a hash key if there is no expect field
	if obsv.Expect != "" {
		hash := "OBS" + "EN" + sanitize.ReplaceAllSpaces(obsv.Entity) +
			"QU" + sanitize.ReplaceAllSpaces(obsv.Query) +
			"IN" + sanitize.ReplaceAllSpaces(obsv.Instance) +
			"EX" + sanitize.ReplaceAllSpaces(obsv.Expect)
		result = append(result, hash)
	}
	return result
}

func (obsv Observation) Empty() bool {
	if obsv.Entity == "" ||
		obsv.Query == "" ||
		obsv.Instance == "" {
		return true
	}
	return false
}

// ---------------------------------------------------------------

// Actions
// ---------------------------------------------------------------
type Action struct {
	Path   string   `yaml:"path" json:"path"`
	Script string   `yaml:"script" json:"script"`
	Exe    string   `yaml:"exe,omitempty" json:"exe,omitempty"`
	Args   []string `yaml:"args,omitempty" json:"args,omitempty"`
}

type ActionResult struct {
	Succeeded bool   `json:"succeeded"`
	Output    string `json:"output"`
	Logs      string `json:"logs"`
	Action    Action `json:"action"`
}

type ActionResults struct {
	Actions map[string]ActionResult `json:"actions"`
}

func (actn Action) HashKeys() []string {
	// Actions can't conflict unless it's the name
	return []string{}
}

func (actn Action) Empty() bool {
	if actn.Path == "" && actn.Script == "" {
		return true
	}
	if actn.Exe == "" {
		return true
	}
	return false
}

// ---------------------------------------------------------------

// Reactions
// ---------------------------------------------------------------
type Condition struct {
	Check string      `yaml:"check" json:"check"`
	Value interface{} `yaml:"value" json:"check"`
}

type Reaction struct {
	Observation string    `yaml:"observation" json:"observation"`
	Action      string    `yaml:"action" json:"action"`
	Condition   Condition `yaml:"condition" json:"condition"`
}

type ReactionResult struct {
	Succeeded bool     `json:"succeeded"`
	Skipped   bool     `json:"skipped"`
	Output    string   `json:"output"`
	Logs      string   `json:"logs"`
	Message   string   `json:"message"`
	Reaction  Reaction `json:"reaction"`
}

type ReactionResults struct {
	Reactions    map[string]ReactionResult    `json:"reactions"`
	Observations map[string]ObservationResult `json:"observations"`
}

func (rctn Reaction) HashKeys() []string {
	// Reactions can't conflict unless it's the name
	return []string{}
}

func (rctn Reaction) Empty() bool {
	if rctn.Observation == "" ||
		rctn.Action == "" ||
		rctn.Condition.Check == "" ||
		rctn.Condition.Value == "" {
		return true
	}
	return false
}

// ---------------------------------------------------------------

// Implements
// ---------------------------------------------------------------
type Correction struct {
	Entity      string   `yaml:"entity"`
	Query       string   `yaml:"query"`
	Starts_From []string `yaml:"starts_from"`
	Results_In  string   `yaml:"results_in"`
}

type ReactionImplement struct {
	Corrects Correction `yaml:"corrects,omitempty"`
	Args     []string   `yaml:"args"`
}

type ObservationImplement struct {
	Entity string   `yaml:"entity"`
	Query  string   `yaml:"query"`
	Args   []string `yaml:"args"`
}

type Implement struct {
	Path     string               `yaml:"path,omitempty"`
	Script   string               `yaml:"script,omitempty"`
	Exe      string               `yaml:"exe"`
	Reacts   ReactionImplement    `yaml:"reacts,omitempty"`
	Observes ObservationImplement `yaml:"observes,omitempty"`
}

func emptyObserves(impl Implement) bool {
	if impl.Observes.Entity == "" ||
		impl.Observes.Query == "" ||
		impl.Observes.Args == nil {
		return true
	} else {
		return false
	}
}

func emptyReacts(impl Implement) bool {
	if impl.Reacts.Args == nil {
		return true
	} else {
		return false
	}
}

func emptyCorrects(impl Implement) bool {
	if impl.Reacts.Corrects.Entity == "" ||
		impl.Reacts.Corrects.Query == "" ||
		impl.Reacts.Corrects.Starts_From == nil ||
		impl.Reacts.Corrects.Results_In == "" {
		return true
	}
	return false
}

// Implements have very specific collision behavior:
//
// If the implement can react _and_ correct then it
// can conflict with other implements if they both
// attempt to correct the same thing. Implements
// that have the same value for all correction fields
// conflict.
//
// If the implement can observe then it conflicts with
// any other implement that observes the same entity/query
func (impl Implement) HashKeys() []string {
	result := []string{}
	if emptyReacts(impl) == false && emptyCorrects(impl) == false {
		// Sort the starts_from hash to ensure that the hash
		// we create can be checked against another hash
		// with the same values but in a different order
		var hash_starts string
		raw_starts := impl.Reacts.Corrects.Starts_From
		if raw_starts != nil {
			sort.Strings(raw_starts)
			hash_starts = strings.Join(raw_starts, "-")
		}
		react_hash := "IMPLRCT" + "EN" + sanitize.ReplaceAllSpaces(impl.Reacts.Corrects.Entity) +
			"QU" + sanitize.ReplaceAllSpaces(impl.Reacts.Corrects.Query) +
			"SF" + sanitize.ReplaceAllSpaces(hash_starts) +
			"RI" + sanitize.ReplaceAllSpaces(impl.Reacts.Corrects.Results_In)
		result = append(result, react_hash)
	}
	if emptyObserves(impl) == false {
		observe_hash := "IMPLOBS" + "EN" + sanitize.ReplaceAllSpaces(impl.Observes.Entity) +
			"QU" + sanitize.ReplaceAllSpaces(impl.Observes.Query)
		result = append(result, observe_hash)
	}
	return result
}

// I'm not positive this is true, but can't think
// of whether or not an implement can omit
// both reacting and observing (I'm pretty
// sure they are useless without that)
func (impl Implement) Empty() bool {
	if impl.Path == "" && impl.Script == "" {
		return true
	}
	if impl.Exe == "" {
		return true
	}
	if emptyReacts(impl) && emptyObserves(impl) {
		return true
	}
	return false
}

// ---------------------------------------------------------------

// Everything together
// ---------------------------------------------------------------
type Regulation struct {
	Reactions    map[string]Reaction    `yaml:"reactions,omitempty"`
	Observations map[string]Observation `yaml:"observations,omitempty"`
	Implements   map[string]Implement   `yaml:"implements,omitempty"`
	Actions      map[string]Action      `yaml:"actions,omitempty"`
}

// Idempotent function for merging new data in to Regulation
// struct. Can be used more than once to read data from multiple
// sources
func ParseRegulation(raw_data []byte, data *Regulation) *rgerror.RGerror {
	unmarshald_data := Regulation{}
	err := yaml.Unmarshal(raw_data, &unmarshald_data)
	if err != nil {
		return &rgerror.RGerror{
			Kind:    rgerror.ExecError,
			Message: fmt.Sprintf("Failed to parse yaml:\n%s", err),
			Origin:  err,
		}
	}
	rgerr := ConcatRegulation(data, &unmarshald_data)
	if rgerr != nil {
		return rgerr
	}
	return nil
}

// Yeah this is big and ugly and could probably have helper functions,
// but I don't want to do that much interface magic and pass enough
// strings around to make the messages different and helpful.
func ConcatRegulation(first *Regulation, second *Regulation) *rgerror.RGerror {
	var conflicts map[string]string = make(map[string]string)
	if first.Observations == nil {
		first.Observations = make(map[string]Observation)
	}
	if first.Reactions == nil {
		first.Reactions = make(map[string]Reaction)
	}
	if first.Actions == nil {
		first.Actions = make(map[string]Action)
	}
	if first.Implements == nil {
		first.Implements = make(map[string]Implement)
	}
	for obsv_name, obsv := range second.Observations {
		if _, conflicted := first.Observations[obsv_name]; conflicted == true {
			return &rgerror.RGerror{
				Kind:    rgerror.InvalidInput,
				Message: fmt.Sprintf("Duplicate observation name %s", obsv_name),
				Origin:  nil,
			}
		}
		for _, key := range obsv.HashKeys() {
			if conflict, conflicted := conflicts[key]; conflicted == true {
				// When observations have a collision that's not necessarily
				// a conflict, we have to check if the expect field is different.
				//
				// If the field _is_ different then there is a conflict, otherwise
				// it's fine. In the case where they are the same we don't need to
				// add this latest observation to the conflicts map because
				// there's already a matching hash there
				if first.Observations[conflict].Expect == obsv.Expect {
					return &rgerror.RGerror{
						Kind:    rgerror.InvalidInput,
						Message: fmt.Sprintf("Observation %s conflicts with %s", obsv_name, conflict),
						Origin:  nil,
					}
				}
			} else {
				conflicts[key] = obsv_name
			}
		}
		first.Observations[obsv_name] = obsv
	}
	for rctn_name, rctn := range second.Reactions {
		if _, conflicted := first.Reactions[rctn_name]; conflicted == true {
			return &rgerror.RGerror{
				Kind:    rgerror.InvalidInput,
				Message: fmt.Sprintf("Duplicate reaction name %s", rctn_name),
				Origin:  nil,
			}
		}
		for _, key := range rctn.HashKeys() {
			if conflict, conflicted := conflicts[key]; conflicted == true {
				return &rgerror.RGerror{
					Kind:    rgerror.InvalidInput,
					Message: fmt.Sprintf("Reaction %s conflicts with %s", rctn_name, conflict),
					Origin:  nil,
				}
			} else {
				conflicts[key] = rctn_name
			}
		}
		first.Reactions[rctn_name] = rctn
	}
	for actn_name, actn := range second.Actions {
		if _, conflicted := first.Actions[actn_name]; conflicted == true {
			return &rgerror.RGerror{
				Kind:    rgerror.InvalidInput,
				Message: fmt.Sprintf("Duplicate action name %s", actn_name),
				Origin:  nil,
			}
		}
		for _, key := range actn.HashKeys() {
			if conflict, conflicted := conflicts[key]; conflicted == true {
				return &rgerror.RGerror{
					Kind:    rgerror.InvalidInput,
					Message: fmt.Sprintf("Action %s conflicts with %s", actn_name, conflict),
					Origin:  nil,
				}
			} else {
				conflicts[key] = actn_name
			}
		}
		first.Actions[actn_name] = actn
	}
	// Ensure that the default impls are added first so that
	// any attempts to add an impl with the same name as a
	// default will always conflict.
	for default_impl_name, default_impl := range DEFAULT_IMPLS {
		// Don't even check for collisions or anything, just re-add
		// all the defaults every time.
		first.Implements[default_impl_name] = default_impl
	}
	for impl_name, impl := range second.Implements {
		if _, conflicted := first.Implements[impl_name]; conflicted == true {
			return &rgerror.RGerror{
				Kind:    rgerror.InvalidInput,
				Message: fmt.Sprintf("Duplicate implement name %s", impl_name),
				Origin:  nil,
			}
		}
		for _, key := range impl.HashKeys() {
			if conflict, conflicted := conflicts[key]; conflicted == true {
				return &rgerror.RGerror{
					Kind:    rgerror.InvalidInput,
					Message: fmt.Sprintf("Implement %s conflicts with %s", impl_name, conflict),
					Origin:  nil,
				}
			} else {
				conflicts[key] = impl_name
			}
		}
		first.Implements[impl_name] = impl
	}
	return nil
}

// Replaces a special string in a list of arguments (used for observations and
// reaction impls) with specific data from elsewhere
func ComputeArgs(arg_spec []string, obsv Observation) []string {
	var args []string
	for _, a := range arg_spec {
		switch a {
		case "instance":
			args = append(args, obsv.Instance)
		default:
			args = append(args, a)
		}
	}
	return args
}

func SelectAction(actn_name string, actns map[string]Action) *Action {
	if selected_action, found := actns[actn_name]; found {
		return &selected_action
	}
	return nil
}

func SelectObservation(obsv_name string, obsvs map[string]Observation) *Observation {
	if selected_obs, found := obsvs[obsv_name]; found {
		return &selected_obs
	}
	return nil
}

func SelectObservationResult(obsv_name string, obsv_results map[string]ObservationResult) *ObservationResult {
	if selected_obsv_result, found := obsv_results[obsv_name]; found {
		return &selected_obsv_result
	}
	return nil
}

func SelectImplementActionByName(impl_name string, impls map[string]Implement) *Action {
	if selected_impl, found := impls[impl_name]; found {
		return &Action{
			Path: selected_impl.Path,
			Exe:  selected_impl.Exe,
			Args: selected_impl.Reacts.Args,
		}
	}
	return nil
}

func SelectImplementActionForCorrection(obsv Observation, obsv_result ObservationResult, impls map[string]Implement) (string, *Action) {
	for impl_name, impl := range impls {
		if impl.Reacts.Corrects.Entity == obsv.Entity &&
			impl.Reacts.Corrects.Query == obsv.Query &&
			impl.Reacts.Corrects.Results_In == obsv.Expect {
			for _, state := range impl.Reacts.Corrects.Starts_From {
				if state == obsv_result.Result {
					return impl_name, &Action{
						Path: impl.Path,
						Exe:  impl.Exe,
						Args: impl.Reacts.Args,
					}
				}
			}
		}
	}
	return "", nil
}
