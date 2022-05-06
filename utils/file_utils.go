package utils

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	. "github.com/puppetlabs/regulator/rgerror"
	. "github.com/puppetlabs/regulator/validator"
)

const STDIN_IDENTIFIER string = "__STDIN__"

func readFromStdin() string {
	var builder strings.Builder
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		builder.WriteString(scanner.Text() + "\n")
	}
	return builder.String()
}

func ChooseFileOrStdin(specfile string, use_stdin bool) (string, *RGerror) {
	if use_stdin {
		if len(specfile) > 0 {
			return "", &RGerror{
				InvalidInput,
				fmt.Sprintf("Cannot specify both a file and to use stdin"),
				nil,
			}
		}
		return STDIN_IDENTIFIER, nil
	} else {
		// Validate that the thing is actually a file on disk before
		// going any further
		arr := ValidateParams(
			[]Validator{
				Validator{"specfile", specfile, []ValidateType{NotEmpty, IsFile}},
			})
		if arr != nil {
			return "", arr
		}
		return specfile, nil
	}
}

func ReadFileOrStdin(maybe_file string) ([]byte, *RGerror) {
	var raw_data []byte
	var airr *RGerror
	if maybe_file == STDIN_IDENTIFIER {
		raw_data = []byte(readFromStdin())
	} else {
		raw_data, airr = ReadFileInChunks(maybe_file)
		if airr != nil {
			return nil, airr
		}
	}
	return raw_data, nil
}

func ReadFileInChunks(location string) ([]byte, *RGerror) {
	f, err := os.OpenFile(location, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return nil, &RGerror{
			ExecError,
			fmt.Sprintf("Failed to open file:\n%s", err),
			err,
		}
	}
	defer f.Close()

	// Create a buffer, read 32 bytes at a time
	byte_buffer := make([]byte, 32)
	file_contents := make([]byte, 0)
	for {
		bytes_read, err := f.Read(byte_buffer)
		if bytes_read > 0 {
			file_contents = append(file_contents, byte_buffer[:bytes_read]...)
		}
		if err != nil {
			if err != io.EOF {
				return nil, &RGerror{
					ExecError,
					fmt.Sprintf("Failed to read file:\n%s", err),
					err,
				}
			} else {
				break
			}
		}
	}
	return file_contents, nil
}
