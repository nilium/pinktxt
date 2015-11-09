package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/gogo/protobuf/proto"

	desc "github.com/nilium/pinktxt/internal/plugin/google/protobuf"
	compiler "github.com/nilium/pinktxt/internal/plugin/google/protobuf/compiler"
)

var _ = (*compiler.CodeGeneratorRequest)(nil)
var _ = desc.FieldDescriptorProto_TYPE_DOUBLE

func heapString(s string) *string {
	return &s
}

type Template interface {
	Execute(io.Writer, string, interface{}) error
	ExecuteTemplate(io.Writer, string, interface{}) error
}

var files map[string]*compiler.CodeGeneratorResponse_File

type Params map[string][]string

func (p Params) Int(key string, def int) int {
	var ok bool
	var vals []string
	if vals, ok = p[key]; !ok || len(vals) == 0 {
		return def
	}

	if num, err := strconv.Atoi(vals[0]); err == nil {
		return num
	}

	return def
}

func (p Params) Bool(key string, def bool) bool {
	var ok bool
	var vals []string
	if vals, ok = p[key]; !ok || len(vals) == 0 {
		return def
	}
	v := strings.ToLower(vals[0])
	return v == "yes" || v == "true"
}

func (p Params) Get(key string) string {
	v := p[key]
	if len(v) > 0 {
		return v[0]
	}
	return ""
}

func splitQuotedOn(splitters ...rune) func(rune) bool {
	shouldSplit := make(map[rune]struct{}, len(splitters))
	for _, r := range splitters {
		shouldSplit[r] = struct{}{}
	}

	inquote, escape := false, false
	return func(r rune) bool {
		if escape {
			escape = false
			return false
		}

		if !escape && r == '"' {
			inquote = !inquote
		}

		if inquote {
			escape = r == '\\'
			return false
		}

		_, ok := shouldSplit[r]
		return ok
	}
}

func parseParameters(params string) Params {
	pairs := strings.FieldsFunc(params, splitQuotedOn(';'))
	r := make(map[string][]string, len(pairs))
	for _, pair := range pairs {
		values := strings.FieldsFunc(pair, splitQuotedOn('=', ','))
		key := values[0]
		if len(pair) == 1 {
			r[key] = append(r[key], "")
			continue
		}
		for i := 1; i < len(values); i++ {
			val := values[i]
			if len(val) > 0 && val[0] == '"' {
				var err error
				val, err = strconv.Unquote(val)
				if err != nil {
					panic(err)
				}
			}

			r[key] = append(r[key], val)
		}
	}
	return Params(r)
}

type typeFinder struct {
	Request *compiler.CodeGeneratorRequest
}

func (t typeFinder) req() *compiler.CodeGeneratorRequest {
	return t.Request
}

func (t typeFinder) findEnum(in []*desc.EnumDescriptorProto, name string) *desc.EnumDescriptorProto {
	for _, e := range in {
		if "."+e.GetName() == name {
			return e
		}
	}
	return nil
}

func (t typeFinder) findExtension(in []*desc.FieldDescriptorProto, name string) *desc.FieldDescriptorProto {
	for _, e := range in {
		if "."+e.GetName() == name {
			return e
		}
	}
	return nil
}

func (t typeFinder) findType(in *desc.DescriptorProto, name string) interface{} {
	root := "." + in.GetName()
	if root == name {
		return in
	}

	prefix := root + "."
	if !strings.HasPrefix(prefix, name) {
		return nil
	}

	name = strings.TrimPrefix(name, root)
	for _, in := range in.GetNestedType() {
		if m := t.findType(in, name); m != nil {
			return m
		}
	}

	if e := t.findEnum(in.GetEnumType(), name); e != nil {
		return e
	}

	if e := t.findEnum(in.GetEnumType(), name); e != nil {
		return e
	}

	if e := t.findExtension(in.GetExtension(), name); e != nil {
		return e
	}

	return nil
}

