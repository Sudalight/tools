package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"go/types"
	"io/fs"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

var (
	debug     = flag.Bool("debug", false, "debug mode")
	rootDir   = flag.String("dir", ".", "directories to read and write files")
	pkgPrefix = flag.String("rpp", ".", "redirect pkg prefix")
)

// Generator holds the state of the analysis. Primarily used to buffer
// the output for format.Source.
type Generator struct {
	buf  bytes.Buffer // Accumulated output for each type part.
	pkgs []*Package   // Package we are scanning.
}

type Package struct {
	name  string
	defs  map[*ast.Ident]types.Object
	files []*File
}

// File holds a single parsed file and associated data.
type File struct {
	pkg  *Package  // Package to which this file belongs.
	file *ast.File // Parsed AST.
	// These fields are reset for each type being generated.
	typeNames []string // Variable names defined.
}

func main() {
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Llongfile)
	if !*debug {
		log.SetFlags(0)
		log.SetPrefix("redirect: ")
	}
	filepath.WalkDir(*rootDir, walkDir)
}

func walkDir(path string, d fs.DirEntry, err error) error {
	if !d.IsDir() {
		return nil
	}
	log.Println(path)
	if !strings.HasPrefix(path, "./") {
		path = "./" + path
	}
	g := Generator{}
	g.parsePackage(path)
	for i := range g.pkgs {
		g.buf = bytes.Buffer{}

		g.Write("package %s\n\n", g.pkgs[i].name)

		g.Write("import \"%s\"\n\n", *pkgPrefix+strings.TrimPrefix(path, *rootDir))

		g.generate(i)
		// Format the output.
		src := g.format()

		// Write to file.
		baseName := "models.go"
		outputName := path + "/" + strings.ToLower(baseName)
		err := ioutil.WriteFile(outputName, src, 0644)
		if err != nil {
			log.Fatalf("writing output: %s", err)
		}
	}
	return nil
}

func (g *Generator) generate(index int) {
	for i := range g.pkgs[index].files {
		file := g.pkgs[index].files[i]
		ast.Inspect(file.file, file.genDecl)

		if len(file.typeNames) == 0 {
			continue
		}

		for j := range file.typeNames {
			g.Write("// Deprecated, use %s instead.\n", *pkgPrefix)
			g.Write("type %s = %s.%s\n\n", file.typeNames[j], g.pkgs[index].name, file.typeNames[j])
		}
	}
}

// parsePackage analyzes the single package constructed from the patterns and tags.
// parsePackage exits if there is an error.
func (g *Generator) parsePackage(path string) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles |
			packages.NeedCompiledGoFiles | packages.NeedImports |
			packages.NeedTypes | packages.NeedTypesSizes |
			packages.NeedSyntax | packages.NeedTypesInfo,
		Tests: false,
	}
	pkgs, err := packages.Load(cfg, path)
	if err != nil {
		log.Fatal(err)
	}
	if len(pkgs) < 1 {
		log.Fatal("error: no packages found")
	}

	for i := range pkgs {
		if len(pkgs[i].GoFiles) < 1 {
			continue
		}
		g.addPackage(pkgs[i])
	}
}

// addPackage adds a type checked Package and its syntax files to the generator.
func (g *Generator) addPackage(pkg *packages.Package) {
	p := &Package{
		name:  pkg.Name,
		defs:  pkg.TypesInfo.Defs,
		files: make([]*File, len(pkg.Syntax)),
	}

	for i, file := range pkg.Syntax {
		p.files[i] = &File{
			file: file,
			pkg:  p,
		}
	}
	g.pkgs = append(g.pkgs, p)
}

func (f *File) genDecl(node ast.Node) bool {
	decl, ok := node.(*ast.GenDecl)
	if !ok || decl.Tok != token.TYPE {
		// We only care about TYPE declarations.
		return true
	}
	for _, spec := range decl.Specs {
		f.typeNames = append(f.typeNames, spec.(*ast.TypeSpec).Name.Obj.Name)
	}
	return false
}

func (g *Generator) Write(format string, args ...interface{}) {
	fmt.Fprintf(&g.buf, format, args...)
}

// format returns the gofmt-ed contents of the Generator's buffer.
func (g *Generator) format() []byte {
	src, err := format.Source(g.buf.Bytes())
	if err != nil {
		// Should never happen, but can arise when developing this code.
		// The user can compile the output to see the error.
		log.Printf("warning: internal error: invalid Go generated: %s", err)
		log.Printf("warning: compile the package to analyze the error")
		return g.buf.Bytes()
	}
	return src
}
