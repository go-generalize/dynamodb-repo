package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/structtag"
	"github.com/go-utils/cont"
	"golang.org/x/xerrors"
)

func init() {
	for _, x := range supportType {
		if x != typeTime {
			supportType = append(supportType, "[]"+x)
		}
	}
}

func main() {
	l := len(os.Args)
	if l < 2 {
		fmt.Println("You have to specify the struct name of target")
		os.Exit(1)
	}

	if err := run(os.Args[1]); err != nil {
		log.Fatal(err.Error())
	}
}

func run(structName string) error {
	fs := token.NewFileSet()
	pkgs, err := parser.ParseDir(fs, ".", nil, parser.AllErrors)

	if err != nil {
		panic(err)
	}

	for name, v := range pkgs {
		if strings.HasSuffix(name, "_test") {
			continue
		}

		return traverse(v, fs, structName)
	}

	return nil
}

func traverse(pkg *ast.Package, fs *token.FileSet, structName string) error {
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

				return generate(gen, fs, structType)
			}
		}
	}

	return xerrors.Errorf("no such struct: %s", structName)
}

func generate(gen *generator, fs *token.FileSet, structType *ast.StructType) error {
	fieldLabel = gen.StructName + queryLabel
	for _, field := range structType.Fields.List {
		// structの各fieldを調査
		if len(field.Names) != 1 {
			return xerrors.New("`field.Names` must have only one element")
		}
		name := field.Names[0].Name

		pos := fs.Position(field.Pos()).String()

		typeName := getTypeName(field.Type)
		if !cont.Contains(supportType, typeName) {
			log.Printf(
				"%s: the type of `%s` is an invalid type in struct `%s` [%s]\n",
				pos, name, gen.StructName, typeName,
			)
			continue
		}

		if strings.HasPrefix(typeName, "[]") {
			gen.SliceExist = true
		}

		if field.Tag == nil {
			fieldInfo := &FieldInfo{
				DynamoTag: name,
				Field:     name,
				FieldType: typeName,
			}
			gen.FieldInfos = append(gen.FieldInfos, fieldInfo)
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
			if err != nil {
				gen.FieldInfos = append(gen.FieldInfos, fieldInfo)
				continue
			}
			sp := strings.Split(dynamoTag.Value(), ",")
			if len(sp) == 1 {
				fieldInfo.DynamoTag = sp[0]
				gen.FieldInfos = append(gen.FieldInfos, fieldInfo)

				continue
			} else {
				if err := keyFieldHandle(gen, dynamoTag.Value(), name, typeName, pos); err != nil {
					return xerrors.Errorf("error in keyFieldHandle: %w", err)
				}
			}
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

	return nil
}

func keyFieldHandle(gen *generator, dynamoTag, name, typeName, pos string) error {
	sp := strings.Split(dynamoTag, ",")
	if len(sp) < 2 {
		return xerrors.Errorf("%s: key field for dynamo should have dynamo:\"xxx,hash\" tag", pos)
	}
	switch sp[1] {
	case "hash":
		gen.HashKeyFieldName = name
		gen.HashKeyFieldType = typeName

		if gen.HashKeyFieldType != typeInt &&
			gen.HashKeyFieldType != typeInt64 &&
			gen.HashKeyFieldType != typeString {
			return xerrors.Errorf("%s: supported key types are int, int64, string", pos)
		}

		gen.HashKeyValueName = name
		if sp[0] == "" {
			gen.HashKeyFieldTagName = name
		} else {
			gen.HashKeyFieldTagName = sp[0]
		}
	case "range":
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
		if sp[0] == "" {
			gen.RangeKeyFieldTagName = name
		} else {
			gen.RangeKeyFieldTagName = sp[0]
		}
	}

	return nil
}

func uppercaseExtraction(name string, dupMap map[string]int) (lower string) {
	for _, x := range name {
		if 65 <= x && x <= 90 {
			lower += string(x + 32)
		}
	}
	if _, ok := dupMap[lower]; !ok {
		dupMap[lower] = 1
	} else {
		dupMap[lower]++
		lower = fmt.Sprintf("%s%d", lower, dupMap[lower])
	}
	return
}

func dynamoTagCheck(pos string, tags *structtag.Tags) (string, error) {
	if dynamoTag, err := tags.Get("dynamo"); err == nil {
		tag := strings.Split(dynamoTag.Value(), ",")[0]
		if tag == "" {
			return "", nil
		}
		if !valueCheck.MatchString(tag) {
			return "", xerrors.Errorf("%s: key field for `dynamo` should have other than blanks and symbols tag", pos)
		}
		if strings.Contains("0123456789", string(tag[0])) {
			return "", xerrors.Errorf("%s: key field for `dynamo` should have prefix other than numbers required", pos)
		}
		return tag, nil
	}
	return "", nil
}