func (t typeFinder) Find(name string) interface{} {
	if name == "" || name[0] != '.' {
		return nil
	}

	req := t.req()
	var pkg *desc.FileDescriptorProto
	for _, p := range req.GetProtoFile() {
		if p.Package == nil {
			continue
		}

		root := "." + p.GetPackage()
		prefix := root + "."

		if strings.HasPrefix(name, prefix) {
			pkg = p

			break
		}

		if name == root {
			return p
		}
	}

	if pkg == nil {
		return nil
	}

	for _, m := range pkg.GetMessageType() {
		if found := t.findType(m, name); found != nil {
			return found
		}
	}

	if e := t.findEnum(pkg.GetEnumType(), name); e != nil {
		return e
	}

	if e := t.findExtension(pkg.GetExtension(), name); e != nil {
		return e
	}

	// TODO: Services
	// TODO: Groups

	return nil
}

type FlatTypeRoot struct {
	Request   *compiler.CodeGeneratorRequest
	Visible   *FlatTypes
	Exported  *FlatTypes
	Params    Params
	HasData   bool
	ExecParam interface{}
}

type FlatTypes struct {
	Files      map[string]*desc.FileDescriptorProto
	Enums      map[string]*desc.EnumDescriptorProto
	Messages   map[string]*desc.DescriptorProto
	Extensions map[string]*desc.FieldDescriptorProto
	Services   map[string]*desc.ServiceDescriptorProto
}

func (f *FlatTypes) Package() interface{} {
	if len(f.Files) == 1 {
		for _, v := range f.Files {
			return v
		}
	} else if len(f.Files) == 0 {
		return nil
	}
	return f.Files
}

func (f *FlatTypes) populateMessageTypes(m []*desc.DescriptorProto, prefix string) {
	for _, d := range m {
		name := prefix + d.GetName()
		f.Messages[name] = d

		prefix := name + "."
		f.populateMessageTypes(d.GetNestedType(), prefix)
		f.populateEnums(d.GetEnumType(), prefix)
		f.populateExtensions(d.GetExtension(), prefix)
	}
}

func (f *FlatTypes) populateEnums(m []*desc.EnumDescriptorProto, prefix string) {
	for _, e := range m {
		f.Enums[prefix+e.GetName()] = e
	}
}

func (f *FlatTypes) populateExtensions(m []*desc.FieldDescriptorProto, prefix string) {
	for _, e := range m {
		f.Extensions[prefix+e.GetName()] = e
	}
}

func (f *FlatTypes) populateServices(m []*desc.ServiceDescriptorProto, prefix string) {
	for _, e := range m {
		f.Services[prefix+e.GetName()] = e
	}
}

func flatTypesForFile(pkg *desc.FileDescriptorProto, out *FlatTypes) *FlatTypes {
	if out == nil {
		out = &FlatTypes{
			Files:      make(map[string]*desc.FileDescriptorProto),
			Enums:      make(map[string]*desc.EnumDescriptorProto),
			Messages:   make(map[string]*desc.DescriptorProto),
			Extensions: make(map[string]*desc.FieldDescriptorProto),
			Services:   make(map[string]*desc.ServiceDescriptorProto),
		}
	}

	out.Files[pkg.GetName()] = pkg
	prefix := "." + pkg.GetPackage() + "."
	out.populateMessageTypes(pkg.GetMessageType(), prefix)
	out.populateEnums(pkg.GetEnumType(), prefix)
	out.populateExtensions(pkg.GetExtension(), prefix)
	out.populateServices(pkg.GetService(), prefix)

	return out
}

func getFlatTypes(req *compiler.CodeGeneratorRequest, exported bool, out *FlatTypes) *FlatTypes {
	include := func(string) bool { return true }
	if exported {
		include = func(name string) bool {
			for _, n := range req.GetFileToGenerate() {
				if n == name {
					return true
				}
			}
			return false
		}
	}

	for _, pkg := range req.GetProtoFile() {
		if !include(pkg.GetName()) {
			continue
		}

		out = flatTypesForFile(pkg, out)
	}

	return out
}

