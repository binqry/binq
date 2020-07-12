package cli

import (
	"fmt"
	"strings"
	"testing"

	"github.com/progrhyme/binq"
)

type testCaseRun struct {
	args   []string
	exit   int
	outStr string
	errStr string
	check  func(t *testing.T)
}

// Tests for all available commands using table-driven tests
// but no operation which affects filesystem
func TestRunAll(t *testing.T) {
	prog := "binq"
	testCases := buildTestRunAllCases(prog)

	// Run test cases
	for _, tt := range testCases {
		name := fmt.Sprintf("%d:%s", tt.exit, strings.Join(tt.args, "_"))
		t.Run(name, func(t *testing.T) { subtestRun(t, prog, tt) })
	}
}

func subtestRun(t *testing.T, prog string, tc testCaseRun) {
	out := &strings.Builder{}
	err := &strings.Builder{}
	cmd := NewCLI(out, err)
	exit := cmd.Run(append([]string{prog}, tc.args...))
	if exit != tc.exit {
		t.Errorf("[Exit] Got: %d, Expected: %d", exit, tc.exit)
	}
	if tc.outStr == "" {
		if out.String() != "" {
			t.Errorf("[Stdout] Got: %s, Expected: %s", out.String(), tc.outStr)
		}
	} else if !strings.Contains(out.String(), tc.outStr) {
		t.Errorf("[Stdout] Got: %s, Expected: %s", out.String(), tc.outStr)
	}
	if tc.errStr == "" {
		if err.String() != "" {
			t.Errorf("[Stderr] Got: %s, Expected: %s", err.String(), tc.errStr)
		}
	} else if !strings.Contains(err.String(), tc.errStr) {
		t.Errorf("[Stderr] Got: %s, Expected: %s", err.String(), tc.errStr)
	}
	if tc.check != nil {
		tc.check(t)
	}
}

type testCommandInfo struct {
	helpText string
}

