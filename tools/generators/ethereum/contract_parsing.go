package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

// The extracted name + payability of methods from ABI JSON.
type methodPayableInfo struct {
	Name    string
	Payable bool
}

var (
	classNameRegexp *regexp.Regexp
	shortVarRegexp  *regexp.Regexp
)

func init() {
	var err error
	classNameRegexp, err = regexp.Compile("ImplV.*")
	if err != nil {
		panic(fmt.Sprintf(
			"Failed to compile class name regular expression: [%v].",
			"ImplV.*",
		))
	}

	shortVarRegexp, err = regexp.Compile("([A-Z])[^A-Z]*")
	if err != nil {
		panic(fmt.Sprintf(
			"Failed to compile class name regular expression: [%v].",
			"([A-Z])[^A-Z]*",
		))
	}
}

// The following structs are sent into the templates for compilation.
type contractInfo struct {
	EthereumConfigReader string
	Class                string
	AbiClass             string
	FullVar              string
	ShortVar             string
	DashedName           string
	ConstMethods         []methodInfo
	NonConstMethods      []methodInfo
	Events               []eventInfo
}

type paramInfo struct {
	Name      string
	Type      string
	ParsingFn string
}

type methodInfo struct {
	CapsName          string
	LowerName         string
	DashedName        string
	Modifiers         string
	Payable           bool
	CommandCallable   bool
	Params            string
	ParamDeclarations string
	ParamInfos        []paramInfo
	Return            returnInfo
}

type returnInfo struct {
	Multi        bool
	Type         string
	Declarations string
	Vars         string
}

type eventInfo struct {
	CapsName                  string
	LowerName                 string
	SubscriptionCapsName      string
	ShortVar                  string
	SubscriptionShortVar      string
	IndexedFilters            string
	ParamExtractors           string
	ParamDeclarations         string
	IndexedFilterExtractors   string
	IndexedFilterDeclarations string
	IndexedFilterFields       string
}

func buildContractInfo(
	configReader string,
	abiClassName string,
	abi *abi.ABI,
	payableInfo []methodPayableInfo,
) contractInfo {
	payableMethods := make(map[string]struct{})
	for _, methodPayableInfo := range payableInfo {
		if methodPayableInfo.Payable {
			normalizedName := camelCase(methodPayableInfo.Name)
			_, ok := payableMethods[normalizedName]
			for idx := 0; ok; idx++ {
				normalizedName = fmt.Sprintf("%s%d", normalizedName, idx)
				_, ok = payableMethods[normalizedName]
			}
			payableMethods[normalizedName] = struct{}{}
		}
	}

	goClassName := classNameRegexp.ReplaceAll([]byte(abiClassName), nil)
	shortVar := strings.ToLower(string(shortVarRegexp.ReplaceAll(
		[]byte(goClassName),
		[]byte("$1"),
	)))
	dashedName := strings.ToLower(string(shortVarRegexp.ReplaceAll(
		[]byte(lowercaseFirst(string(goClassName))),
		[]byte("-$0"),
	)))
	constMethods, nonConstMethods := buildMethodInfo(payableMethods, abi.Methods)
	events := buildEventInfo(shortVar, abi.Events)

	return contractInfo{
		configReader,
		string(goClassName),
		abiClassName,
		lowercaseFirst(string(goClassName)),
		string(shortVar),
		string(dashedName),
		constMethods,
		nonConstMethods,
		events,
	}
}

