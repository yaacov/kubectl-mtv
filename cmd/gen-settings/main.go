// Command gen-settings parses ForkliftControllerSpec from the forklift source
// and generates pkg/cmd/settings/types_generated.go with the AllSettings map.
package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"regexp"
	"strings"
	"text/template"
)

type SettingDef struct {
	Name        string
	Type        string // "TypeString", "TypeBool", "TypeInt"
	Default     string // Go literal: `""`, `"true"`, `20`, `nil`
	Description string
	Category    string // e.g. "CategoryFeature"
}

var sectionToCategory = map[string]string{
	"Feature Gates":                              "CategoryFeature",
	"Container Images":                           "CategoryImage",
	"Controller Resource Configuration":          "CategoryController",
	"Inventory Resource Configuration":           "CategoryInventory",
	"API Resource Configuration":                 "CategoryAPI",
	"CLI Download Resource Configuration":        "CategoryCLIDownload",
	"MCP Server Lightspeed Integration":          "CategoryMCP",
	"MCP Server Resource Configuration":          "CategoryMCP",
	"MCP Server Settings":                        "CategoryMCP",
	"UI Plugin Resource Configuration":           "CategoryUIPlugin",
	"Validation Resource Configuration":          "CategoryValidation",
	"ConfigMap Names":                            "CategoryConfigMaps",
	"Migration Settings & Timeouts":              "CategoryPerformance",
	"Ansible Automation Platform (AAP) Settings": "CategoryAAP",
	"Migration Feature-Specific Settings":        "CategoryFeature",
	"Storage & Performance Settings":             "CategoryPerformance",
	"Virt-v2v Settings":                          "CategoryVirtV2V",
	"Hooks Settings":                             "CategoryHook",
	"OVA Settings":                               "CategoryOVA",
	"HyperV Settings":                            "CategoryHyperV",
	"OVA Proxy Settings":                         "CategoryOVAProxy",
	"Volume Populator Settings":                  "CategoryPopulator",
	"Logging & General Settings":                 "CategoryAdvanced",
	"Metrics Settings":                           "CategoryAdvanced",
}

var defaultRe = regexp.MustCompile(`\+kubebuilder:default=(.+)`)
var enumBoolRe = regexp.MustCompile(`\+kubebuilder:validation:Enum="true";"false"`)

func main() {
	inputPath := flag.String("input", "", "Path to forkliftcontroller.go")
	outputPath := flag.String("output", "", "Output path for types_generated.go")
	flag.Parse()

	if *inputPath == "" || *outputPath == "" {
		fmt.Fprintf(os.Stderr, "Usage: gen-settings -input <path> -output <path>\n")
		os.Exit(1)
	}

	settings, err := parseSpec(*inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing spec: %v\n", err)
		os.Exit(1)
	}

	if err := writeOutput(*outputPath, settings); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Generated %d settings → %s\n", len(settings), *outputPath)
}

func parseSpec(path string) ([]SettingDef, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse file: %w", err)
	}

	var specType *ast.TypeSpec
	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}
		for _, spec := range genDecl.Specs {
			ts := spec.(*ast.TypeSpec)
			if ts.Name.Name == "ForkliftControllerSpec" {
				specType = ts
				break
			}
		}
		if specType != nil {
			break
		}
	}

	if specType == nil {
		return nil, fmt.Errorf("ForkliftControllerSpec not found in %s", path)
	}

	structType, ok := specType.Type.(*ast.StructType)
	if !ok {
		return nil, fmt.Errorf("ForkliftControllerSpec is not a struct")
	}

	// Build a sorted list of section header positions from all comments in the file.
	// Section headers are standalone comments like "// Feature Gates" that appear
	// between fields but are NOT attached to any field's Doc.
	type sectionMark struct {
		line     int
		category string
	}
	var sectionMarks []sectionMark

	for _, cg := range f.Comments {
		if len(cg.List) != 1 {
			continue
		}
		text := strings.TrimPrefix(cg.List[0].Text, "//")
		text = strings.TrimSpace(text)
		if cat, ok := sectionToCategory[text]; ok {
			line := fset.Position(cg.Pos()).Line
			sectionMarks = append(sectionMarks, sectionMark{line: line, category: cat})
		}
	}

	// For each field, determine its category based on the last section header before it.
	currentCategory := "CategoryAdvanced"
	var settings []SettingDef

	for _, field := range structType.Fields.List {
		if len(field.Names) == 0 {
			continue
		}
		if field.Tag == nil {
			continue
		}

		fieldLine := fset.Position(field.Pos()).Line

		// Update category to the last section header before this field
		for _, sm := range sectionMarks {
			if sm.line < fieldLine {
				currentCategory = sm.category
			} else {
				break
			}
		}

		jsonName := extractJSONTag(field.Tag.Value)
		if jsonName == "" || jsonName == "-" {
			continue
		}

		commentLines := extractCommentLines(field.Doc)
		if len(commentLines) == 0 {
			continue
		}

		description := extractDescription(commentLines)
		isBool := hasBoolEnum(commentLines)
		defaultVal := extractDefault(commentLines)
		goType := typeExprString(field.Type)
		settingType := determineType(goType, isBool)

		defaultLiteral := formatDefault(defaultVal, settingType)

		settings = append(settings, SettingDef{
			Name:        jsonName,
			Type:        settingType,
			Default:     defaultLiteral,
			Description: description,
			Category:    currentCategory,
		})
	}

	return settings, nil
}