func main() {
	log.SetPrefix("pinktxt: ")
	log.SetFlags(0)

	var resp = new(compiler.CodeGeneratorResponse)
	defer func() {
		output, err := proto.Marshal(resp)
		if err != nil {
			log.Fatalf("Error encoding %T: %v", resp, err)
			return
		}

		if resp.Error != nil {
			log.Printf("Error in response: %s", *resp.Error)
		}

		for len(output) > 0 {
			n, err := os.Stdout.Write(output)
			if n > len(output) {
				n = len(output)
			}
			output = output[n:]

			if err != nil {
				time.Sleep(time.Millisecond * 500)
				log.Printf("Error writing output to standard out: %v", err)
			}
		}
	}()

	var req compiler.CodeGeneratorRequest
	{
		var input bytes.Buffer
		if n, err := input.ReadFrom(os.Stdin); err != nil {
			resp.Error = heapString("error reading from standard input: " + err.Error())
			return
		} else if n == 0 {
			resp.Error = heapString("no input provided")
			return
		}

		if err := proto.Unmarshal(input.Bytes(), &req); err != nil {
			resp.Error = heapString("error unmarshalling from standard input: " + err.Error())
			return
		}
	}

	params := parseParameters(req.GetParameter())
	files := make(map[string]*bytes.Buffer)
	root := FlatTypeRoot{
		Request:   &req,
		Visible:   getFlatTypes(&req, false, nil),
		Exported:  getFlatTypes(&req, true, nil),
		Params:    params,
		HasData:   false,
		ExecParam: nil,
	}

	left, right := "(*", "*)"
	if p := params.Get("left"); len(p) > 0 {
		left = p
	}
	if p := params.Get("right"); len(p) > 0 {
		right = p
	}

	// This code is all awful but at least it gets the job done right now.
	var tx *template.Template
	tx = template.New("").Delims(left, right).Funcs(mergeTypeChecks(template.FuncMap{
		"error":    errors.New,
		"find":     (typeFinder{root.Request}).Find,
		"nl":       func() string { return "\n" },
		"rmprefix": func(prefix, s string) string { return strings.TrimPrefix(s, prefix) },
		"rmsuffix": func(suffix, s string) string { return strings.TrimSuffix(s, suffix) },
		"trim":     func(cuts, s string) string { return strings.Trim(s, cuts) },
		"trimr":    func(cuts, s string) string { return strings.TrimLeft(s, cuts) },
		"triml":    func(cuts, s string) string { return strings.TrimRight(s, cuts) },
		"trimws":   strings.TrimSpace,
		"repeat":   func(count int, s string) string { return strings.Repeat(s, count) },
		"indent": func(indent string, levels int, s string) string {
			l := strings.Split(s, "\n")
			indent = strings.Repeat(indent, levels)
			for i := range l {
				if len(l[i]) > 0 {
					l[i] = indent + l[i]
				}
			}
			return strings.Join(l, "\n")
		},
		"unindent": func(indent string, levels int, s string) string {
			l := strings.Split(s, "\n")
			indent = strings.Repeat(indent, levels)
			for i := range l {
				if strings.HasPrefix(l[i], indent) {
					l[i] = strings.TrimPrefix(l[i], indent)
				}
			}
			return strings.Join(l, "\n")
		},
		"json": func(d interface{}) (string, error) {
			b, err := json.Marshal(d)
			return string(b), err
		},
		"basename": path.Base,
		"dirname":  path.Dir,
		"prettyjson": func(prefix, indent string, d interface{}) (string, error) {
			b, err := json.MarshalIndent(d, prefix, indent)
			return string(b), err
		},
		"map": func(pairs ...interface{}) map[interface{}]interface{} {
			m := make(map[interface{}]interface{})
			for i := 0; i < len(pairs); i += 2 {
				m[pairs[i]] = pairs[i+1]
			}
			return m
		},
		"exec": func(name string, dot ...interface{}) (string, error) {
			var data interface{} = dot
			if len(dot) == 1 {
				data = dot[0]
			} else if len(dot) == 0 {
				data = nil
			}

			var buf bytes.Buffer
			var err error
			if name != "" {
				err = tx.ExecuteTemplate(&buf, name, data)
			} else {
				err = tx.Execute(&buf, data)
			}

			return buf.String(), err
		},
		"fexec": func(name, outfile string, data ...interface{}) error {
			subroot := root
			if len(data) == 1 {
				d := data[0]
				if _, ok := d.(FlatTypeRoot); !ok {
					subroot.ExecParam = d
				}
			} else if len(data) > 1 {
				subroot.ExecParam = data
			}

			var out io.Writer = ioutil.Discard
			if len(name) > 0 {
				b, ok := files[outfile]
				if !ok {
					b = &bytes.Buffer{}
					files[outfile] = b
				}
				out = b
				subroot.HasData = b.Len() > 0
			}

			if name != "" {
				return tx.ExecuteTemplate(out, name, subroot)
			} else {
				return tx.Execute(out, subroot)
			}
		},
		"flatpkg": func(pkg *desc.FileDescriptorProto) *FlatTypes {
			if pkg == nil {
				return nil
			}

			return flatTypesForFile(pkg, nil)
		},
		"rxquote": regexp.QuoteMeta,
		"gsubr": func(regex, repl, subj string) (string, error) {
			rx, err := regexp.Compile(regex)
			if err != nil {
				return "", err
			}
			return rx.ReplaceAllString(subj, repl), nil
		},
		"gsubl": func(regex, repl, subj string) (string, error) {
			rx, err := regexp.Compile(regex)
			if err != nil {
				return "", err
			}
			return rx.ReplaceAllLiteralString(subj, repl), nil
		},
		"gsub": func(old, new, s string) string {
			return strings.Replace(s, old, new, -1)
		},
		"subln": func(old, new, s string, count int) string {
			return strings.Replace(s, old, new, count)
		},
		"log":   func(d ...interface{}) error { log.Print(d...); return nil },
		"logln": func(d ...interface{}) error { log.Println(d...); return nil },
		"logf":  func(f string, d ...interface{}) error { log.Printf(f, d...); return nil },
		"option": func(name string, pkg *desc.FileDescriptorProto) interface{} {
			var bits []string
			for _, v := range pkg.GetOptions().GetUninterpretedOption() {
				bits = bits[0:0]
				if strings.Join(bits, ".") != name {
					continue
				}
				switch {
				case v.IdentifierValue != nil:
					return *v.IdentifierValue
				case v.PositiveIntValue != nil:
					return *v.PositiveIntValue
				case v.NegativeIntValue != nil:
					return *v.NegativeIntValue
				case v.DoubleValue != nil:
					return *v.DoubleValue
				case v.AggregateValue != nil:
					return *v.AggregateValue
				case v.StringValue != nil:
					return string(v.StringValue)
				default:
					return nil
				}
			}
			return nil
		},
	}))

	tx, err := tx.ParseFiles(params["template"]...)
	if err != nil {
		resp.Error = heapString("error parsing template(s): " + err.Error())
		return
	}

	templates := params["template"]
	if tn := params["exec"]; len(tn) > 0 {
		templates = tn
	}

	for _, name := range templates {
		if err := tx.ExecuteTemplate(ioutil.Discard, name, root); err != nil {
			resp.Error = heapString(err.Error())
			return
		}
	}

	for name, buf := range files {
		f := &compiler.CodeGeneratorResponse_File{
			Name:    heapString(name),
			Content: heapString(buf.String()),
		}

		log.Printf("OUT=%q", name)
		resp.File = append(resp.File, f)
	}
}
