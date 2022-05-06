package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/puppetlabs/regulator/local"
	"github.com/puppetlabs/regulator/remote"
	. "github.com/puppetlabs/regulator/rgerror"
	"github.com/puppetlabs/regulator/utils"
	"github.com/puppetlabs/regulator/version"
)

type CLICommand struct {
	Verb        string
	Noun        string
	ExecutionFn func()
}

// shouldHaveArgs does two things:
// * validate that the number of args that aren't flags have been provided (i.e. the number of strings
//    after "regulator" that aren't flags)
// * parse the remaining flags
//
// If the wrong number of args is passed it prints helpful usage
func shouldHaveArgs(num_args int, usage string, description string, flagset *flag.FlagSet) {
	real_args := num_args + 1
	passed_fs := flagset != nil
	for _, arg := range os.Args {
		if arg == "-h" {
			fmt.Fprintf(os.Stderr, "Usage:\n  %s\n\nDescription:\n  %s\n\n", usage, description)
			if passed_fs {
				fmt.Fprintf(os.Stderr, "Available flags:\n")
				flagset.PrintDefaults()
			}
			os.Exit(0)
		}
	}
	if len(os.Args) < real_args {
		fmt.Fprintf(os.Stderr, "Error running command:\n\nInvalid input, not enough arguments.\n\nUsage:\n  %s\n\nDescription:\n  %s\n\n", usage, description)
		if passed_fs {
			fmt.Fprintf(os.Stderr, "Available flags:\n")
			flagset.PrintDefaults()
		}
		os.Exit(1)
	} else if len(os.Args) > real_args && passed_fs {
		flagset.Parse(os.Args[real_args:])
	}
}

// handleCommandRGerror catches InvalidInput RGerrors and prints usage
// if that was the error thrown. IF a different type of RGerror is thrown
// it just prints the error.
//
// If the command succeeds handleCommandRGerror exits the whole go process
// with code 0
func handleCommandRGerror(airr *RGerror, usage string, description string, flagset *flag.FlagSet) {
	if airr != nil {
		if airr.Kind == InvalidInput {
			fmt.Fprintf(os.Stderr, "%s\nUsage:\n  %s\n\nDescription:\n  %s\n\n", airr, usage, description)
			if flagset != nil {
				flagset.PrintDefaults()
			}
		} else {
			fmt.Fprintf(os.Stderr, "Error running command:\n\n%s\n", airr)
		}
		os.Exit(1)
	}
	os.Exit(0)
}

