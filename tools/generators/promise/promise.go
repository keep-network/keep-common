//go:generate go run github.com/keep-network/keep-common/tools/generators/template promise.go.tmpl promise_template_content.go
// Code generation execution command requires the package to be set to `main`.
package main

import (
	"bytes"
	"fmt"
	"log"
	"path"
	"text/template"

	"github.com/keep-network/keep-common/pkg/generate"
)

// Promises code generator.
// Execute `go generate` command in current directory to generate Promises code.

// Directory to which generated code will be exported.
const outDir string = "./async"

// Configuration for the generator
type promiseConfig struct {
	// Type which promise will handle.
	Type string
	// Prefix for naming the promises.
	Prefix string
	// Name of the generated file.
	outputFile string
}

func main() {
	configs := []promiseConfig{
		// Promise for `*big.Int` type.
		// There is a test for this promise named `big_int_promise_test.go`.
		// We need a test to validate correctness of generated promises.
		{
			Type:       "*big.Int",
			Prefix:     "BigInt",
			outputFile: "big_int_promise.go",
		},
		{
			Type:       "*event.Entry",
			Prefix:     "RelayEntry",
			outputFile: "relay_entry_promise.go",
		},
		{
			Type:       "*event.GroupTicketSubmission",
			Prefix:     "GroupTicket",
			outputFile: "group_ticket_submission_promise.go",
		},
		{
			Type:       "*event.GroupRegistration",
			Prefix:     "GroupRegistration",
			outputFile: "group_registration_promise.go",
		},
		{
			Type:       "*event.Request",
			Prefix:     "RelayRequest",
			outputFile: "relay_entry_requested_promise.go",
		},
		{
			Type:       "*event.DKGResultSubmission",
			Prefix:     "DKGResultSubmission",
			outputFile: "dkg_result_submission_promise.go",
		},
	}

	if err := generatePromisesCode(configs); err != nil {
		log.Fatalf("promises generation failed [%v]", err)
	}
}

// Generates promises based on a given `promiseConfig`
func generatePromisesCode(promisesConfig []promiseConfig) error {
	promiseTemplate, err :=
		template.
			New("promise").
			Parse(promiseTemplateContent)
	if err != nil {
		return fmt.Errorf("template creation failed [%v]", err)
	}

	for _, promiseConfig := range promisesConfig {
		outputFilePath := path.Join(outDir, promiseConfig.outputFile)

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
