package main

import (
	"fmt"
	"sort"
	"strings"
)

// joinCategories converts a categories array value to a comma-separated string.
// The API returns categories as []string rather than a single string.
func joinCategories(cats any) string {
	a, ok := cats.(array)
	if !ok {
		return ""
	}
	parts := make([]string, 0, len(a))
	for _, c := range a {
		if s, ok := c.(string); ok {
			parts = append(parts, s)
		}
	}
	return strings.Join(parts, ",")
}

// gqlVerbsAndFunctions fetches all OPAL verbs and functions from the tenant.
// Actual API schema uses "categories" (array of strings) not "category" (single string).
var gqlVerbsAndFunctions = compileGqlQuery(
	`query VerbsAndFunctions {
		verbsAndFunctions {
			verbs { name description categories }
			functions { name description categories returnType }
		}
	}`,
	"data", "verbsAndFunctions",
)

func cmdOpalVerbs(fa FuncArgs) error {
	result, err := gqlVerbsAndFunctions.query(fa.cfg, fa.op, fa.hc, object{})
	if err != nil {
		return err
	}
	res, ok := result.(object)
	if !ok {
		return fmt.Errorf("opal verbs: unexpected response type")
	}
	verbsRaw, _ := res["verbs"].(array)

	type verbEntry struct {
		name       string
		categories string
		desc       string
	}
	var verbs []verbEntry
	for _, v := range verbsRaw {
		m, ok := v.(object)
		if !ok {
			continue
		}
		name, _ := m["name"].(string)
		categories := joinCategories(m["categories"])
		desc, _ := m["description"].(string)
		verbs = append(verbs, verbEntry{name, categories, desc})
	}
	sort.Slice(verbs, func(i, j int) bool {
		return verbs[i].name < verbs[j].name
	})
	for _, v := range verbs {
		fmt.Fprintf(fa.op, "%s\t%s\t%s\n", v.name, v.categories, v.desc)
	}
	return nil
}

func cmdOpalFunctions(fa FuncArgs) error {
	result, err := gqlVerbsAndFunctions.query(fa.cfg, fa.op, fa.hc, object{})
	if err != nil {
		return err
	}
	res, ok := result.(object)
	if !ok {
		return fmt.Errorf("opal functions: unexpected response type")
	}
	funcsRaw, _ := res["functions"].(array)

	type funcEntry struct {
		name       string
		categories string
		returnType string
		desc       string
	}
	var funcs []funcEntry
	for _, f := range funcsRaw {
		m, ok := f.(object)
		if !ok {
			continue
		}
		name, _ := m["name"].(string)
		categories := joinCategories(m["categories"])
		returnType, _ := m["returnType"].(string)
		desc, _ := m["description"].(string)
		funcs = append(funcs, funcEntry{name, categories, returnType, desc})
	}
	sort.Slice(funcs, func(i, j int) bool {
		return funcs[i].name < funcs[j].name
	})
	for _, f := range funcs {
		fmt.Fprintf(fa.op, "%s\t%s\t%s\t%s\n", f.name, f.categories, f.returnType, f.desc)
	}
	return nil
}
