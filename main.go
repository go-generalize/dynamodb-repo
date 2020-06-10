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
	"golang.org/x/xerrors"
)

func main() {
	var (
		prefix = flag.String("prefix", "", "Prefix for table name")
	)
	flag.Parse()

	l := flag.NArg()

	if l < 1 {
		fmt.Println("You have to specify the struct name of target")
		os.Exit(1)
	}

	if err := run(flag.Arg(0), *prefix); err != nil {
		log.Fatal(err.Error())
	}
}

func run(structName, prefix string) error {
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
	for _, field := range structType.Fields.List {
		// structの各fieldを調査
		if len(field.Names) != 1 {
			return xerrors.New("`field.Names` must have only one element")
		}
		name := field.Names[0].Name

		pos := fs.Position(field.Pos()).String()

		typeName := getTypeName(field.Type)

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

			if err := dynamoTagCheck(pos, sp[0]); err != nil {
				return xerrors.Errorf("tag validator failed: %w", err)
			}
			if len(sp) == 1 {
				if sp[0] != "" {
					fieldInfo.DynamoTag = sp[0]
				}
				gen.FieldInfos = append(gen.FieldInfos, fieldInfo)
				continue
			}
			if err := keyFieldHandle(gen, sp[0], sp[1], name, typeName, pos); err != nil {
				return xerrors.Errorf("error in keyFieldHandle: %w", err)
			}
			if gen.HashKeyFieldName != "" {
				if _, err := tags.Get("auto"); err == nil {
					gen.AutoGeneration = true
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

func keyFieldHandle(gen *generator, label, keyKind, name, typeName, pos string) error {
	switch keyKind {
	case "hash":
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
		if label == "" {
			gen.RangeKeyFieldTagName = name
		} else {
			gen.RangeKeyFieldTagName = label
		}
	}

	return nil
}
