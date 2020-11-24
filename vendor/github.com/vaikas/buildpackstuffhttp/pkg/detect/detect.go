package detect

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	//	"unicode"
)

type Function struct {
	File   string // full path to the file
	Source string // full source file
}

const (
	httpImport = `"net/http"`
)

type paramType int

const (
	notSuportedType paramType = iota
	responseWriterType
	ptrRequestType
)

type functionSignature struct {
	in  []paramType
	out []paramType
}

// Valid function signatures are like so (defined in: github.com/cloudevents/sdk-go/v2/client/receiver.go):
// * func(http.ResponseWriter, *http.Request)
var validFunctions = map[string]functionSignature{
	"func(http.ResponseWriter, *http.Request)": functionSignature{in: []paramType{responseWriterType, ptrRequestType}, out: []paramType{}},
}

// imports keeps track of which files that we care about are imported as which
// local imports. For example if you import:
// 	nethttp "net/http"
// localHTTPImport would be set to nethttp
type imports struct {
	localHTTPImport string
}

type FunctionDetails struct {
	Name      string
	Package   string
	Signature string
}

func ReadAndCheckFile(filename string) *FunctionDetails {
	file, err := os.Open(filename)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer file.Close()

	// read the whole file in
	srcbuf, err := ioutil.ReadAll(file)
	if err != nil {
		log.Println(err)
		return nil
	}
	return CheckFile(&Function{File: filename, Source: string(srcbuf)})
}

func CheckFile(f *Function) *FunctionDetails {
	// file set
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, f.File, f.Source, 0)
	if err != nil {
		log.Println(err)
		return nil
	}

	c := imports{}
	var functionName = ""
	var signature = ""

	// main inspection
	ast.Inspect(astFile, func(n ast.Node) bool {
		switch fn := n.(type) {
		// Check if the file imports the cloud events SDK that we're expecting
		case *ast.File:
			for _, i := range fn.Imports {
				if i.Path.Value == httpImport {
					if i.Name == nil {
						c.localHTTPImport = "http"
					} else {
						c.localHTTPImport = i.Name.String()
					}
				}
			}

			for _, d := range fn.Decls {
				if f, ok := d.(*ast.FuncDecl); ok {
					functionName = f.Name.Name
					if f.Recv != nil {
						fmt.Println("Found receiver ", f.Recv)
					}
					if sig := checkFunction(c, f.Type); sig != "" {
						signature = sig
					}
				}
			}
		}
		return true
	})
	if signature != "" && functionName != "" {
		return &FunctionDetails{Name: functionName, Signature: signature}
	}
	return nil
}

// checkFunction takes a function signature and returns a friendly (string)
// representation of the supported function or "" if the function signature
// is not supported.
// For example
// func Receive(http.ResponseWriter, *http.Request) {
// would return:
// func(http.ResponseWriter, *http.Request)
func checkFunction(c imports, f *ast.FuncType) string {
	fs := functionSignature{}
	if f == nil {
		return ""
	}
	if f.Params != nil {
		for _, p := range f.Params.List {
			t := typeToParamType(c, p.Type)
			fs.in = append(fs.in, t)
		}
	}
	if f.Results != nil {
		for _, r := range f.Results.List {
			t := typeToParamType(c, r.Type)
			fs.out = append(fs.out, t)
		}
	}

	for k, v := range validFunctions {
		if len(fs.in) == len(v.in) && len(fs.out) == len(v.out) {
			match := true
			for i := range fs.in {
				if fs.in[i] != v.in[i] {
					match = false
					continue
				}
			}
			for i := range fs.out {
				if fs.out[i] != v.out[i] {
					match = false
					continue
				}
			}
			if match {
				return k
			}
		}
	}
	return ""
}

// typeToParamType will take import paths and an expression and try to map
// it to know paramType. If supported paramType is not found, will return
// notSupportedType
func typeToParamType(c imports, e ast.Expr) paramType {
	switch e := e.(type) {
	// Check if pointer to Event
	case *ast.StarExpr:
		if s, ok := e.X.(*ast.SelectorExpr); ok {
			// We only support pointer to Event
			if s.Sel.String() == "Request" {
				if im, ok := s.X.(*ast.Ident); ok {
					if im.Name == c.localHTTPImport {
						return ptrRequestType
					}
				}
			}
		}
	case *ast.SelectorExpr:
		if e.Sel.String() == "ResponseWriter" {
			if im, ok := e.X.(*ast.Ident); ok {
				if c.localHTTPImport != "" && im.Name == c.localHTTPImport {
					return responseWriterType
				}
			}
		}
	}
	return notSuportedType
}
