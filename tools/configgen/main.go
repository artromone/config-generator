package main

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

type ConfigField struct {
	Name         string
	EnvVar       string
	GoType       string
	DefaultValue string
	Required     bool
	YamlPath     string
	Tags         string
}

type ConfigStruct struct {
	Name   string
	Fields []ConfigField
}

type YamlNode struct {
	Key      string
	Value    interface{}
	Children map[string]*YamlNode
	EnvVars  []string
	Path     string
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run tools/configgen/main.go [generate|validate]")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "generate":
		generateConfig()
	case "validate":
		validateConfig()
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func generateConfig() {
	templateContent, err := os.ReadFile("config/config.yaml.template")
	if err != nil {
		panic(fmt.Sprintf("Failed to read template: %v", err))
	}

	var yamlData interface{}
	if err := yaml.Unmarshal(templateContent, &yamlData); err != nil {
		panic(fmt.Sprintf("Failed to parse YAML: %v", err))
	}

	root := buildConfigTree(yamlData, "")
	envVars := extractEnvVarsFromContent(string(templateContent))
	structs := generateStructsFromTree(root, envVars)

	generateGoCode(structs, envVars)
	generateEnvFiles(envVars)

	fmt.Println("✅ Configuration files generated successfully!")
}

func buildConfigTree(data interface{}, path string) *YamlNode {
	node := &YamlNode{
		Children: make(map[string]*YamlNode),
		Path:     path,
	}

	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			childPath := key
			if path != "" {
				childPath = path + "." + key
			}

			child := buildConfigTree(value, childPath)
			child.Key = key
			node.Children[key] = child
		}
	case []interface{}:
		node.Value = data
	default:
		node.Value = data
		if str, ok := data.(string); ok {
			envVars := extractEnvVarsFromString(str)
			node.EnvVars = envVars
		}
	}

	return node
}

func extractEnvVarsFromString(s string) []string {
	re := regexp.MustCompile(`\$\{([^}:]+)(?::([^}]*))?\}`)
	matches := re.FindAllStringSubmatch(s, -1)

	var vars []string
	for _, match := range matches {
		vars = append(vars, match[1])
	}
	return vars
}

func extractEnvVarsFromContent(content string) []ConfigField {
	re := regexp.MustCompile(`\$\{([^}:]+)(?::([^}]*))?\}`)
	matches := re.FindAllStringSubmatch(content, -1)

	var fields []ConfigField
	seen := make(map[string]bool)

	for _, match := range matches {
		envVar := match[1]
		defaultValue := ""
		if len(match) > 2 {
			defaultValue = match[2]
		}

		if seen[envVar] {
			continue
		}
		seen[envVar] = true

		field := ConfigField{
			Name:         envVarToFieldName(envVar),
			EnvVar:       envVar,
			GoType:       inferGoType(defaultValue),
			DefaultValue: defaultValue,
			Required:     defaultValue == "",
			YamlPath:     getYamlPathFromEnvVar(content, envVar),
		}

		fields = append(fields, field)
	}

	return fields
}

func generateStructsFromTree(node *YamlNode, envVars []ConfigField) []ConfigStruct {
	var allStructs []ConfigStruct
	structMap := make(map[string]bool) // Для предотвращения дублирования

	// Создаем главную структуру Config
	mainStruct := ConfigStruct{
		Name:   "Config",
		Fields: []ConfigField{},
	}

	// Обрабатываем каждый раздел верхнего уровня
	for key, child := range node.Children {
		if shouldSkipField(key, child) {
			continue
		}

		structName := strings.Title(toCamelCase(key))

		// Генерируем структуру и все её вложенные структуры
		childStructs := generateStructFromNode(child, structName, envVars, key, structMap)

		if len(childStructs) > 0 && len(childStructs[0].Fields) > 0 {
			// Добавляем все сгенерированные структуры
			allStructs = append(allStructs, childStructs...)

			// Добавляем поле в главную структуру
			mainStruct.Fields = append(mainStruct.Fields, ConfigField{
				Name:     structName,
				GoType:   structName + "Struct",
				YamlPath: key,
				Tags:     fmt.Sprintf("`koanf:\"%s\"`", key),
			})
		}
	}

	// Добавляем главную структуру в начало
	result := []ConfigStruct{mainStruct}
	result = append(result, allStructs...)

	return result
}

func shouldSkipField(key string, node *YamlNode) bool {
	// Пропускаем YAML якоря и простые строковые значения без переменных окружения
	if len(node.Children) == 0 && len(node.EnvVars) == 0 {
		return true
	}
	return false
}

