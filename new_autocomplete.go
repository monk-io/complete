package complete

import (
	"strings"

	"github.com/google/shlex"
)

// a tree representation of CLI commands & sub commands
type CompTree struct {
	Flags map[string]*CompTree
	// FIXME: Args & Sub may be able to be combined since thus far there doesn't seem to necessitate a difference
	Args       map[string]*CompTree
	Sub        map[string]*CompTree
	Dynamic    func(prefix string) []Suggestion
	Desc       string
	Name       string
	TakesValue bool
}

type SearchMethod func(s, query string) bool

var (
	SuggestSomething = []Suggestion{{Name: ""}}
	SuggestNothing   []Suggestion
	SuggestFlag      = []Suggestion{{Name: "--"}}
)

type Suggestion struct {
	Name string
	Desc string
}

//
//monk > cluster|user|load etc
//monk - > -- >  flags
//monk -- > flags
//monk cluster > info|peers etc
//monk cluster peers > --nocolor etc
//monk run > list of runnables
//
//
// Псевдокод
//* prev:
//	* флаг
//		* есть predictor -> execute и return
//		* takes values -> return nothing
//		* next
//	* команда
//		* search subcommands
//		* predictor?
//		* -- > флаги
//* - or --  - флаги
//* flag + = || flag + " " - > flag value
//* subcommand

//
// задачи:
//выполнять динамик функции
//считать - так же как и --

func AutoComplete(text string, completionTree *CompTree, queryFunc SearchMethod) ([]Suggestion, error) {
	fields, err := shlex.Split(strings.TrimSpace(text))
	if err != nil {
		return nil, err
	}
	suggestions := autoComplete(completionTree, fields, text, queryFunc)
	return suggestions, nil
}

func autoComplete(completionTree *CompTree, fields []string, text string, queryFunc SearchMethod) []Suggestion {
	// current (last) command (not flag)
	curr := completionTree
	// global flags
	globalFlags := completionTree.Flags

	last := fields[len(fields)-1]
	// i.e. in "git checkout --fie", remove "git"
	fields = fields[1:]
	for i, field := range fields {
		if strings.HasPrefix(field, "-") {
			continue
		}
		if i == len(fields)-1 && !strings.HasSuffix(text, " ") {
			continue
		}
		if curr.Sub[field] != nil {
			curr = curr.Sub[field]
		}
	}
	// get suggestions
	s := make([]Suggestion, 0, 20)
	name := last
	switch {
	case curr == nil:
		return s
	case strings.HasSuffix(text, " "):
		// complete command argument | flag value | subcommands | flags
		return completeSpace(curr, globalFlags, name)
	case strings.HasPrefix(last, "-") && strings.HasSuffix(last, "="):
		// complete flag value
		return completeFlagValue(curr, globalFlags, name)
	case strings.HasPrefix(last, "-"):
		// complete flag
		return completeFlags(curr, globalFlags, name, fields)
	default:
		// complete partial ("search") argument
		return searchArgs(curr, queryFunc, last)
	}
}

// search for args or subcommand
func searchArgs(curr *CompTree, searchFunc SearchMethod, segment string) []Suggestion {
	var resp []Suggestion
	// command argument
	if curr.TakesValue {
		if curr.Dynamic != nil {
			return curr.Dynamic(segment)
		}
		return SuggestSomething
	}
	// subcommands
	for k, v := range curr.Sub {
		if searchFunc(k, segment) {
			resp = append(resp, Suggestion{
				Name: k,
				Desc: v.Desc,
			})
		}
	}
	return resp
}
func searchFlags(flags map[string]*CompTree, searchTerm string, fields []string) []Suggestion {
	var resp []Suggestion
	for k, v := range flags {
		if strings.HasPrefix(k, searchTerm) || searchTerm == "" {
			if searchTerm != "" && containsFlag(fields, k) { // don't suggest flag which is already set
				continue
			}
			resp = append(resp, Suggestion{
				Name: k,
				Desc: v.Desc,
			})
		}
	}
	return resp
}

func containsFlag(arr []string, x string) bool {
	if contains(arr, "--"+x) {
		return true
	}
	return contains(arr, "-"+x)
}

func completeFlagValue(curr *CompTree, globalFlags map[string]*CompTree, name string) []Suggestion {
	name = strings.Trim(name, "-=")

	// if command flag expects a value
	if flag := curr.Flags[name]; flag != nil {
		// flag has custom suggestion method
		if flag.Dynamic != nil {
			return flag.Dynamic("")
		}
		// do not suggest anything, wait for user to input value
		if flag.TakesValue {
			return SuggestSomething
		}
	}
	// if global flag expects a value
	if flag := globalFlags[name]; flag != nil {
		// flag has custom suggestion method
		if flag.Dynamic != nil {
			return flag.Dynamic("")
		}
		// do not suggest anything, wait for user to input value
		if flag.TakesValue {
			return SuggestSomething
		}
	}
	return SuggestNothing
}

func completeFlags(curr *CompTree, globalFlags map[string]*CompTree, name string, fields []string) []Suggestion {
	searchTerm := strings.TrimLeft(name, "-")
	// complete flag of current command
	if s := searchFlags(curr.Flags, searchTerm, fields); len(s) > 0 {
		return s
	}
	if s := searchFlags(globalFlags, searchTerm, fields); len(s) > 0 {
		return s
	}
	return SuggestNothing
}

// TODO: think about suggesting multiple args?
// TODO: fix suggesting flags after command arguments
func completeSpace(curr *CompTree, globalFlags map[string]*CompTree, name string) []Suggestion {
	isFlag := strings.HasPrefix(name, "-")
	isCurrent := curr.Name == name // is last arg is current command name

	name = strings.Trim(name, "- ")

	if isFlag {
		// if command flag expects a value
		if flag := curr.Flags[name]; flag != nil {
			// flag has custom suggestion method
			if flag.Dynamic != nil {
				return flag.Dynamic("")
			}
			// do not suggest anything, wait for user to input value
			if flag.TakesValue {
				return SuggestSomething
			}
		}
		// if global flag expects a value
		if flag := globalFlags[name]; flag != nil {
			// flag has custom suggestion method
			if flag.Dynamic != nil {
				return flag.Dynamic("")
			}
			// do not suggest anything, wait for user to input value
			if flag.TakesValue {
				return SuggestSomething
			}
		}
	}

	// if current command has custom suggestion method and
	// last arg was its name or flag which don't expect argument
	if curr.Dynamic != nil && (isCurrent || isFlag) {
		return curr.Dynamic("")
	}

	// suggest subcommand
	if curr.Sub != nil {
		var resp []Suggestion
		for k, v := range curr.Sub {
			resp = append(resp, Suggestion{
				Name: k,
				Desc: v.Desc,
			})
		}
		return resp
	}

	// suggest flags (no subcommands or values is expected so suggest only thing that is available: flags
	if isCurrent && (len(curr.Flags) > 0 || len(globalFlags) > 0) {
		return SuggestFlag
	}
	// there are no flags or last arg is command argument so we can suggest nothing
	// TODO: case when it's after FLAG argument not command
	return SuggestNothing
}
