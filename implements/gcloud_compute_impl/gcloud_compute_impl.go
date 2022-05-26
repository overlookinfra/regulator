package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/puppetlabs/regulator/cli"
	"github.com/puppetlabs/regulator/localexec"
	"github.com/puppetlabs/regulator/rgerror"
)

func runGcloudInstanceList(gcloud_project string) ([]map[string]interface{}, *rgerror.RGerror) {
	output, logs, rgerr := localexec.ExecReadOutput("gcloud", []string{"compute", "instances", "list", "--format=json", "--project=" + gcloud_project})
	if rgerr != nil {
		return nil, rgerr
	}
	fmt.Fprint(os.Stderr, logs)
	var instance_data []map[string]interface{}
	json.Unmarshal([]byte(output), &instance_data)
	return instance_data, nil
}

func readRunningInstances(state string, gcloud_project string) *rgerror.RGerror {
	var instance_count int = 0
	instance_data, rgerr := runGcloudInstanceList(gcloud_project)
	if rgerr != nil {
		return rgerr
	}
	for _, instance := range instance_data {
		if instance["status"] == state {
			instance_count++
		}
	}
	fmt.Printf("%d", instance_count)
	return nil
}

func readInstanceNames(state string, gcloud_project string) *rgerror.RGerror {
	var instance_names []string
	instance_data, rgerr := runGcloudInstanceList(gcloud_project)
	if rgerr != nil {
		return rgerr
	}
	for _, instance := range instance_data {
		if instance["status"] == state {
			instance_names = append(instance_names, instance["name"].(string))
		}
	}
	sort.Strings(instance_names)
	fmt.Printf("%s", strings.Join(instance_names, ","))
	return nil
}

func main() {
	command_list := []cli.Command{
		{
			Verb: "count",
			Noun: "instances",
			ExecutionFn: func() {
				usage := "gcloud_compute_impl count instances [STATE] [GCLOUD_PROJECT]"
				description := "count the number of gcloud instances in STATE"
				cli.ShouldHaveArgs(4, usage, description, nil)
				cli.HandleCommandRGerror(
					readRunningInstances(os.Args[3], os.Args[4]),
					usage,
					description,
					nil,
				)
				os.Exit(0)
			},
		},
		{
			Verb: "list",
			Noun: "instances",
			ExecutionFn: func() {
				usage := "gcloud_compute_impl list instances [STATE] [GCLOUD_PROJECT]"
				description := "return instance names for any instances in STATE"
				cli.ShouldHaveArgs(4, usage, description, nil)
				cli.HandleCommandRGerror(
					readInstanceNames(os.Args[3], os.Args[4]),
					usage,
					description,
					nil,
				)
				os.Exit(0)
			},
		},
	}
	cli.RunCommand("gcloud_compute_impl", command_list)
}
