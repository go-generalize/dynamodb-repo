package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"text/template"

	field "github.com/go-utils/meta"

	_ "github.com/go-generalize/dynamodb-repo/statik"
	"github.com/go-utils/plural"
	"github.com/iancoleman/strcase"
	"github.com/rakyll/statik/fs"
)

var statikFS http.FileSystem

func init() {
	var err error
	statikFS, err = fs.New()
	if err != nil {
		log.Fatal(err)
	}
}

type KeyKind string

const (
	KeyKindHash  KeyKind = "hash"
	KeyKindRange KeyKind = "range"
)

type FieldParsedTags struct {
	Name     string
	KeyKind  KeyKind
	IsUnique bool
	Raw      []string
}

type FieldInfo struct {
	DynamoTag string
	Field     string
	FieldType string
	Tags      *FieldParsedTags
}

type ImportInfo struct {
	Name string
}

type UniqueField struct {
	VarName     string
	StructName  string
	SubjectName string
	field.Field
}

type generator struct {
	AppVersion        string
	PackageName       string
	ImportName        string
	ImportList        []ImportInfo
	GeneratedFileName string
	FileName          string
	StructName        string
	TableName         string

	RepositoryStructName    string
	RepositoryInterfaceName string

	HashKeyFieldName    string
	HashKeyFieldTagName string
	HashKeyFieldType    string
	HashKeyValueName    string // lower camel case

	RangeKeyFieldName    string
	RangeKeyFieldTagName string
	RangeKeyFieldType    string
	RangeKeyValueName    string // lower camel case

	FieldInfos []*FieldInfo

	AutoGeneration      bool
	BoolCriteriaCnt     int
	FieldInfoForIndexes *FieldInfo
	SliceExist          bool

	EnableCreateTime    bool
	CreateTimeName      string
	CreateTimeType      string
	EnableUpdateTime    bool
	UpdateTimeName      string
	UpdateTimeDynamoTag string
	UpdateTimeType      string
	EnableDDA           bool

	MetaFields   map[string]*field.Field
	UniqueFields map[string]*UniqueField
}

func (g *generator) setting() {
	g.AppVersion = AppVersion
	g.RepositoryInterfaceName = g.StructName + "Repository"
	g.RepositoryStructName = strcase.ToLowerCamel(g.RepositoryInterfaceName)
}

func (g *generator) generate(writer io.Writer) {
	g.setting()
	funcMap := g.setFuncMap()
	contents := getFileContents("gen")

	t := template.Must(template.New("Template").Funcs(funcMap).Parse(contents))

	if err := t.Execute(writer, g); err != nil {
		log.Printf("failed to execute template: %+v", err)
	}
}

func (g *generator) generateConstant(writer io.Writer) {
	contents := getFileContents("constant")

	t := template.Must(template.New("TemplateConstant").Parse(contents))

	if err := t.Execute(writer, g); err != nil {
		log.Printf("failed to execute template: %+v", err)
	}
}

func (g *generator) generateMisc(writer io.Writer) {
	contents := getFileContents("misc")

	t := template.Must(template.New("TemplateConstant").Parse(contents))

	if err := t.Execute(writer, g); err != nil {
		log.Printf("failed to execute template: %+v", err)
	}
}

func (g *generator) setFuncMap() template.FuncMap {
	return template.FuncMap{
		"Parse": func(fieldType string) string {
			fieldType = strings.TrimPrefix(fieldType, "[]")
			fn := "Int"
			switch fieldType {
			case typeInt:
			case typeInt64:
				fn = "Int64"
			case typeFloat64:
				fn = "Float64"
			case typeString:
				fn = "String"
			case typeBool:
				fn = "Bool"
			default:
				panic("invalid types")
			}
			return fn
		},
		"HasPrefixSlice": func(types string) bool {
			return strings.HasPrefix(types, "[]")
		},
		"RangeKeyArgCheckPairs": func() string {
			if g.RangeKeyFieldName != "" {
				return fmt.Sprintf("pairs map[%s]%s", g.HashKeyFieldType, g.RangeKeyFieldType)
			}
			return fmt.Sprintf("%s []%s", plural.Convert(g.HashKeyValueName), g.HashKeyFieldType)
		},
		"RangeKeyForTerms": func() string {
			if g.RangeKeyFieldName != "" {
				return fmt.Sprintf("key, value := range pairs")
			}
			return fmt.Sprintf("_, %s := range %s", g.HashKeyFieldTagName, plural.Convert(g.HashKeyValueName))
		},
		"FuncNameByValue": func() string {
			if g.RangeKeyFieldName != "" {
				return "Pair"
			}
			return g.HashKeyFieldName
		},
		"GenerationKey": func() string {
			switch g.HashKeyFieldType {
			case typeString:
				return "uuid.New().String()"
			case typeInt:
				return "int(uuid.New().ID())"
			case typeInt64:
				return "int64(uuid.New().ID())"
			}
			return ""
		},
		"Now": func(typeName string) string {
			if strings.HasSuffix(typeName, ".UnixTime") {
				return "dda.UnixTime(time.Now())"
			}
			return "time.Now()"
		},
		"PluralForm": func(word string) string {
			return plural.Convert(word)
		},
		"HasKey": func(fields map[string]*field.Field, key string) bool {
			_, ok := fields[key]
			return ok
		},
		"GetMetaKeyWithPath": func(fields map[string]*field.Field, key string) string {
			p := fields[key].ParentPath
			if p == "" {
				return key
			}

			return fmt.Sprintf("%s.%s", p, key)
		},
	}
}
