package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/structtag"
	field "github.com/go-utils/meta"
	"golang.org/x/xerrors"
)

var (
	prefix        = flag.String("prefix", "", "Prefix for table name")
	disableMeta   = flag.Bool("disable-meta", false, "Disable meta embed for Lock")
	isShowVersion = flag.Bool("v", false, "print version")
)

func main() {
	flag.Parse()

	if *isShowVersion {
		fmt.Printf("DynamoDB Model Generator: %s\n", AppVersion)
		return
	}

	l := flag.NArg()

	if l < 1 {
		fmt.Println("You have to specify the struct name of target")
		os.Exit(1)
	}

	if err := run(flag.Arg(0), *prefix, *disableMeta); err != nil {
		log.Fatal(err.Error())
	}
}

func run(structName, prefix string, isDisableMeta bool) error {
	disableMeta = &isDisableMeta
	fs := token.NewFileSet()
	pkgs, err := parser.ParseDir(fs, ".", nil, parser.AllErrors)

	if err != nil {
		panic(err)
	}

	for name, v := range pkgs {
		if strings.HasSuffix(name, "_test") {
			continue
		}

		return traverse(v, fs, structName, prefix)
	}

	return nil
}

func traverse(pkg *ast.Package, fs *token.FileSet, structName, prefix string) error {
	gen := &generator{PackageName: pkg.Name}
	for name, file := range pkg.Files {
		gen.FileName = strings.TrimSuffix(filepath.Base(name), ".go")
		gen.GeneratedFileName = gen.FileName + "_gen"

		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			if genDecl.Tok != token.TYPE {
				continue
			}

			for _, spec := range genDecl.Specs {
				// 型定義
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				name := typeSpec.Name.Name

				if name != structName {
					continue
				}

				// structの定義
				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					continue
				}
				gen.StructName = name
				gen.TableName = prefix + name

				return generate(gen, fs, structType)
			}
		}
	}

	return xerrors.Errorf("no such struct: %s", structName)
}

