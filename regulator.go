package main

import (
	"flag"
	"os"

	"github.com/puppetlabs/regulator/cli"
	"github.com/puppetlabs/regulator/local"
	"github.com/puppetlabs/regulator/localfile"
	"github.com/puppetlabs/regulator/remote"
)

func main() {
	// Use flagsets from the https://pkg.go.dev/flag package
	// to define CLI flags
	//
	// None of the commands below should call .Parse on any
	// flagsets directly. cli.ShouldHaveArgs() will call .Parse
	// on the flagset if it is passed one.
	//
	// Things need to be parsed inside cli.ShouldHaveArgs so that
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
	command_list := []cli.Command{
		{
			Verb: "observe",
			Noun: "local",
			ExecutionFn: func() {
				usage := "regulator observe local [FLAGS]"
				description := "Run observation code on the local system and print out the resulting observations"
				cli.ShouldHaveArgs(2, usage, description, local_flag_set)
				input_file, rgerr := localfile.ChooseFileOrStdin(*local_input_file, *local_use_stdin)
				if rgerr != nil {
					cli.HandleCommandRGerror(rgerr, usage, description, local_flag_set)
				}
				cli.HandleCommandRGerror(
					local.CLIObserve(input_file),
					usage,
					description,
					local_flag_set,
				)
			},
		},
		{
			Verb: "observe",
			Noun: "remote",
			ExecutionFn: func() {
				usage := "regulator observe remote [TARGET] [FLAGS]"
				description := "Run observation on a target"
				cli.ShouldHaveArgs(3, usage, description, remote_flag_set)
				input_file, rgerr := localfile.ChooseFileOrStdin(*remote_input_file, *remote_use_stdin)
				if rgerr != nil {
					cli.HandleCommandRGerror(rgerr, usage, description, remote_flag_set)
				}
				cli.HandleCommandRGerror(
					remote.CLIObserve(input_file, *username, os.Args[3], *port),
					usage,
					description,
					remote_flag_set,
				)
			},
		},
		{
			Verb: "react",
			Noun: "local",
			ExecutionFn: func() {
				usage := "regulator react local [FLAGS]"
				description := "React to an observation on the local system"
				cli.ShouldHaveArgs(2, usage, description, local_flag_set)
				input_file, rgerr := localfile.ChooseFileOrStdin(*local_input_file, *local_use_stdin)
				if rgerr != nil {
					cli.HandleCommandRGerror(rgerr, usage, description, local_flag_set)
				}
				cli.HandleCommandRGerror(
					local.CLIReact(input_file),
					usage,
					description,
					local_flag_set,
				)
			},
		},
		{
			Verb: "react",
			Noun: "remote",
			ExecutionFn: func() {
				usage := "regulator react remote [TARGET] [FLAGS]"
				description := "React to an observation on a target"
				cli.ShouldHaveArgs(3, usage, description, remote_flag_set)
				input_file, rgerr := localfile.ChooseFileOrStdin(*remote_input_file, *remote_use_stdin)
				if rgerr != nil {
					cli.HandleCommandRGerror(rgerr, usage, description, remote_flag_set)
				}
				cli.HandleCommandRGerror(
					remote.CLIReact(input_file, *username, os.Args[3], *port),
					usage,
					description,
					remote_flag_set,
				)
			},
		},
		{
			Verb: "run",
			Noun: "local",
			ExecutionFn: func() {
				usage := "regulator run local [ACTION NAME] [FLAGS]"
				description := "Run an action on the local system"
				cli.ShouldHaveArgs(3, usage, description, local_flag_set)
				input_file, rgerr := localfile.ChooseFileOrStdin(*local_input_file, *local_use_stdin)
				if rgerr != nil {
					cli.HandleCommandRGerror(rgerr, usage, description, local_flag_set)
				}
				cli.HandleCommandRGerror(
					local.CLIRun(input_file, os.Args[3]),
					usage,
					description,
					local_flag_set,
				)
			},
		},
		{
			Verb: "run",
			Noun: "remote",
			ExecutionFn: func() {
				usage := "regulator run remote [ACTION NAME] [TARGET] [FLAGS]"
				description := "Run actions on a target"
				cli.ShouldHaveArgs(4, usage, description, remote_flag_set)
				input_file, rgerr := localfile.ChooseFileOrStdin(*remote_input_file, *remote_use_stdin)
				if rgerr != nil {
					cli.HandleCommandRGerror(rgerr, usage, description, remote_flag_set)
				}
				cli.HandleCommandRGerror(
					remote.CLIRun(input_file, os.Args[3], *username, os.Args[4], *port),
					usage,
					description,
					remote_flag_set,
				)
			},
		},
		{
			Verb: "setup",
			Noun: "remote",
			ExecutionFn: func() {
				usage := "regulator setup remote [TARGET] [FLAGS]"
				description := "Run actions on a target"
				cli.ShouldHaveArgs(3, usage, description, setup_flag_set)
				cli.HandleCommandRGerror(
					remote.CLISetup(*setup_username, os.Args[3], *setup_port),
					usage,
					description,
					setup_flag_set,
				)
			},
		},
	}

	cli.RunCommand("regulator", command_list)
}
