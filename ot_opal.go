package main

import (
	"fmt"
	"sort"
)

// gqlVerbsAndFunctions fetches all OPAL verbs and functions from the tenant.
var gqlVerbsAndFunctions = compileGqlQuery(
	`query VerbsAndFunctions {
		verbsAndFunctions {
			verbs { name description category }
			functions { name description category returnType }
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
		name     string
		category string
		desc     string
	}
	var verbs []verbEntry
	for _, v := range verbsRaw {
		m, ok := v.(object)
		if !ok {
			continue
		}
		name, _ := m["name"].(string)
		category, _ := m["category"].(string)
		desc, _ := m["description"].(string)
		verbs = append(verbs, verbEntry{name, category, desc})
	}
	sort.Slice(verbs, func(i, j int) bool {
		return verbs[i].name < verbs[j].name
	})
	for _, v := range verbs {
		fmt.Fprintf(fa.op, "%s\t%s\t%s\n", v.name, v.category, v.desc)
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
		category   string
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
		category, _ := m["category"].(string)
		returnType, _ := m["returnType"].(string)
		desc, _ := m["description"].(string)
		funcs = append(funcs, funcEntry{name, category, returnType, desc})
	}
	sort.Slice(funcs, func(i, j int) bool {
		return funcs[i].name < funcs[j].name
	})
	for _, f := range funcs {
		fmt.Fprintf(fa.op, "%s\t%s\t%s\t%s\n", f.name, f.category, f.returnType, f.desc)
	}
	return nil
}