func buildMethodInfo(
	payableMethods map[string]struct{},
	methodsByName map[string]abi.Method,
) (constMethods []methodInfo, nonConstMethods []methodInfo) {
	nonConstMethods = make([]methodInfo, 0, len(methodsByName))
	constMethods = make([]methodInfo, 0, len(methodsByName))

	for name, method := range methodsByName {
		normalizedName := camelCase(name)
		dashedName := strings.ToLower(string(shortVarRegexp.ReplaceAll(
			[]byte(normalizedName),
			[]byte("-$0"),
		)))

		_, payable := payableMethods[normalizedName]
		commandCallable := true

		modifiers := make([]string, 0, 0)
		if payable {
			modifiers = append(modifiers, "payable")
		}
		if method.Const {
			modifiers = append(modifiers, "constant")
		}
		modifierString := strings.Join(modifiers, " ")
		if len(modifiers) > 0 {
			modifierString += " "
		}

		paramDeclarations := ""
		params := ""
		paramInfos := make([]paramInfo, 0, 0)

		for index, param := range method.Inputs {
			goType := param.Type.Type.String()
			paramName := param.Name
			if paramName == "" {
				paramName = fmt.Sprintf("arg%v", index)
			}

			paramDeclarations += fmt.Sprintf("%v %v,\n", paramName, goType)
			params += fmt.Sprintf("%v,\n", paramName)

			parsingFn := ""
			switch param.Type.String() {
			case "bytes":
				parsingFn = "hexutil.Decode"
			case "address":
				parsingFn = "ethutil.AddressFromHex"
			case "uint256":
				parsingFn = "hexutil.DecodeBig"
			default:
				commandCallable = false
			}
			paramInfos = append(
				paramInfos,
				paramInfo{
					Name:      paramName,
					Type:      param.Type.String(),
					ParsingFn: parsingFn,
				})
		}

		returned := returnInfo{}
		if len(method.Outputs) > 1 {
			returned.Multi = true
			returned.Type = strings.Replace(normalizedName, "get", "", 1)

			for _, output := range method.Outputs {
				goType := output.Type.Type.String()

				returned.Declarations += fmt.Sprintf(
					"\t%v %v\n",
					uppercaseFirst(output.Name),
					goType,
				)
				returned.Vars += fmt.Sprintf("%v,", output.Name)
			}
		} else if len(method.Outputs) == 0 {
			returned.Multi = false
		} else {
			returned.Multi = false
			returned.Type = method.Outputs[0].Type.Type.String()
			returned.Vars += "ret,"
		}

		info := methodInfo{
			uppercaseFirst(normalizedName),
			lowercaseFirst(normalizedName),
			dashedName,
			modifierString,
			payable,
			commandCallable,
			params,
			paramDeclarations,
			paramInfos,
			returned,
		}

		if method.Const {
			constMethods = append(constMethods, info)
		} else {
			nonConstMethods = append(nonConstMethods, info)
		}
	}

	return constMethods, nonConstMethods
}

func buildEventInfo(
	contractShortVar string,
	eventsByName map[string]abi.Event,
) []eventInfo {
	eventInfos := make([]eventInfo, 0, len(eventsByName))
	for name, event := range eventsByName {

		capsName := uppercaseFirst(name)
		lowerName := lowercaseFirst(name)
		subscriptionCapsName := uppercaseFirst(contractShortVar) +
			capsName +
			"Subscription"

		shortVar := strings.ToLower(string(shortVarRegexp.ReplaceAll(
			[]byte(name),
			[]byte("$1"),
		)))
		subscriptionShortVar := shortVar + "s"

		paramDeclarations := ""
		paramExtractors := ""
		indexedFilterExtractors := ""
		indexedFilterDeclarations := ""
		indexedFilterFields := ""
		indexedFilters := ""
		for _, param := range event.Inputs {
			upperParam := uppercaseFirst(param.Name)
			goType := param.Type.Type.String()

			paramDeclarations += fmt.Sprintf("%v %v,\n", upperParam, goType)
			paramExtractors += fmt.Sprintf("event.%v,\n", upperParam)
			if param.Indexed {
				indexedFilterExtractors += fmt.Sprintf("%v.%vFilter,\n", subscriptionShortVar, param.Name)
				indexedFilterDeclarations += fmt.Sprintf("%vFilter []%v,\n", param.Name, goType)
				indexedFilterFields += fmt.Sprintf("%vFilter []%v\n", param.Name, goType)
				indexedFilters += fmt.Sprintf("%vFilter,\n", param.Name)
			}
		}

		paramDeclarations += "blockNumber uint64,\n"
		paramExtractors += "event.Raw.BlockNumber,\n"

		eventInfos = append(eventInfos, eventInfo{
			capsName,
			lowerName,
			subscriptionCapsName,
			shortVar,
			subscriptionShortVar,
			indexedFilters,
			paramExtractors,
			paramDeclarations,
			indexedFilterExtractors,
			indexedFilterDeclarations,
			indexedFilterFields,
		})
	}

	return eventInfos
}

func uppercaseFirst(str string) string {
	if len(str) == 0 {
		return str
	}

	str = strings.TrimPrefix(str, "_")

	return strings.ToUpper(str[0:1]) + str[1:]
}

func lowercaseFirst(str string) string {
	if len(str) == 0 {
		return str
	}

	str = strings.TrimPrefix(str, "_")

	return strings.ToLower(str[0:1]) + str[1:]
}

func camelCase(input string) string {
	parts := strings.Split(input, "_")
	for i, s := range parts {
		if len(s) > 0 {
			parts[i] = strings.ToUpper(s[:1]) + s[1:]
		}
	}
	return lowercaseFirst(strings.Join(parts, ""))
}
