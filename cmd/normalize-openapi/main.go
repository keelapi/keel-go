package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

func main() {
	in := flag.String("in", "", "source OpenAPI document")
	out := flag.String("out", "", "normalized OpenAPI document")
	flag.Parse()

	if *in == "" || *out == "" {
		fmt.Fprintln(os.Stderr, "usage: normalize-openapi -in openapi.json -out normalized.json")
		os.Exit(2)
	}

	data, err := os.ReadFile(*in)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read %s: %v\n", *in, err)
		os.Exit(1)
	}

	var doc any
	if err := json.Unmarshal(data, &doc); err != nil {
		fmt.Fprintf(os.Stderr, "parse %s: %v\n", *in, err)
		os.Exit(1)
	}

	normalized := normalize(doc)
	if root, ok := normalized.(map[string]any); ok {
		root["openapi"] = "3.0.3"
		addMissingPathParameters(root)
	}

	if err := os.MkdirAll(parentDir(*out), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "create output directory: %v\n", err)
		os.Exit(1)
	}

	f, err := os.Create(*out)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create %s: %v\n", *out, err)
		os.Exit(1)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(normalized); err != nil {
		fmt.Fprintf(os.Stderr, "write %s: %v\n", *out, err)
		os.Exit(1)
	}
}

func parentDir(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			if i == 0 {
				return "/"
			}
			return path[:i]
		}
	}
	return "."
}

func normalize(v any) any {
	switch value := v.(type) {
	case []any:
		for i, item := range value {
			value[i] = normalize(item)
		}
		return value
	case map[string]any:
		for key, item := range value {
			value[key] = normalize(item)
		}
		normalizeSchemaObject(value)
		return value
	default:
		return value
	}
}

func normalizeSchemaObject(schema map[string]any) {
	if value, ok := schema["const"]; ok {
		if _, exists := schema["enum"]; !exists {
			schema["enum"] = []any{value}
		}
		delete(schema, "const")
	}

	for _, keyword := range []string{"anyOf", "oneOf"} {
		normalizeNullableUnion(schema, keyword)
	}

	if schema["type"] == "null" {
		delete(schema, "type")
		schema["nullable"] = true
	}
}

func normalizeNullableUnion(schema map[string]any, keyword string) {
	raw, ok := schema[keyword].([]any)
	if !ok {
		return
	}

	nonNull := raw[:0]
	nullable := false
	for _, branch := range raw {
		if branchMap, ok := branch.(map[string]any); ok && branchMap["type"] == "null" {
			nullable = true
			continue
		}
		nonNull = append(nonNull, branch)
	}
	if !nullable {
		schema[keyword] = nonNull
		return
	}

	schema["nullable"] = true
	switch len(nonNull) {
	case 0:
		delete(schema, keyword)
	case 1:
		delete(schema, keyword)
		mergeSchema(schema, nonNull[0])
	default:
		schema[keyword] = nonNull
	}
}

func mergeSchema(dst map[string]any, src any) {
	srcMap, ok := src.(map[string]any)
	if !ok {
		return
	}
	for key, value := range srcMap {
		if _, exists := dst[key]; !exists {
			dst[key] = value
		}
	}
}

func addMissingPathParameters(root map[string]any) {
	paths, ok := root["paths"].(map[string]any)
	if !ok {
		return
	}
	for path, rawPathItem := range paths {
		names := pathParameterNames(path)
		if len(names) == 0 {
			continue
		}

		pathItem, ok := rawPathItem.(map[string]any)
		if !ok {
			continue
		}
		pathLevel := parameterNameSet(pathItem["parameters"])
		for method, rawOperation := range pathItem {
			if !isOperationMethod(method) {
				continue
			}
			operation, ok := rawOperation.(map[string]any)
			if !ok {
				continue
			}
			operationLevel := parameterNameSet(operation["parameters"])
			for _, name := range names {
				if pathLevel[name] || operationLevel[name] {
					continue
				}
				operation["parameters"] = appendParameter(operation["parameters"], name)
				operationLevel[name] = true
			}
		}
	}
}

func pathParameterNames(path string) []string {
	var names []string
	for i := 0; i < len(path); i++ {
		if path[i] != '{' {
			continue
		}
		start := i + 1
		for i < len(path) && path[i] != '}' {
			i++
		}
		if i > start && i < len(path) {
			names = append(names, path[start:i])
		}
	}
	return names
}

func parameterNameSet(raw any) map[string]bool {
	names := map[string]bool{}
	parameters, ok := raw.([]any)
	if !ok {
		return names
	}
	for _, parameter := range parameters {
		param, ok := parameter.(map[string]any)
		if !ok || param["in"] != "path" {
			continue
		}
		if name, ok := param["name"].(string); ok {
			names[name] = true
		}
	}
	return names
}

func appendParameter(raw any, name string) []any {
	parameters, _ := raw.([]any)
	return append(parameters, map[string]any{
		"name":     name,
		"in":       "path",
		"required": true,
		"schema": map[string]any{
			"type": "string",
		},
	})
}

func isOperationMethod(method string) bool {
	switch method {
	case "get", "put", "post", "delete", "options", "head", "patch", "trace":
		return true
	default:
		return false
	}
}
