package main

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"

	_ "unsafe"

	_ "github.com/ethereum/go-ethereum/accounts/abi/bind"
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

// bindStructTypeGo resolves Go bindings for structs. It links to a non-exported
// method of go-ethereum's bind package that is used for Go bindings generation.
//
//go:linkname bindStructTypeGo github.com/ethereum/go-ethereum/accounts/abi/bind.bindStructTypeGo
func bindStructTypeGo(kind abi.Type, structs map[string]struct{}) string

// bindStructTypeGo resolves Go bindings for topics. It links to a non-exported
// method of go-ethereum's bind package that is used for Go bindings generation.
//
//go:linkname bindTopicTypeGo github.com/ethereum/go-ethereum/accounts/abi/bind.bindTopicTypeGo
func bindTopicTypeGo(kind abi.Type, structs map[string]struct{}) string

//go:linkname structured github.com/ethereum/go-ethereum/accounts/abi/bind.structured
func structured(args abi.Arguments) bool

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
	HostChainModule  string
	ChainUtilPackage string
	Class            string
	AbiClass         string
	FullVar          string
	ShortVar         string
	DashedName       string
	ConstMethods     []methodInfo
	NonConstMethods  []methodInfo
	Events           []eventInfo
}

type cmdArgInfo struct {
	Name       string
	Type       string
	GoType     string
	ParsingFn  string
	Structured bool
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
	CmdArgInfos       []cmdArgInfo
	Return            returnInfo
}

type returnInfo struct {
	Multi bool
	// Methods can return multiple outputs. If all the outputs are named they
	// are combined into a struct.
	Structured   bool
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
	hostChainModule string,
	chainUtilPackage string,
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

	structs := make(map[string]struct{})
	constMethods, nonConstMethods := buildMethodInfo(payableMethods, abi.Methods, structs)
	events := buildEventInfo(shortVar, abi.Events, structs)

	return contractInfo{
		hostChainModule,
		chainUtilPackage,
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
	structs map[string]struct{},
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

		if method.StateMutability != "" {
			modifiers = append(modifiers, method.StateMutability)
		}

		// Legacy indicators generated by compiler before v0.6.0
		if payable {
			modifiers = append(modifiers, "payable")
		}
		if method.Constant {
			modifiers = append(modifiers, "constant")
		}

		modifierString := strings.Join(modifiers, " ")
		if len(modifiers) > 0 {
			modifierString += " "
		}

		paramDeclarations := ""
		params := ""
		cmdArgInfos := make([]cmdArgInfo, 0, 0)

		for index, param := range method.Inputs {
			goType := bindType(param.Type, structs)

			var paramName string
			if param.Name == "" {
				paramName = fmt.Sprintf("arg%d", index)
			} else {
				paramName = fmt.Sprintf("arg_%v", param.Name)
			}

			paramDeclarations += fmt.Sprintf("%v %v,\n", paramName, goType)
			params += fmt.Sprintf("%v,\n", paramName)

			// Build cmdArgInfos used for CLI code generator
			cmdParamName := paramName
			cmdParamStructured := param.Type.TupleType != nil
			cmdParsingFn := ""

			if cmdParamStructured {
				cmdParamName += "_json"
			} else {
			goTypeSwitch:
				switch goType {
				case "[]byte":
					cmdParsingFn = "hexutil.Decode(%s)"
				case "common.Address":
					cmdParsingFn = "chainutil.AddressFromHex(%s)"
				case "*big.Int":
					cmdParsingFn = "hexutil.DecodeBig(%s)"
				case "bool":
					cmdParsingFn = "strconv.ParseBool(%s)"
				default:
					intParts := regexp.MustCompile(`^(u|)int([0-9]*)$`).FindStringSubmatch(goType)
					if len(intParts) > 0 {
						switch intParts[2] {
						case "8", "16", "32", "64":
							var template string
							if intParts[1] == "u" {
								template = "decode.ParseUint[uint%s](%%s, %s)"
							} else {
								template = "decode.ParseInt[int%s](%%s, %s)"
							}

							cmdParsingFn = fmt.Sprintf(template, intParts[2], intParts[2])
							break goTypeSwitch
						}
					}

					// TODO: Add support for more types, i.a. slices, arrays.
					fmt.Printf(
						"WARNING: Unsupported param type for method %s:\n"+
							"  ABI Type: %s\n"+
							"  Go Type:  %s\n"+
							"  the method won't be callable with 'ethereum' command\n",
						name,
						param.Type,
						goType,
					)
					commandCallable = false
				}
			}

			cmdArgInfos = append(
				cmdArgInfos,
				cmdArgInfo{
					Name:       cmdParamName,
					Type:       param.Type.String(),
					GoType:     goType,
					ParsingFn:  cmdParsingFn,
					Structured: cmdParamStructured,
				})
		}

		returned := returnInfo{}
		if len(method.Outputs) > 1 {
			returned.Multi = true
			returned.Type = strings.Replace(normalizedName, "get", "", 1)

			for index, output := range method.Outputs {
				goType := bindType(output.Type, structs)

				returned.Declarations += fmt.Sprintf(
					"\t%v %v\n",
					uppercaseFirst(output.Name),
					goType,
				)

				returned.Structured = structured(method.Outputs)

				// For structured outputs return one variable.
				if returned.Structured {
					returned.Vars = "ret,"
					continue
				}

				var varName string
				if output.Name == "" {
					varName = fmt.Sprintf("ret%d", index)
				} else {
					varName = fmt.Sprintf("ret_%v", output.Name)
				}

				returned.Vars += fmt.Sprintf("%v,", varName)
			}
		} else if len(method.Outputs) == 0 {
			returned.Multi = false
		} else {
			returned.Multi = false
			returned.Type = bindType(method.Outputs[0].Type, structs)
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
			cmdArgInfos,
			returned,
		}

		if isMethodConstant(method) {
			constMethods = append(constMethods, info)
		} else {
			nonConstMethods = append(nonConstMethods, info)
		}
	}

	sort.Sort(methodInfoSlice(constMethods))
	sort.Sort(methodInfoSlice(nonConstMethods))

	return constMethods, nonConstMethods
}