func generateStructFromNode(node *YamlNode, structName string, envVars []ConfigField, yamlPath string, structMap map[string]bool) []ConfigStruct {
	if structMap[structName] {
		return []ConfigStruct{} // Уже создана
	}
	structMap[structName] = true

	var allStructs []ConfigStruct

	struct_ := ConfigStruct{
		Name:   structName,
		Fields: []ConfigField{},
	}

	// Обрабатываем дочерние узлы
	for key, child := range node.Children {
		fieldName := strings.Title(toCamelCase(key))
		childPath := yamlPath + "." + key

		if len(child.Children) > 0 {
			// Это вложенная структура
			childStructName := structName + fieldName

			// Рекурсивно генерируем вложенные структуры
			childStructs := generateStructFromNode(child, childStructName, envVars, childPath, structMap)
			allStructs = append(allStructs, childStructs...)

			// Добавляем поле для вложенной структуры
			struct_.Fields = append(struct_.Fields, ConfigField{
				Name:     fieldName,
				GoType:   childStructName + "Struct",
				YamlPath: key,
				Tags:     fmt.Sprintf("`koanf:\"%s\"`", key),
			})
		} else {
			// Это обычное поле
			goType := "string"
			envVar := ""

			// Ищем соответствующую переменную окружения
			for _, env := range envVars {
				if matchesEnvVar(childPath, env.EnvVar) {
					goType = env.GoType
					envVar = env.EnvVar
					break
				}
			}

			// Определяем тип из значения
			if child.Value != nil {
				goType = inferGoTypeFromValue(child.Value)
			}

			tags := fmt.Sprintf("`koanf:\"%s\"`", key)
			if envVar != "" {
				tags = fmt.Sprintf("`koanf:\"%s\" env:\"%s\"`", key, envVar)
			}

			struct_.Fields = append(struct_.Fields, ConfigField{
				Name:     fieldName,
				GoType:   goType,
				YamlPath: key,
				EnvVar:   envVar,
				Tags:     tags,
			})
		}
	}

	// Добавляем текущую структуру в начало списка
	if len(struct_.Fields) > 0 {
		result := []ConfigStruct{struct_}
		result = append(result, allStructs...)
		return result
	}

	return allStructs
}

func matchesEnvVar(yamlPath, envVar string) bool {
	// Улучшенная логика сопоставления
	pathParts := strings.Split(strings.ToLower(yamlPath), ".")
	envParts := strings.Split(strings.ToLower(envVar), "_")

	// Проверяем точное соответствие частей
	for _, envPart := range envParts {
		for _, pathPart := range pathParts {
			if envPart == pathPart {
				return true
			}
		}
	}

	// Проверяем вхождение подстрок
	pathStr := strings.Join(pathParts, "")
	envStr := strings.Join(envParts, "")

	return strings.Contains(envStr, pathStr) || strings.Contains(pathStr, envStr)
}

func inferGoTypeFromValue(value interface{}) string {
	switch v := value.(type) {
	case bool:
		return "bool"
	case int, int32, int64:
		return "int"
	case float32, float64:
		return "float64"
	case string:
		lower := strings.ToLower(v)
		if lower == "true" || lower == "false" {
			return "bool"
		}
		if matched, _ := regexp.MatchString(`^\d+$`, v); matched {
			return "int"
		}
		if matched, _ := regexp.MatchString(`^\d+\.\d+$`, v); matched {
			return "float64"
		}
		return "string"
	case []interface{}:
		return "[]string"
	default:
		return "string"
	}
}

func toCamelCase(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '-' || r == '_'
	})

	for i, part := range parts {
		parts[i] = strings.Title(strings.ToLower(part))
	}

	return strings.Join(parts, "")
}

func envVarToFieldName(envVar string) string {
	return toCamelCase(strings.ToLower(envVar))
}

func inferGoType(defaultValue string) string {
	if defaultValue == "" {
		return "string"
	}

	switch strings.ToLower(defaultValue) {
	case "true", "false":
		return "bool"
	default:
		if matched, _ := regexp.MatchString(`^\d+$`, defaultValue); matched {
			return "int"
		}
		if matched, _ := regexp.MatchString(`^\d+\.\d+$`, defaultValue); matched {
			return "float64"
		}
		return "string"
	}
}

func getYamlPathFromEnvVar(content, envVar string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.Contains(line, "${"+envVar) {
			return extractYamlPath(line)
		}
	}
	return ""
}

func extractYamlPath(line string) string {
	trimmed := strings.TrimSpace(line)
	if strings.Contains(trimmed, ":") {
		return strings.Split(trimmed, ":")[0]
	}
	return ""
}

