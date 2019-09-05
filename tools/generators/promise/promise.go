//go:generate go run github.com/keep-network/keep-common/tools/generators/template promise.go.tmpl promise_template_content.go
// Code generation execution command requires the package to be set to `main`.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/keep-network/keep-common/pkg/generate"
)

// Directory to which generated code will be exported by default.
const defaultOutDir string = "./async"

// promiseConfig is a configuration for the promise generator. It specifies the
// type that the promise will provide asynchronously and the corresponding prefix
// for the promise class.
type promiseConfig struct {
	Type          string // Type which promise will handle.
	CustomPackage string // Custom package for type, if needed.
	Prefix        string // Prefix for promise struct (struct is <Prefix>Promise).
	Filename      string // Filename for promise.
}

// generatorConfig has one container, promises, for a set of promiseConfigs.
type generatorConfig struct {
	Promises []promiseConfig
}

var doHelp = flag.Bool(
	"h",
	false,
	"Display command help.",
)

var generationDir = flag.String(
	"d",
	defaultOutDir,
	"The `directory` to emit generated files to.",
)

func main() {
	flag.Parse()

	if *doHelp {
		fmt.Printf(helpText(path.Base(os.Args[0])))
		os.Exit(0)
	}

	if flag.NArg() < 1 {
		errorAndExit("Please specify at least one Go type.")
	}

	parsedConfigs := parseTypesToConfig(flag.Args())
	if err := generatePromisesCode(*generationDir, parsedConfigs); err != nil {
		errorAndExit(fmt.Sprintf("promises generation failed [%v]\n", err))
	}
}

func errorAndExit(err string) {
	fmt.Fprintf(os.Stderr, err+"\n\n")
	fmt.Println(helpText(path.Base(os.Args[0])))

	os.Exit(1)
}

func helpText(programName string) string {
	builder := strings.Builder{}
	builder.WriteString(programName + " [-d <directory>] <go-type>+\n\n")

	defaultOut := flag.CommandLine.Output()
	flag.CommandLine.SetOutput(&builder)
	flag.CommandLine.PrintDefaults()
	flag.CommandLine.SetOutput(defaultOut)

	builder.WriteString(
		"  <go-type>\n" +
			"    \tOne or more Go types, possibly including fully-qualified\n" +
			"    \tpackage info. When package info is not included, it is\n" +
			"    \tauto-resolved by the Go compiler. This type is the type\n" +
			"    \tmanaged by the promise, so include the pointer * if desired.\n" +
			"\n" +
			"    \tExamples: *big.Int *github.com/my/org/pkg/util.MyType string\n",
	)

	return builder.String()
}

// Generates promises based on a given `promiseConfig`
func generatePromisesCode(generationDir string, promisesConfig []promiseConfig) error {
	promiseTemplate, err :=
		template.
			New("promise").
			Parse(promiseTemplateContent)
	if err != nil {
		return fmt.Errorf("template creation failed [%v]", err)
	}

	for _, promiseConfig := range promisesConfig {
		outputFile := promiseConfig.Filename
		outputFilePath := path.Join(generationDir, outputFile)

		// Generate promise code.
		buffer, err := generateCode(promiseTemplate, &promiseConfig, outputFilePath)
		if err != nil {
			return fmt.Errorf("promise generation failed [%v]", err)
		}

		// Save the promise code to a file.
		if err := generate.SaveBufferToFile(buffer, outputFilePath); err != nil {
			return fmt.Errorf("saving promise code to file failed [%v]", err)
		}
	}
	return nil
}

// Generates a code from template and configuration.
// Returns a buffered code.
func generateCode(codeTemplate *template.Template, config *promiseConfig, outputFilePath string) (*bytes.Buffer, error) {
	var buffer bytes.Buffer

	if err := codeTemplate.Execute(&buffer, config); err != nil {
		return nil, fmt.Errorf("generating code for type %s failed [%v]", config.Type, err)
	}

	if err := generate.OrganizeImports(&buffer, outputFilePath); err != nil {
		return nil, fmt.Errorf("%v; input:\n%s", err, buffer.String())
	}

	return &buffer, nil
}
