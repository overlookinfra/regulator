package utils

import (
	"fmt"
	"path/filepath"
	"regexp"

	. "github.com/puppetlabs/regulator/rgerror"
)

type ValidateType int

const (
	NotEmpty ValidateType = iota
	IsNumber
	IsFile
	IsIP
)

type Validator struct {
	Name     string
	Value    string
	Validate []ValidateType
}

func ValidateParams(params []Validator) *RGerror {
	for _, data := range params {
		for _, validate_type := range data.Validate {
			switch validate_type {
			case NotEmpty:
				if !(len(data.Value) > 0) {
					return &RGerror{
						InvalidInput,
						fmt.Sprintf("'%s' is empty", data.Name),
						nil,
					}
				}
			case IsNumber:
				matcher, _ := regexp.Compile(`^[\d]+$`)
				if !matcher.Match([]byte(data.Value)) {
					return &RGerror{
						InvalidInput,
						fmt.Sprintf("'%s' is not a number, given %s", data.Name, data.Value),
						nil,
					}
				}
			case IsIP:
				matcher, _ := regexp.Compile(`^[\d\.]+$`)
				if !matcher.Match([]byte(data.Value)) {
					return &RGerror{
						InvalidInput,
						fmt.Sprintf("'%s' is not an IP address, given %s", data.Name, data.Value),
						nil,
					}
				}
			case IsFile:
				files, err := filepath.Glob(data.Value)
				if err != nil {
					return &RGerror{
						InvalidInput,
						fmt.Sprintf("Failed attempting to check if '%s' is a file or directory, failure:\n%s", data.Name, err),
						nil,
					}
				}
				if len(files) < 1 {
					return &RGerror{
						InvalidInput,
						fmt.Sprintf("'%s' is not a file or directory, given %s", data.Name, data.Value),
						nil,
					}
				}
			}
		}
	}
	return nil
}