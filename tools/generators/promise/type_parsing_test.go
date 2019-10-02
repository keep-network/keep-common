package main

import (
	"fmt"
	"testing"
)

func TestTypeParsing(t *testing.T) {
	tests := map[string]struct {
		goType          string
		resultingConfig promiseConfig
	}{
		"simple uncapitalized type": {
			goType: "goType",
			resultingConfig: promiseConfig{
				Type:          "goType",
				CustomPackage: "",
				Prefix:        "goType",
				Filename:      "go_type_promise.go",
			},
		},
		"simple capitalized type": {
			goType: "GoType",
			resultingConfig: promiseConfig{
				Type:          "GoType",
				CustomPackage: "",
				Prefix:        "GoType",
				Filename:      "go_type_promise.go",
			},
		},
		"simple package uncapitalized type": {
			goType: "package.goType",
			resultingConfig: promiseConfig{
				Type:          "package.goType",
				CustomPackage: "",
				Prefix:        "PackageGoType",
				Filename:      "package_go_type_promise.go",
			},
		},
		"simple package capitalized type": {
			goType: "package.GoType",
			resultingConfig: promiseConfig{
				Type:          "package.GoType",
				CustomPackage: "",
				Prefix:        "PackageGoType",
				Filename:      "package_go_type_promise.go",
			},
		},
		"fully-qualified package uncapitalized type": {
			goType: "github.com/test/package/sub.goType",
			resultingConfig: promiseConfig{
				Type:          "sub.goType",
				CustomPackage: "github.com/test/package/sub",
				Prefix:        "SubGoType",
				Filename:      "sub_go_type_promise.go",
			},
		},
		"fully-qualified package capitalized type": {
			goType: "github.com/test/package/sub.GoType",
			resultingConfig: promiseConfig{
				Type:          "sub.GoType",
				CustomPackage: "github.com/test/package/sub",
				Prefix:        "SubGoType",
				Filename:      "sub_go_type_promise.go",
			},
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			configs := parseTypesToConfig([]string{test.goType})

			if len(configs) != 1 {
				t.Errorf(
					"\nexpected: [%v]\nactual:   [%v]",
					"1 config",
					fmt.Sprintf("%v configs", len(configs)),
				)
			}

			if test.resultingConfig != configs[0] {
				t.Errorf(
					"\nexpected: [%+v]\nactual:   [%+v]",
					test.resultingConfig,
					configs[0],
				)
			}
		})
	}
}

func TestMultipleTypeParsing(t *testing.T) {
	goTypes := []string{
		"goType",
		"GoType",
		"package.goType",
		"package.GoType",
		"github.com/test/package/sub.goType",
		"github.com/test/package/sub.GoType",
	}

	resultingConfigs := []promiseConfig{
		promiseConfig{
			Type:          "goType",
			CustomPackage: "",
			Prefix:        "goType",
			Filename:      "go_type_promise.go",
		},
		promiseConfig{
			Type:          "GoType",
			CustomPackage: "",
			Prefix:        "GoType",
			Filename:      "go_type_promise.go",
		},
		promiseConfig{
			Type:          "package.goType",
			CustomPackage: "",
			Prefix:        "PackageGoType",
			Filename:      "package_go_type_promise.go",
		},
		promiseConfig{
			Type:          "package.GoType",
			CustomPackage: "",
			Prefix:        "PackageGoType",
			Filename:      "package_go_type_promise.go",
		},
		promiseConfig{
			Type:          "sub.goType",
			CustomPackage: "github.com/test/package/sub",
			Prefix:        "SubGoType",
			Filename:      "sub_go_type_promise.go",
		},
		promiseConfig{
			Type:          "sub.GoType",
			CustomPackage: "github.com/test/package/sub",
			Prefix:        "SubGoType",
			Filename:      "sub_go_type_promise.go",
		},
	}

	configs := parseTypesToConfig(goTypes)

	if len(resultingConfigs) != len(configs) {
		t.Errorf(
			"\nexpected: [%v]\nactual:   [%v]",
			fmt.Sprintf("%v configs", len(resultingConfigs)),
			fmt.Sprintf("%v configs", len(configs)),
		)
	}

	for i, config := range configs {
		if resultingConfigs[i] != config {
			t.Errorf(
				"\nexpected: [%+v]\nactual:   [%+v]\nfull result: [%+v]",
				resultingConfigs[i],
				config,
				configs,
			)
		}
	}
}

func TestFilenameFrom(t *testing.T) {
	tests := map[string]struct {
		promisePrefix string
		filename      string
	}{
		"simple type": {
			promisePrefix: "GoType",
			filename:      "go_type_promise.go",
		},
		"acronym type": {
			promisePrefix: "GoDKGType",
			filename:      "go_d_k_g_type_promise.go",
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			generatedFilename := filenameFrom(test.promisePrefix)

			if test.filename != generatedFilename {
				t.Errorf(
					"\nexpected: [%+v]\nactual:   [%+v]",
					test.filename,
					generatedFilename,
				)
			}
		})
	}
}
