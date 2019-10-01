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

// Takes a set of Go type strings and turns them into promiseConfigs. Types can
// be in a few different forms. These are listed below, with their resolved
// configs.
//
// - [*]goType: The config will create a type, goTypePromise, that carries the
//   given goType (as a pointer if * is included). No package is included.
// - [*]GoType: The config will create a type, GoTypePromise, that carries the
//   given goType (as a pointer if * is included). No package is included.
// - [*]package.GoType: The config will create a type, PackageGoTypePromise,
//   that carries the given package.GoType (as a pointer if * is included). The
//   package is resolved to a fully-qualified package by the Go import pipeline.
// - [*]fully.qualified.com/remote/package.GoType: The config will create a
//   type, PackageGoTypePromise, that carries the given package.GoType (as a
//   pointer if * is included). The package resolution is included as an
//   explicit import.
func parseTypesToConfig(goTypes []string) []promiseConfig {
	promiseConfigs := make([]promiseConfig, 0, len(goTypes))

	for _, goType := range goTypes {
		// Extract pointer asterisk, if specified.
		pointer := ""
		if goType[0] == '*' {
			pointer = "*"
			goType = goType[1:]
		}

		var qualifiedType, customPackage, typeName string

		// Try to split a <package>.<type> by extracting the post-package type
		// name. If the split works, assume we have a package name in the mix
		// and act accordingly. Otherwise, assume a simple type with no package
		// qualification.
		postPackageType := path.Ext(goType)
		if postPackageType != "" {
			// Strip leading dot to extract the type name.
			postPackageType = postPackageType[1:]

			// Drop the type to preserve just the package, then resolve the
			// unqualified package name as packageName. At this point we don't
			// yet know if the package name is fully qualified or not.
			pkg := goType[0 : len(goType)-len(postPackageType)-1]
			packageName := path.Base(pkg)

			// Construct the qualified type, including the pointer asterisk if
			// present.
			qualifiedType = pointer + packageName + "." + postPackageType
			customPackage = ""
			typeName = typeNameFrom(packageName, postPackageType)

			// If the package name and package are the same, this isn't a fully-
			// qualified package, so we'll be letting the compiler figure out
			// the import.
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

// Given a type package and name, construct a full type name that combines the
// package name and the type name into one long joint name.
//
// For example, given big.Int, we create BigInt, and given event.RelayRequest,
// we construct EventRelayRequest.
func typeNameFrom(typePackage string, typeName string) string {
	return strings.ToUpper(string(typePackage[0])) + typePackage[1:] + typeName
}

// Given the prefix of a promise's type name (e.g., BigInt for BigIntPromise, or
// EventRelayRequest for EventRelayRequestPromise), return an appropriate
// filename for the promise file.
//
// This is done by lowercasing all capital letters and following them with
// underscores, and appending a trailing `_promise.go`. Note that in the case of
// acronyms, this can have slightly odd effects (e.g., event_d_k_g_result_...
// for something like EventDKGResult...). Since the file generally won't be
// interacted with, this is an acceptable tradeoff for the algorithmic
// simplicity.
func filenameFrom(promisePrefix string) string {
	return strings.ToLower(promisePrefix[0:1]) + underscoreRegexp.ReplaceAllStringFunc(
		promisePrefix[1:],
		func(match string) string {
			return "_" + strings.ToLower(match)
		},
	) + "_promise.go"
}