func buildEventInfo(
	contractShortVar string,
	eventsByName map[string]abi.Event,
	structs map[string]struct{},
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
			goType := bindType(param.Type, structs)

			paramExtractors += fmt.Sprintf("event.%v,\n", upperParam)

			if param.Indexed {
				// For event's indexed parameter abigen uses dedicated type binding
				// for topic.
				paramDeclarations += fmt.Sprintf("%v %v,\n", upperParam, bindTopicType(param.Type, structs))

				indexedFilterExtractors += fmt.Sprintf("%v.%vFilter,\n", subscriptionShortVar, param.Name)
				indexedFilterDeclarations += fmt.Sprintf("%vFilter []%v,\n", param.Name, goType)
				indexedFilterFields += fmt.Sprintf("%vFilter []%v\n", param.Name, goType)
				indexedFilters += fmt.Sprintf("%vFilter,\n", param.Name)
			} else {
				paramDeclarations += fmt.Sprintf("%v %v,\n", upperParam, goType)
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

	sort.Sort(eventInfoSlice(eventInfos))

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

// For sorting purposes, we define the following interfaces on methodInfo and
// eventInfo slices.
type methodInfoSlice []methodInfo

func (mis methodInfoSlice) Len() int {
	return len(mis)
}
func (mis methodInfoSlice) Less(i, j int) bool {
	return mis[i].LowerName < mis[j].LowerName
}
func (mis methodInfoSlice) Swap(i, j int) {
	mis[i], mis[j] = mis[j], mis[i]
}

type eventInfoSlice []eventInfo

func (eis eventInfoSlice) Len() int {
	return len(eis)
}
func (eis eventInfoSlice) Less(i, j int) bool {
	return eis[i].LowerName < eis[j].LowerName
}
func (eis eventInfoSlice) Swap(i, j int) {
	eis[i], eis[j] = eis[j], eis[i]
}

// Verifies if a method should be considered as constant based on the modifier.
// Constants methods are `view` or `pure`. For compatibility with code generated
// by compilers before v0.6.0 verify also a legacy `Constant` identifier.
func isMethodConstant(method abi.Method) bool {
	return method.StateMutability == "view" ||
		method.StateMutability == "pure" ||
		method.Constant
}

// Converts solidity type to a Go type.
func bindType(kind abi.Type, structs map[string]struct{}) string {
	goType := bindStructTypeGo(kind, structs)

	// Bindings for structs are expected to be generated into the `abi` package
	// by the abigen command.
	if kind.T == abi.TupleTy {
		goType = "abi." + goType
	}

	return goType
}

// Converts solidity topic type to a Go type.
func bindTopicType(kind abi.Type, structs map[string]struct{}) string {
	goType := bindTopicTypeGo(kind, structs)

	// Bindings for structs are expected to be generated into the `abi` package
	// by the abigen command.
	if kind.T == abi.TupleTy {
		goType = "abi." + goType
	}

	return goType
}
