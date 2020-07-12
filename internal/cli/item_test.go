package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"github.com/google/go-cmp/cmp"
)

func TestItemAll(t *testing.T) {
	prog := "binq"

	tmpdir, err := ioutil.TempDir(os.TempDir(), "binq-test-item.*")
	if err != nil {
		t.Fatalf("Error! Failed to create tempdir. %v\n", err)
	}
	defer os.RemoveAll(tmpdir)

	props := getTestItemProperties(tmpdir)
	testCases := buildTestNewCases(props)

	testCases = append(testCases, buildTestReviseCases(props)...)

	for _, tt := range testCases {
		name := fmt.Sprintf("%d:%s", tt.exit, strings.Join(tt.args, "_"))
		t.Run(name, func(t *testing.T) { subtestRun(t, prog, tt) })
	}
}

func getTestItemProperties(outDir string) (props map[string]string) {
	return map[string]string{
		"miniFile":            filepath.Join(outDir, "minimal.json"),
		"fullFile":            filepath.Join(outDir, "complete.json"),
		"urlMini":             "https://example.com/minimal/download/",
		"urlFull":             "https://example.com/download/complete-{{.Version}}-{{.OS}}-{{.Arch}}{{.Ext}}",
		"verNewMini":          "0.0.1",
		"verNewFull":          "0.9.9",
		"verReviseMiniLatest": "0.2.0",
		"verReviseMini":       "0.1.0",
	}
}

type testItemParams struct {
	template, file, outs, errs string
	props                      map[string]string
}

func buildTestNewCases(stash map[string]string) (tests []testCaseRun) {
	paramsList := []testItemParams{
		{
			template: testNewMinimalJSONFormat,
			file:     stash["miniFile"],
			errs:     fmt.Sprintf("Written %s", stash["miniFile"]),
			props: map[string]string{
				"url":     stash["urlMini"],
				"version": stash["verNewMini"],
			},
		},
		{
			template: testNewCompleteJSONFormat,
			file:     stash["fullFile"],
			errs:     fmt.Sprintf("Written %s", stash["fullFile"]),
			props: map[string]string{
				"url":     stash["urlFull"],
				"version": stash["verNewFull"],
				"repKey1": "amd64",
				"repVal1": "x86_64",
				"extKey1": "default",
				"extVal1": ".tar.gz",
				"rfKey1":  "complete-{{.Version}}-{{.OS}}-{{.Arch}}",
				"rfVal1":  "complete",
			},
		},
	}
	for _, params := range paramsList {
		tests = append(tests, buildTestNewCaseWithParams(params)...)
	}
	return tests
}

func buildTestReviseCases(stash map[string]string) (tests []testCaseRun) {
	paramsList := []testItemParams{
		{
			template: testNewMinimalJSONFormat,
			file:     stash["miniFile"],
			outs:     "",
			errs:     "No change",
			props: map[string]string{
				"url":     stash["urlMini"],
				"version": stash["verNewMini"],
			},
		},
		{
			template: testRevisedJSONFormat,
			file:     stash["miniFile"],
			outs:     fmt.Sprintf("Updated %s", stash["miniFile"]),
			errs:     "",
			props: map[string]string{
				"url":        stash["urlMini"],
				"version":    stash["verReviseMiniLatest"],
				"oldVersion": stash["verNewMini"],
			},
		},
		{
			template: testRevisedRichJSONFormat,
			file:     stash["miniFile"],
			outs:     fmt.Sprintf("Updated %s", stash["miniFile"]),
			errs:     "",
			props: map[string]string{
				"url":           stash["urlMini"],
				"version":       stash["verReviseMini"],
				"latestVersion": stash["verReviseMiniLatest"],
				"oldVersion":    stash["verNewMini"],
				"sumKey1":       filepath.Base(stash["miniFile"]),
				"sumVal1":       "dummy",
				"urlArg":        "https://example.com/revised/download/",
				"repKey1":       "amd64",
				"repVal1":       "x86_64",
				"extKey1":       "default",
				"extVal1":       ".tar.gz",
				"rfKey1":        "complete-{{.Version}}-{{.OS}}-{{.Arch}}",
				"rfVal1":        "complete",
			},
		},
	}
	for _, params := range paramsList {
		tests = append(tests, buildTestReviseCaseWithParams(params))
	}
	return tests
}