func main() {
	// Use flagsets from the https://pkg.go.dev/flag package
	// to define CLI flags
	//
	// None of the commands below should call .Parse on any
	// flagsets directly. shouldHaveArgs() will call .Parse
	// on the flagset if it is passed one.
	//
	// Things need to be parsed inside shouldHaveArgs so that
	// the flag package can ignore any required commands
	// before parsing
	local_flag_set := flag.NewFlagSet("local_options", flag.ExitOnError)
	local_input_file := local_flag_set.String("file", "", "Path to spec yaml file (must use one of --file or --stdin)")
	local_use_stdin := local_flag_set.Bool("stdin", false, "Read spec from stdin (must use one of --file or --stdin)")

	remote_flag_set := flag.NewFlagSet("remote_options", flag.ExitOnError)
	remote_input_file := remote_flag_set.String("file", "", "Path to spec yaml file (must use one of --file or --stdin)")
	remote_use_stdin := remote_flag_set.Bool("stdin", false, "Read spec from stdin (must use one of --file or --stdin)")
	username := remote_flag_set.String("user", os.Getenv("USER"), "Username to use when connecting via SSH")
	port := remote_flag_set.String("port", "22", "Port to use for ssh connections")

	setup_flag_set := flag.NewFlagSet("setup_options", flag.ExitOnError)
	setup_username := setup_flag_set.String("user", os.Getenv("USER"), "Username to use when connecting via SSH")
	setup_port := setup_flag_set.String("port", "22", "Port to use for ssh connections")

	// All CLI commands should follow naming rules of powershell approved verbs:
	// https://docs.microsoft.com/en-us/powershell/scripting/developer/cmdlet/approved-verbs-for-windows-powershell-commands?view=powershell-7.2
	//
	// Also, try to keep these in alphabetical order. The list is already long enough
	command_list := []CLICommand{
		{"observe", "local",
			func() {
				usage := "regulator observe local [FLAGS]"
				description := "Run observation code on the local system and print out the resulting observations"
				shouldHaveArgs(2, usage, description, local_flag_set)
				input_file, rgerr := utils.ChooseFileOrStdin(*local_input_file, *local_use_stdin)
				if rgerr != nil {
					handleCommandRGerror(rgerr, usage, description, local_flag_set)
				}
				handleCommandRGerror(
					local.CLIObserve(input_file),
					usage,
					description,
					local_flag_set,
				)
			},
		},
		{"observe", "remote",
			func() {
				usage := "regulator observe remote [TARGET] [FLAGS]"
				description := "Run observation on a target"
				shouldHaveArgs(3, usage, description, remote_flag_set)
				input_file, rgerr := utils.ChooseFileOrStdin(*remote_input_file, *remote_use_stdin)
				if rgerr != nil {
					handleCommandRGerror(rgerr, usage, description, remote_flag_set)
				}
				handleCommandRGerror(
					remote.CLIObserve(input_file, *username, os.Args[3], *port),
					usage,
					description,
					remote_flag_set,
				)
			},
		},
		{"react", "local",
			func() {
				usage := "regulator react local [FLAGS]"
				description := "React to an observation on the local system"
				shouldHaveArgs(2, usage, description, local_flag_set)
				input_file, rgerr := utils.ChooseFileOrStdin(*local_input_file, *local_use_stdin)
				if rgerr != nil {
					handleCommandRGerror(rgerr, usage, description, local_flag_set)
				}
				handleCommandRGerror(
					local.CLIReact(input_file),
					usage,
					description,
					local_flag_set,
				)
			},
		},
		{"react", "remote",
			func() {
				usage := "regulator react remote [TARGET] [FLAGS]"
				description := "React to an observation on a target"
				shouldHaveArgs(3, usage, description, remote_flag_set)
				input_file, rgerr := utils.ChooseFileOrStdin(*remote_input_file, *remote_use_stdin)
				if rgerr != nil {
					handleCommandRGerror(rgerr, usage, description, remote_flag_set)
				}
				handleCommandRGerror(
					remote.CLIReact(input_file, *username, os.Args[3], *port),
					usage,
					description,
					remote_flag_set,
				)
			},
		},
		{"run", "local",
			func() {
				usage := "regulator run local [ACTION NAME] [FLAGS]"
				description := "Run an action on the local system"
				shouldHaveArgs(3, usage, description, local_flag_set)
				input_file, rgerr := utils.ChooseFileOrStdin(*local_input_file, *local_use_stdin)
				if rgerr != nil {
					handleCommandRGerror(rgerr, usage, description, local_flag_set)
				}
				handleCommandRGerror(
					local.CLIRun(input_file, os.Args[3]),
					usage,
					description,
					local_flag_set,
				)
			},
		},
		{"run", "remote",
			func() {
				usage := "regulator run remote [ACTION NAME] [TARGET] [FLAGS]"
				description := "Run actions on a target"
				shouldHaveArgs(4, usage, description, remote_flag_set)
				input_file, rgerr := utils.ChooseFileOrStdin(*remote_input_file, *remote_use_stdin)
				if rgerr != nil {
					handleCommandRGerror(rgerr, usage, description, remote_flag_set)
				}
				handleCommandRGerror(
					remote.CLIRun(input_file, os.Args[3], *username, os.Args[4], *port),
					usage,
					description,
					remote_flag_set,
				)
			},
		},
		{"setup", "remote",
			func() {
				usage := "regulator setup remote [TARGET] [FLAGS]"
				description := "Run actions on a target"
				shouldHaveArgs(3, usage, description, setup_flag_set)
				handleCommandRGerror(
					remote.CLISetup(*setup_username, os.Args[3], *setup_port),
					usage,
					description,
					setup_flag_set,
				)
			},
		},
	}

	if len(os.Args) > 2 {
		for _, command := range command_list {
			if os.Args[1] == command.Verb && os.Args[2] == command.Noun {
				command.ExecutionFn()
			}
		}
	}

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version":
			fmt.Fprintf(os.Stdout, "%s\n", version.VERSION)
			os.Exit(0)
		case "-h":
			// do nothing, it will print the usage message below
		default:
			// If we've arrived here, that means the args passed don't match an existing command
			// --version or -h
			fmt.Printf("Unknown regulator command \"%s\"\n\n", strings.Join(os.Args, " "))
		}
	}

	fmt.Printf("Usage:\n  regulator [COMMAND] [OBJECT] [ARGUMENTS] [FLAGS]\n\nAvailable commands:\n")
	for _, command := range command_list {
		fmt.Printf("    %s %s\n", command.Verb, command.Noun)
	}
	os.Exit(1)
}
