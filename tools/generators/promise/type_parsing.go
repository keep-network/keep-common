package main

import (
	"fmt"
	"path"
	"regexp"
	"strings"
)

var underscoreRegexp *regexp.Regexp

func init() {
	regexp, err := regexp.Compile("[A-Z]")
	if err != nil {
		panic(fmt.Sprintf(
			"Failed to compile underscore regexp [A-Z]: [%v].",
			err,
		))
	}

	underscoreRegexp = regexp
}

func parseTypesToConfig(goTypes []string) []promiseConfig {
	promiseConfigs := make([]promiseConfig, 0, len(goTypes))

	for _, goType := range goTypes {
		// Extract pointer character, if specified.
		pointer := ""
		if goType[0] == '*' {
			pointer = "*"
			goType = goType[1:]
		}

		var qualifiedType, customPackage, typeName string

		// Split out a post-package-name object. If the split works, assume we
		// have a package name in the mix and act accordingly.
		postPackageType := path.Ext(goType)
		if postPackageType != "" {
			// Strip leading dot.
			postPackageType = postPackageType[1:]
			pkg := goType[0 : len(goType)-len(postPackageType)-1]
			packageName := path.Base(pkg)

			qualifiedType = pointer + packageName + "." + postPackageType
			customPackage = ""
			typeName = typeNameFrom(packageName, postPackageType)

			// If the package name and package are the same, this isn't a fully-
			// qualified package, and we're letting the compiler figure out the
			// import.
			if pkg != packageName {
				customPackage = pkg
			}
		} else {
			qualifiedType = pointer + goType
			customPackage = ""
			typeName = goType
		}

		filename := filenameFrom(typeName)

		promiseConfigs = append(promiseConfigs, promiseConfig{
			Type:          qualifiedType,
			CustomPackage: customPackage,
			Prefix:        typeName,
			Filename:      filename,
		})
	}

	return promiseConfigs
}

func typeNameFrom(typePackage string, typeName string) string {
	return strings.ToUpper(string(typePackage[0])) + typePackage[1:] + typeName
}

func filenameFrom(promisePrefix string) string {
	return strings.ToLower(promisePrefix[0:1]) + underscoreRegexp.ReplaceAllStringFunc(
		promisePrefix[1:],
		func(match string) string {
			return "_" + strings.ToLower(match)
		},
	) + "_promise.go"
}