func buildTestRunAllCases(prog string) []testCaseRun {
	commands := buildTestCommandInfoMap(prog)
	invalidFlg := "--no-such-option"
	flagError := fmt.Sprintf("Error! Parsing arguments failed. unknown flag: %s", invalidFlg)

	return []testCaseRun{
		// Without subcommand (but install)
		{args: []string{}, exit: exitNG, outStr: "", errStr: commands["root"].helpText},
		{args: []string{"--help"}, exit: exitOK, outStr: "", errStr: commands["root"].helpText},
		{
			args: []string{invalidFlg}, exit: exitNG, outStr: "",
			errStr: strings.Join([]string{flagError, commands["root"].helpText}, "\n"),
		},

		// version
		{args: []string{"version"}, exit: exitOK, outStr: fmt.Sprintf("Version: %s", binq.Version), errStr: ""},

		// install
		{args: []string{"install", "--help"}, exit: exitOK, outStr: "", errStr: commands["install"].helpText},
		{
			args: []string{"install", invalidFlg}, exit: exitNG, outStr: "",
			errStr: strings.Join([]string{flagError, commands["install"].helpText}, "\n"),
		},
		{
			args: []string{"install"}, exit: exitNG, outStr: "",
			errStr: strings.Join([]string{"Error! Target is not specified!", commands["install"].helpText}, "\n"),
		},

		// new
		{args: []string{"new", "--help"}, exit: exitOK, outStr: "", errStr: commands["new"].helpText},
		{args: []string{"new", invalidFlg}, exit: exitNG, outStr: "", errStr: flagError},
		{
			args: []string{"new"}, exit: exitNG, outStr: "",
			errStr: strings.Join([]string{"Error! URL Format is not specified", commands["new"].helpText}, "\n"),
		},

		// revise
		{args: []string{"revise", "--help"}, exit: exitOK, outStr: "", errStr: commands["revise"].helpText},
		{args: []string{"revise", invalidFlg}, exit: exitNG, outStr: "", errStr: flagError},
		{
			args: []string{"revise"}, exit: exitNG, outStr: "",
			errStr: strings.Join([]string{
				"Error! Both JSON file and VERSION must be specified",
				commands["revise"].helpText}, "\n"),
		},
		{
			args: []string{"revise", "no-such-file.json"}, exit: exitNG, outStr: "",
			errStr: strings.Join([]string{
				"Error! Both JSON file and VERSION must be specified",
				commands["revise"].helpText}, "\n"),
		},
		{
			args: []string{"revise", "no-such-file.json", "0.1"}, exit: exitNG, outStr: "", errStr: "Error! Can't read item file: ",
		},

		// verify
		{args: []string{"verify", "--help"}, exit: exitOK, outStr: "", errStr: commands["verify"].helpText},
		{args: []string{"verify", invalidFlg}, exit: exitNG, outStr: "", errStr: flagError},
		{
			args: []string{"verify"}, exit: exitNG, outStr: "",
			errStr: strings.Join([]string{
				"Error! Target is not specified",
				commands["verify"].helpText}, "\n"),
		},
		{
			args: []string{"verify", "no-such-file.json"}, exit: exitNG, outStr: "", errStr: "Error! Can't read item file: ",
		},

		// register
		{args: []string{"register", "--help"}, exit: exitOK, outStr: "", errStr: commands["register"].helpText},
		{args: []string{"register", invalidFlg}, exit: exitNG, outStr: "", errStr: flagError},
		{
			args: []string{"register"}, exit: exitNG, outStr: "",
			errStr: strings.Join([]string{
				"Error! Both PATH_OF_INDEX and PATH_OF_ITEM must be specified",
				commands["register"].helpText}, "\n"),
		},
		{
			args: []string{"register", "invalid-index-filename.json"}, exit: exitNG, outStr: "",
			errStr: strings.Join([]string{
				"Error! Both PATH_OF_INDEX and PATH_OF_ITEM must be specified",
				commands["register"].helpText}, "\n"),
		},
		{
			args: []string{"register", "invalid-index-filename.json", "no-such-file.json"},
			exit: exitNG, outStr: "", errStr: "Error! INDEX JSON filename must be \"index.json\".",
		},

		// modify
		{args: []string{"modify", "--help"}, exit: exitOK, outStr: "", errStr: commands["modify"].helpText},
		{args: []string{"modify", invalidFlg}, exit: exitNG, outStr: "", errStr: flagError},
		{
			args: []string{"modify"}, exit: exitNG, outStr: "",
			errStr: strings.Join([]string{
				"Error! Both PATH_OF_INDEX and NAME_OF_ITEM must be specified",
				commands["modify"].helpText}, "\n"),
		},
		{
			args: []string{"modify", "invalid-index-filename.json"}, exit: exitNG, outStr: "",
			errStr: strings.Join([]string{
				"Error! Both PATH_OF_INDEX and NAME_OF_ITEM must be specified",
				commands["modify"].helpText}, "\n"),
		},
		{
			args: []string{"modify", "invalid-index-filename.json", "no-such-item"},
			exit: exitNG, outStr: "", errStr: "Error! INDEX JSON filename must be \"index.json\".",
		},

		// deregister
		{args: []string{"deregister", "--help"}, exit: exitOK, outStr: "", errStr: commands["deregister"].helpText},
		{args: []string{"deregister", invalidFlg}, exit: exitNG, outStr: "", errStr: flagError},
		{
			args: []string{"deregister"}, exit: exitNG, outStr: "",
			errStr: strings.Join([]string{
				"Error! Both PATH_OF_INDEX and NAME_OF_ITEM must be specified",
				commands["deregister"].helpText}, "\n"),
		},
		{
			args: []string{"deregister", "invalid-index-filename.json"}, exit: exitNG, outStr: "",
			errStr: strings.Join([]string{
				"Error! Both PATH_OF_INDEX and NAME_OF_ITEM must be specified",
				commands["deregister"].helpText}, "\n"),
		},
		{
			args: []string{"deregister", "invalid-index-filename.json", "no-such-item"},
			exit: exitNG, outStr: "", errStr: "Error! INDEX JSON filename must be \"index.json\".",
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

	info["verify"] = testCommandInfo{fmt.Sprintf(`Summary:
  Download a specified version in %s Item JSON and Verify its checksum.
  And update the checksum when needed.

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
