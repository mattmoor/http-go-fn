package detect

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

type Function struct {
	File   string // full path to the file
	Source string // full source file
}

// FunctionArg describes a single argument (input or output) for a given variable
// For example a function from "net/http" that conforms to the Handler function, that looks
// like this "func(http.ResponseWriter, *http.Request)" would have two arguments like so:
// FunctionArg{ImportPath: "net/http", Name: "ResponseWriter"}
// FunctionArg{ImportPath: "net/http", Name: "Request", Pointer: true}
type FunctionArg struct {
	ImportPath string `json:"importPath,omitempty"`
	Name       string `json:"name"`
	Pointer    bool   `json:"pointer,omitempty"`
}

func (fa *FunctionArg) String() string {
	ret := ""
	if fa.Pointer {
		ret += "*"
	}
	// If there's a slash In the path, pull Out the last part of the path, otherwise use full
	// for things like "context", "fmt", etc.
	pkg := fa.ImportPath
	if strings.Contains(fa.ImportPath, "/") {
		pathPieces := strings.Split(fa.ImportPath, "/")
		pkg = pathPieces[len(pathPieces)-1]
	}
	if pkg != "" {
		ret += pkg + "." + fa.Name
	} else {
		ret += fa.Name
	}
	return ret
}

type FunctionSignature struct {
	In  []FunctionArg `json:"in,omitempty"`
	Out []FunctionArg `json:"out,omitempty"`
}

type FunctionSignatures struct {
	FunctionSignatures []FunctionSignature `json:"functionSignatures"`
}

func (fs *FunctionSignature) String() string {
	s := "func("
	for i, in := range fs.In {
		s += in.String()
		if i != len(fs.In)-1 {
			s += ", "
		}
	}
	s += ")"
	if len(fs.Out) > 0 {
		if len(fs.Out) > 1 {
			s += " ("
		}
		for i, out := range fs.Out {
			s += out.String()
			if i != len(fs.Out)-1 {
				s += ", "
			}
		}
		if len(fs.Out) > 1 {
			s += ")"
		}
	}
	return s
}

type FunctionDetails struct {
	Name      string
	Package   string
	Signature string
}

type Detector struct {
	sigs []FunctionSignature
}

func NewDetector(sigs []FunctionSignature) *Detector {
	return &Detector{sigs: sigs}
}

func NewDetectorFromFile(fileName string) (*Detector, error) {
	in, err := readFile(fileName)
	if err != nil {
		return nil, err
	}
	var fs FunctionSignatures
	if err := json.Unmarshal([]byte(in), &fs); err != nil {
		return nil, err
	}
	return &Detector{sigs: fs.FunctionSignatures}, err
}

func readFile(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// read the whole file In
	srcbuf, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(srcbuf), nil
}

func (d *Detector) ReadAndCheckFile(filename string) (*FunctionDetails, error) {
	src, err := readFile(filename)
	if err != nil {
		return nil, err
	}
	return d.CheckFile(&Function{File: filename, Source: string(src)})
}

func (d *Detector) CheckFile(f *Function) (*FunctionDetails, error) {
	// file set
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, f.File, f.Source, 0)
	if err != nil {
		return nil, err
	}

	// imports keeps track of which files that we care about are imported as which
	// local imports. For example if you import:
	// 	nethttp "net/http"
	// localImports["nethttp"] -> "net/http"
	localImports := make(map[string]string)

	var functionName = ""
	var signature = ""

	// main inspection
	ast.Inspect(astFile, func(n ast.Node) bool {
		switch fn := n.(type) {
		// Create the mapping of localimports to the full import paths
		case *ast.File:
			for _, i := range fn.Imports {
				// We need to unquote the path first since it's quoted
				impPath, err := strconv.Unquote(i.Path.Value)
				if err != nil {
					fmt.Printf("failed to unquote import path %q : %s\n", i.Path.Value, err)
					return false
				}

				if i.Name != nil {
					// There's a local import path, use that
					localImports[i.Name.String()] = impPath
				} else {
					// There isn't a local import defined, so use the last part
					// of the import path
					pathPieces := strings.Split(impPath, "/")
					localImports[pathPieces[len(pathPieces)-1]] = impPath
				}
			}
			for li, imp := range localImports {
				fmt.Printf("%q => %q\n", li, imp)
			}

			for _, decl := range fn.Decls {
				if f, ok := decl.(*ast.FuncDecl); ok {
					functionName = f.Name.Name
					if f.Recv != nil {
						fmt.Println("Found receiver ", f.Recv)
					}
					if sig := d.checkFunction(localImports, f.Type); sig != "" {
						signature = sig
					}
				}
			}
		}
		return true
	})
	if signature != "" && functionName != "" {
		return &FunctionDetails{Name: functionName, Signature: signature}, nil
	}
	return nil, nil
}

// checkFunction takes a function signature and returns a friendly (string)
// representation of the supported function or "" if the function signature
// is not supported.
// For example
// func Receive(http.ResponseWriter, *http.Request) {
// would return:
// func(http.ResponseWriter, *http.Request)
func (d *Detector) checkFunction(c map[string]string, f *ast.FuncType) string {
	fs := FunctionSignature{}
	if f == nil {
		return ""
	}
	if f.Params != nil {
		for _, p := range f.Params.List {
			t := typeToFunctionArg(c, p.Type)
			fs.In = append(fs.In, t)
		}
	}
	if f.Results != nil {
		for _, r := range f.Results.List {
			t := typeToFunctionArg(c, r.Type)
			fs.Out = append(fs.Out, t)
		}
	}

	for _, a := range fs.In {
		fmt.Printf("Input arg: %+v\n", a)
	}
	for _, a := range fs.Out {
		fmt.Printf("Output arg: %+v\n", a)
	}

	for _, v := range d.sigs {
		sig := v.String()
		fmt.Printf("Checking function signature: %q\n", sig)
		if len(fs.In) == len(v.In) && len(fs.Out) == len(v.Out) {
			match := true
			for i := range fs.In {
				if fs.In[i] != v.In[i] {
					match = false
					continue
				}
			}
			for i := range fs.Out {
				if fs.Out[i] != v.Out[i] {
					match = false
					continue
				}
			}
			if match {
				return sig
			}
		}
	}
	return ""
}

// typeToFunctionArg will take import paths and an expression and maps it to
// a FunctionArg.
func typeToFunctionArg(c map[string]string, e ast.Expr) FunctionArg {
	switch e := e.(type) {
	// Check if pointer to Event
	case *ast.StarExpr:
		if s, ok := e.X.(*ast.SelectorExpr); ok {
			if im, ok := s.X.(*ast.Ident); ok {
				return FunctionArg{ImportPath: c[im.Name], Name: s.Sel.String(), Pointer: true}
			}
		}
	case *ast.SelectorExpr:
		if im, ok := e.X.(*ast.Ident); ok {
			return FunctionArg{ImportPath: c[im.Name], Name: e.Sel.String()}
		}
	case *ast.Ident:
		// Built In... Or something else?
		return FunctionArg{Name: e.Name}
	}
	return FunctionArg{}
}
