package cli

import (
	"fmt"
	"strings"
	"testing"

	"github.com/progrhyme/binq"
)

type testRunAllTable struct {
	args   []string
	exit   int
	outStr string
	errStr string
}

// Tests for all available commands using table-driven tests
// but no operation which affects filesystem
func TestRunAll(t *testing.T) {
	prog := "binq"
	testCases := buildTestRunAllCases(prog)

	// Run test cases
	for _, tt := range testCases {
		name := fmt.Sprintf("%d:%s", tt.exit, strings.Join(tt.args, "_"))
		t.Run(name, func(t *testing.T) {
			out := &strings.Builder{}
			err := &strings.Builder{}
			cmd := NewCLI(out, err)
			exit := cmd.Run(append([]string{prog}, tt.args...))
			if exit != tt.exit {
				t.Errorf("[Exit] Got: %d, Expected: %d", exit, tt.exit)
			}
			if tt.outStr == "" {
				if out.String() != "" {
					t.Errorf("[Stdout] Got: %s, Expected: %s", out.String(), tt.outStr)
				}
			} else if !strings.Contains(out.String(), tt.outStr) {
				t.Errorf("[Stdout] Got: %s, Expected: %s", out.String(), tt.outStr)
			}
			if tt.errStr == "" {
				if err.String() != "" {
					t.Errorf("[Stderr] Got: %s, Expected: %s", err.String(), tt.errStr)
				}
			} else if !strings.Contains(err.String(), tt.errStr) {
				t.Errorf("[Stderr] Got: %s, Expected: %s", err.String(), tt.errStr)
			}
		})
	}
}

type testCommandInfo struct {
	helpText string
}

func buildTestRunAllCases(prog string) []testRunAllTable {
	commands := buildTestCommandInfoMap(prog)
	invalidFlg := "--no-such-option"
	flagError := fmt.Sprintf("Error! Parsing arguments failed. unknown flag: %s", invalidFlg)

	return []testRunAllTable{
		// Without subcommand (but install)
		{[]string{}, exitNG, "", commands["root"].helpText},
		{[]string{"--help"}, exitOK, "", commands["root"].helpText},
		{
			[]string{invalidFlg}, exitNG, "",
			strings.Join([]string{flagError, commands["root"].helpText}, "\n"),
		},

		// version
		{[]string{"version"}, exitOK, fmt.Sprintf("Version: %s", binq.Version), ""},

		// install
		{[]string{"install", "--help"}, exitOK, "", commands["install"].helpText},
		{
			[]string{"install", invalidFlg}, exitNG, "",
			strings.Join([]string{flagError, commands["install"].helpText}, "\n"),
		},
		{
			[]string{"install"}, exitNG, "",
			strings.Join([]string{"Error! Target is not specified!", commands["install"].helpText}, "\n"),
		},

		// new
		{[]string{"new", "--help"}, exitOK, "", commands["new"].helpText},
		{[]string{"new", invalidFlg}, exitNG, "", flagError},
		{
			[]string{"new"}, exitNG, "",
			strings.Join([]string{"Error! URL Format is not specified", commands["new"].helpText}, "\n"),
		},

		// revise
		{[]string{"revise", "--help"}, exitOK, "", commands["revise"].helpText},
		{[]string{"revise", invalidFlg}, exitNG, "", flagError},
		{
			[]string{"revise"}, exitNG, "",
			strings.Join([]string{
				"Error! Both JSON file and VERSION must be specified",
				commands["revise"].helpText}, "\n"),
		},
		{
			[]string{"revise", "no-such-file.json"}, exitNG, "",
			strings.Join([]string{
				"Error! Both JSON file and VERSION must be specified",
				commands["revise"].helpText}, "\n"),
		},
		{
			[]string{"revise", "no-such-file.json", "0.1"}, exitNG, "", "Error! Can't read item file: ",
		},

		// register
		{[]string{"register", "--help"}, exitOK, "", commands["register"].helpText},
		{[]string{"register", invalidFlg}, exitNG, "", flagError},
		{
			[]string{"register"}, exitNG, "",
			strings.Join([]string{
				"Error! Both PATH_OF_INDEX and PATH_OF_ITEM must be specified",
				commands["register"].helpText}, "\n"),
		},
		{
			[]string{"register", "invalid-index-filename.json"}, exitNG, "",
			strings.Join([]string{
				"Error! Both PATH_OF_INDEX and PATH_OF_ITEM must be specified",
				commands["register"].helpText}, "\n"),
		},
		{
			[]string{"register", "invalid-index-filename.json", "no-such-file.json"},
			exitNG, "", "Error! INDEX JSON filename must be \"index.json\".",
		},

		// modify
		{[]string{"modify", "--help"}, exitOK, "", commands["modify"].helpText},
		{[]string{"modify", invalidFlg}, exitNG, "", flagError},
		{
			[]string{"modify"}, exitNG, "",
			strings.Join([]string{
				"Error! Both PATH_OF_INDEX and NAME_OF_ITEM must be specified",
				commands["modify"].helpText}, "\n"),
		},
		{
			[]string{"modify", "invalid-index-filename.json"}, exitNG, "",
			strings.Join([]string{
				"Error! Both PATH_OF_INDEX and NAME_OF_ITEM must be specified",
				commands["modify"].helpText}, "\n"),
		},
		{
			[]string{"modify", "invalid-index-filename.json", "no-such-item"},
			exitNG, "", "Error! INDEX JSON filename must be \"index.json\".",
		},

		// deregister
		{[]string{"deregister", "--help"}, exitOK, "", commands["deregister"].helpText},
		{[]string{"deregister", invalidFlg}, exitNG, "", flagError},
		{
			[]string{"deregister"}, exitNG, "",
			strings.Join([]string{
				"Error! Both PATH_OF_INDEX and NAME_OF_ITEM must be specified",
				commands["deregister"].helpText}, "\n"),
		},
		{
			[]string{"deregister", "invalid-index-filename.json"}, exitNG, "",
			strings.Join([]string{
				"Error! Both PATH_OF_INDEX and NAME_OF_ITEM must be specified",
				commands["deregister"].helpText}, "\n"),
		},
		{
			[]string{"deregister", "invalid-index-filename.json", "no-such-item"},
			exitNG, "", "Error! INDEX JSON filename must be \"index.json\".",
		},
	}
}

func buildTestCommandInfoMap(prog string) map[string]testCommandInfo {
	info := make(map[string]testCommandInfo)
	info["root"] = testCommandInfo{fmt.Sprintf(`Summary:
  "%s" does download & extract binary or archive via HTTP; then locate executable files into target
  directory.

Usage:`, prog)}

	info["install"] = testCommandInfo{`Summary:
  Download & extract binary or archive via HTTP; then locate executable files into target directory.

Syntax:`}

	info["new"] = testCommandInfo{fmt.Sprintf(`Summary:
  Generate a template Item JSON for %s.

Usage:`, prog)}

	info["revise"] = testCommandInfo{fmt.Sprintf(`Summary:
  Revise a version in Item JSON for %s.

Usage:`, prog)}

	info["register"] = testCommandInfo{fmt.Sprintf(`Summary:
  Register or Update Item content on Local %s Index Dataset.

Usage:`, prog)}

	info["modify"] = testCommandInfo{fmt.Sprintf(`Summary:
  Modify the indice properties of an Item in Local %s Index Dataset.

Usage:`, prog)}

	info["deregister"] = testCommandInfo{fmt.Sprintf(`Summary:
  Deregister an Item from Local %s Index Dataset.

Usage:`, prog)}
	return info
}