func extractJSONTag(tagLiteral string) string {
	tag := strings.Trim(tagLiteral, "`")
	idx := strings.Index(tag, `json:"`)
	if idx < 0 {
		return ""
	}
	rest := tag[idx+6:]
	end := strings.Index(rest, `"`)
	if end < 0 {
		return ""
	}
	jsonVal := rest[:end]
	if comma := strings.Index(jsonVal, ","); comma >= 0 {
		jsonVal = jsonVal[:comma]
	}
	return jsonVal
}

func extractCommentLines(cg *ast.CommentGroup) []string {
	if cg == nil {
		return nil
	}
	var lines []string
	for _, c := range cg.List {
		text := strings.TrimPrefix(c.Text, "//")
		text = strings.TrimSpace(text)
		if text != "" {
			lines = append(lines, text)
		}
	}
	return lines
}

func extractDescription(lines []string) string {
	var descParts []string
	for _, line := range lines {
		if strings.HasPrefix(line, "+") {
			continue
		}
		descParts = append(descParts, line)
	}
	desc := strings.Join(descParts, " ")
	desc = strings.TrimSuffix(desc, ".")
	return desc
}

func hasBoolEnum(lines []string) bool {
	for _, line := range lines {
		if enumBoolRe.MatchString(line) {
			return true
		}
	}
	return false
}

func extractDefault(lines []string) string {
	for _, line := range lines {
		m := defaultRe.FindStringSubmatch(line)
		if len(m) == 2 {
			return m[1]
		}
	}
	return ""
}

func typeExprString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + typeExprString(t.X)
	case *ast.SelectorExpr:
		return typeExprString(t.X) + "." + t.Sel.Name
	default:
		return "unknown"
	}
}

func determineType(goType string, isBool bool) string {
	if goType == "*int32" || goType == "int32" {
		return "TypeInt"
	}
	if isBool {
		return "TypeBool"
	}
	return "TypeString"
}

func formatDefault(raw string, settingType string) string {
	if raw == "" {
		switch settingType {
		case "TypeInt":
			return "nil"
		case "TypeBool":
			return "nil"
		default:
			return `""`
		}
	}

	// Strip surrounding quotes if present
	unquoted := strings.Trim(raw, `"`)

	switch settingType {
	case "TypeInt":
		return unquoted
	case "TypeBool":
		if unquoted == "true" {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%q", unquoted)
	}
}

var outputTmpl = template.Must(template.New("gen").Parse(`// Code generated by cmd/gen-settings. DO NOT EDIT.
// Source: ForkliftControllerSpec from kubev2v/forklift

package settings

// AllSettings contains all ForkliftController settings derived from ForkliftControllerSpec.
var AllSettings = map[string]SettingDefinition{
{{- range .}}
	"{{.Name}}": {
		Name:        "{{.Name}}",
		Type:        {{.Type}},
		Default:     {{.Default}},
		Description: {{printf "%q" .Description}},
		Category:    {{.Category}},
	},
{{- end}}
}
`))

func writeOutput(path string, settings []SettingDef) (retErr error) {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		if cErr := f.Close(); cErr != nil && retErr == nil {
			retErr = cErr
		}
	}()

	return outputTmpl.Execute(f, settings)
}