func generateGoCode(structs []ConfigStruct, envVars []ConfigField) {
	tmpl := `// Code generated by configgen. DO NOT EDIT.
package config

import (
	"fmt"
	"strings"
	"github.com/knadh/koanf/v2"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
)

{{range .Structs}}
type {{.Name}}{{if ne .Name "Config"}}Struct{{end}} struct {
{{range .Fields}}	{{.Name}} {{.GoType}} {{.Tags}}
{{end}}}

{{end}}

func NewConfig() (*Config, error) {
	k := koanf.New(".")
	
	// Загружаем конфигурацию из файла
	if err := k.Load(file.Provider("config/config.yaml"), yaml.Parser()); err != nil {
		return nil, fmt.Errorf("error loading config file: %w", err)
	}

	// Переопределяем переменными окружения
	if err := k.Load(env.Provider("", ".", func(s string) string {
		return strings.Replace(strings.ToLower(s), "_", ".", -1)
	}), nil); err != nil {
		return nil, fmt.Errorf("error loading env vars: %w", err)
	}

	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) Validate() error {
	// Добавьте свою валидацию здесь
	return nil
}
`

	data := struct {
		Structs []ConfigStruct
	}{
		Structs: structs,
	}

	t := template.Must(template.New("config").Parse(tmpl))

	file, err := os.Create("config/config.go")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	if err := t.Execute(file, data); err != nil {
		panic(err)
	}
}

// Остальные функции остаются без изменений...
func generateEnvFiles(fields []ConfigField) {
	generateEnvExample(fields)
	generateEnvLocal(fields)
}

func generateEnvExample(fields []ConfigField) {
	file, err := os.Create(".env.example")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	file.WriteString("# Generated environment variables\n")
	file.WriteString("# Copy this file to .env.local and fill in your values\n\n")

	sort.Slice(fields, func(i, j int) bool {
		return fields[i].EnvVar < fields[j].EnvVar
	})

	for _, field := range fields {
		if field.Required {
			file.WriteString(fmt.Sprintf("%s=\n", field.EnvVar))
		} else {
			file.WriteString(fmt.Sprintf("%s=%s\n", field.EnvVar, field.DefaultValue))
		}
	}
}

func generateEnvLocal(fields []ConfigField) {
	if _, err := os.Stat(".env.local"); err == nil {
		return
	}

	file, err := os.Create(".env.local")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	file.WriteString("# Local environment variables\n")
	file.WriteString("# Add your actual values here\n\n")

	for _, field := range fields {
		file.WriteString(fmt.Sprintf("%s=%s\n", field.EnvVar, field.DefaultValue))
	}
}

func validateConfig() {
	fmt.Println("🔍 Validating configuration...")

	templateFields := extractEnvVarsFromTemplate()
	envFields := extractEnvVarsFromEnvFile()

	missing := findMissingVars(templateFields, envFields)
	if len(missing) > 0 {
		fmt.Printf("❌ Missing environment variables: %v\n", missing)
		os.Exit(1)
	}

	extra := findExtraVars(templateFields, envFields)
	if len(extra) > 0 {
		fmt.Printf("⚠️  Extra environment variables: %v\n", extra)
	}

	fmt.Println("✅ Configuration validation passed!")
}

func extractEnvVarsFromTemplate() []string {
	content, err := os.ReadFile("config/config.yaml.template")
	if err != nil {
		panic(err)
	}

	re := regexp.MustCompile(`\$\{([^}:]+)(?::([^}]*))?\}`)
	matches := re.FindAllStringSubmatch(string(content), -1)

	var vars []string
	seen := make(map[string]bool)
	for _, match := range matches {
		envVar := match[1]
		if !seen[envVar] {
			vars = append(vars, envVar)
			seen[envVar] = true
		}
	}

	return vars
}

func extractEnvVarsFromEnvFile() []string {
	content, err := os.ReadFile(".env.example")
	if err != nil {
		return []string{}
	}

	var vars []string
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			parts := strings.Split(line, "=")
			if len(parts) >= 2 {
				vars = append(vars, parts[0])
			}
		}
	}

	return vars
}

func findMissingVars(template, env []string) []string {
	envMap := make(map[string]bool)
	for _, v := range env {
		envMap[v] = true
	}

	var missing []string
	for _, v := range template {
		if !envMap[v] {
			missing = append(missing, v)
		}
	}

	return missing
}

func findExtraVars(template, env []string) []string {
	templateMap := make(map[string]bool)
	for _, v := range template {
		templateMap[v] = true
	}

	var extra []string
	for _, v := range env {
		if !templateMap[v] {
			extra = append(extra, v)
		}
	}

	return extra
}