func buildTestNewCaseWithParams(params testItemParams) (tests []testCaseRun) {
	kv := params.props
	args := []string{"new", kv["url"], "--version", kv["version"]}
	if params.props["repKey1"] != "" {
		replaceArg := fmt.Sprintf("%s:%s", kv["repKey1"], kv["repVal1"])
		args = append(args, []string{"--replace", replaceArg}...)
	}
	if params.props["extKey1"] != "" {
		extension := fmt.Sprintf("%s:%s", kv["extKey1"], kv["extVal1"])
		args = append(args, []string{"--ext", extension}...)
	}
	if params.props["rfKey1"] != "" {
		renameArg := fmt.Sprintf("%s:%s", kv["rfKey1"], kv["rfVal1"])
		args = append(args, []string{"--rename", renameArg}...)
	}

	t := template.Must(template.New("item.json").Delims("<<", ">>").Parse(params.template))
	render := &strings.Builder{}
	t.Execute(render, kv)
	wantJSON := render.String()

	tests = append(tests, testCaseRun{
		args:   args,
		exit:   exitOK,
		outStr: wantJSON,
		errStr: "",
	})

	tests = append(tests, testCaseRun{
		args:   append(args, []string{"--file", params.file}...),
		exit:   exitOK,
		outStr: "",
		errStr: params.errs,
		check: func(t *testing.T) {
			raw, err := ioutil.ReadFile(params.file)
			if err != nil {
				t.Errorf("Can't read output file: %s. Error: %v", params.file, err)
				return
			}
			if diff := cmp.Diff(wantJSON, string(raw)); diff != "" {
				t.Errorf("Output JSON file has mismatch (-want +got):\n%s", diff)
			}
		},
	})

	return tests
}

func buildTestReviseCaseWithParams(params testItemParams) (tc testCaseRun) {
	kv := params.props
	args := []string{"revise", params.file, "--version", kv["version"]}
	if params.props["sumKey1"] != "" {
		sumArg := fmt.Sprintf("%s:%s", kv["sumKey1"], kv["sumVal1"])
		args = append(args, []string{"--sum", sumArg}...)
	}
	if params.props["urlArg"] != "" {
		args = append(args, []string{"--url", params.props["urlArg"]}...)
	}
	if params.props["repKey1"] != "" {
		replaceArg := fmt.Sprintf("%s:%s", kv["repKey1"], kv["repVal1"])
		args = append(args, []string{"--replace", replaceArg}...)
	}
	if params.props["extKey1"] != "" {
		extension := fmt.Sprintf("%s:%s", kv["extKey1"], kv["extVal1"])
		args = append(args, []string{"--ext", extension}...)
	}
	if params.props["rfKey1"] != "" {
		renameArg := fmt.Sprintf("%s:%s", kv["rfKey1"], kv["rfVal1"])
		args = append(args, []string{"--rename", renameArg}...)
	}

	t := template.Must(template.New("item.json").Delims("<<", ">>").Parse(params.template))
	render := &strings.Builder{}
	t.Execute(render, kv)
	wantJSON := render.String()

	return testCaseRun{
		args:   args,
		exit:   exitOK,
		outStr: params.outs,
		errStr: params.errs,
		check: func(t *testing.T) {
			raw, err := ioutil.ReadFile(params.file)
			if err != nil {
				t.Errorf("Can't read output file: %s. Error: %v", params.file, err)
				return
			}
			if diff := cmp.Diff(wantJSON, string(raw)); diff != "" {
				t.Errorf("Output JSON file has mismatch (-want +got):\n%s", diff)
			}
		},
	}
}

const testNewMinimalJSONFormat = `{
  "meta": {
    "url-format": "<<.url>>"
  },
  "latest": {
    "version": "<<.version>>"
  },
  "versions": [
    {
      "version": "<<.version>>"
    }
  ]
}
`

const testNewCompleteJSONFormat = `{
  "meta": {
    "url-format": "<<.url>>",
    "replacements": {
      "<<.repKey1>>": "<<.repVal1>>"
    },
    "extension": {
      "<<.extKey1>>": "<<.extVal1>>"
    },
    "rename-files": {
      "<<.rfKey1>>": "<<.rfVal1>>"
    }
  },
  "latest": {
    "version": "<<.version>>"
  },
  "versions": [
    {
      "version": "<<.version>>"
    }
  ]
}
`

const testRevisedJSONFormat = `{
  "meta": {
    "url-format": "<<.url>>"
  },
  "latest": {
    "version": "<<.version>>"
  },
  "versions": [
    {
      "version": "<<.version>>"
    },
    {
      "version": "<<.oldVersion>>"
    }
  ]
}
`

const testRevisedRichJSONFormat = `{
  "meta": {
    "url-format": "<<.url>>"
  },
  "latest": {
    "version": "<<.latestVersion>>"
  },
  "versions": [
    {
      "version": "<<.latestVersion>>"
    },
    {
      "version": "<<.version>>",
      "checksums": [
        {
          "file": "<<.sumKey1>>",
          "sha256": "<<.sumVal1>>"
        }
      ],
      "url-format": "<<.urlArg>>",
      "replacements": {
        "<<.repKey1>>": "<<.repVal1>>"
      },
      "extension": {
        "<<.extKey1>>": "<<.extVal1>>"
      },
      "rename-files": {
        "<<.rfKey1>>": "<<.rfVal1>>"
      }
    },
    {
      "version": "<<.oldVersion>>"
    }
  ]
}
`