func generate(gen *generator, fs *token.FileSet, structType *ast.StructType) error {
	var metaList map[string]*field.Field
	if !*disableMeta {
		var err error
		fList := field.ListAllField(structType.Fields, "", false)
		metas, _, err := field.SearchMetaProperties(fList)
		if err != nil {
			return err
		}
		metaList = make(map[string]*field.Field)
		for _, m := range metas {
			metaList[m.Name] = m
		}
	}
	gen.MetaFields = metaList
	fieldInfos := make([]*FieldInfo, 0, len(structType.Fields.List))

	for _, field := range structType.Fields.List {
		// structの各fieldを調査
		if len(field.Names) > 1 {
			return xerrors.New("`field.Names` must have only one element")
		}
		name := ""
		if field.Names == nil || len(field.Names) == 0 {
			switch field.Type.(type) {
			case *ast.Ident:
				name = field.Type.(*ast.Ident).Name
			case *ast.SelectorExpr:
				name = field.Type.(*ast.SelectorExpr).Sel.Name
			}
		} else {
			name = field.Names[0].Name
		}

		pos := fs.Position(field.Pos()).String()

		typeName := getTypeName(field.Type)

		switch name {
		case CreatedAt, CreateTime:
			if gen.EnableCreateTime == true {
				return xerrors.New(fmt.Sprintf("duplicated %s or %s", CreatedAt, CreateTime))
			}
			gen.EnableCreateTime = true
			gen.CreateTimeName = name
			gen.CreateTimeType = typeName
		case UpdatedAt, UpdateTime:
			if gen.EnableUpdateTime == true {
				return xerrors.New(fmt.Sprintf("duplicated %s or %s", UpdatedAt, UpdateTime))
			}
			gen.EnableUpdateTime = true
			gen.UpdateTimeName = name
			gen.UpdateTimeDynamoTag = name
			gen.UpdateTimeType = typeName
		}

		if strings.HasPrefix(typeName, "[]") {
			gen.SliceExist = true
		}

		if field.Tag == nil {
			fieldInfo := &FieldInfo{
				DynamoTag: name,
				Field:     name,
				FieldType: typeName,
				Tags:      nil,
			}
			gen.FieldInfos = append(gen.FieldInfos, fieldInfo)
			fieldInfos = append(fieldInfos, fieldInfo)
			continue
		}

		if tags, err := structtag.Parse(strings.Trim(field.Tag.Value, "`")); err != nil {
			log.Printf(
				"%s: tag for %s in struct %s in %s",
				pos, name, gen.StructName, gen.GeneratedFileName+".go",
			)
			continue
		} else {
			fieldInfo := &FieldInfo{
				DynamoTag: name,
				Field:     name,
				FieldType: typeName,
			}
			dynamoTag, err := tags.Get("dynamo")
			fieldInfos = append(fieldInfos, fieldInfo)
			if err != nil {
				gen.FieldInfos = append(gen.FieldInfos, fieldInfo)
				continue
			}
			sp := strings.Split(dynamoTag.Value(), ",")
			fieldInfo.Tags = &FieldParsedTags{
				Raw: sp,
			}
			for index, s := range sp {
				sp[index] = strings.TrimSpace(s)
			}

			if err := dynamoTagCheck(pos, sp[0]); err != nil {
				return xerrors.Errorf("tag validator failed: %w", err)
			}
			if len(sp) == 1 {
				if sp[0] != "" {
					fieldInfo.DynamoTag = sp[0]
					switch name {
					case UpdatedAt, UpdateTime:
						gen.UpdateTimeDynamoTag = sp[0]
					}
					fieldInfo.Tags.Name = sp[0]
				}
				gen.FieldInfos = append(gen.FieldInfos, fieldInfo)
				continue
			}

			fieldInfo.Tags = parseTags(sp)
			if fieldInfo.Tags.IsUnique {
				if sp[0] != "" {
					fieldInfo.DynamoTag = sp[0]
				}
				gen.FieldInfos = append(gen.FieldInfos, fieldInfo)
			}

			if err := keyFieldHandle(gen, sp[0], fieldInfo.Tags.KeyKind, name, typeName, pos); err != nil {
				return xerrors.Errorf("error in keyFieldHandle: %w", err)
			}
			if gen.HashKeyFieldName != "" {
				if _, err := tags.Get("auto"); err == nil {
					gen.AutoGeneration = true
				}
			}
		}
	}

	gen.UniqueFields = make(map[string]*UniqueField)
	for _, f := range fieldInfos {
		if f.Tags == nil || !f.Tags.IsUnique {
			continue
		}
		varName := fmt.Sprintf("%s%s", gen.TableName, f.Field)
		lowVarName := strings.ToLower(string(varName[0])) + varName[1:]

		gen.UniqueFields[f.Field] = &UniqueField{
			VarName:     lowVarName,
			StructName:  fmt.Sprintf("%sUnique", varName),
			SubjectName: fmt.Sprintf("%sSubject", lowVarName),
			Field: field.Field{
				Name: f.Field,
				Type: f.FieldType,
			},
		}
	}
	if len(gen.UniqueFields) == 0 {
		gen.UniqueFields = nil
	}

	if gen.EnableCreateTime || gen.EnableUpdateTime {
		if !(gen.EnableCreateTime && gen.EnableUpdateTime) {
			return xerrors.New("requires both CreatedAt and UpdatedAt")
		}
		if gen.CreateTimeType != gen.UpdateTimeType {
			return xerrors.Errorf("the type is different")
		}
		if strings.HasSuffix(gen.CreateTimeType, ".UnixTime") {
			gen.EnableDDA = true
		}
	}

	{
		fp, err := os.Create(gen.GeneratedFileName + ".go")
		if err != nil {
			panic(err)
		}
		defer fp.Close()

		gen.generate(fp)
	}

	{
		fp, err := os.Create("constant.go")
		if err != nil {
			panic(err)
		}
		defer fp.Close()
		gen.generateConstant(fp)
	}

	{
		fp, err := os.Create("misc.go")
		if err != nil {
			panic(err)
		}
		defer fp.Close()
		gen.generateMisc(fp)
	}

	return nil
}

func parseTags(tags []string) *FieldParsedTags {
	p := new(FieldParsedTags)
	p.Raw = tags

	for i, tag := range tags {
		if i == 0 {
			p.Name = tag
			continue
		}

		switch strings.ToLower(tag) {
		case "hash":
			p.KeyKind = KeyKindHash
		case "range":
			p.KeyKind = KeyKindRange
		case "unique":
			p.IsUnique = true
		}
	}
	return p
}

func keyFieldHandle(gen *generator, label string, keyKind KeyKind, name, typeName, pos string) error {
	switch keyKind {
	case KeyKindHash:
		gen.HashKeyFieldName = name
		gen.HashKeyFieldType = typeName

		if gen.HashKeyFieldType != typeInt &&
			gen.HashKeyFieldType != typeInt64 &&
			gen.HashKeyFieldType != typeString {
			return xerrors.Errorf("%s: supported key types are int, int64, string", pos)
		}

		gen.HashKeyValueName = name
		if label == "" {
			gen.HashKeyFieldTagName = name
		} else {
			gen.HashKeyFieldTagName = label
		}
	case KeyKindRange:
		if gen.RangeKeyFieldName != "" || gen.RangeKeyFieldType != "" {
			return xerrors.Errorf("%s: RangeKey already exists", pos)
		}
		gen.RangeKeyFieldName = name
		gen.RangeKeyFieldType = typeName

		if gen.RangeKeyFieldType != typeInt &&
			gen.RangeKeyFieldType != typeInt64 &&
			gen.RangeKeyFieldType != typeString {
			return xerrors.Errorf("%s: supported key types are int, int64, string", pos)
		}

		gen.RangeKeyValueName = name
		if label == "" {
			gen.RangeKeyFieldTagName = name
		} else {
			gen.RangeKeyFieldTagName = label
		}
	}

	return nil
}
