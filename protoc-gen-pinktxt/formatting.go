package main

import (
	"encoding/json"
	"errors"
	"log"
	"path"
	"regexp"
	"strings"
	"text/template"

	desc "github.com/nilium/pinktxt/internal/plugin/google/protobuf"
)

var defaultTemplateFuncs = template.FuncMap{
	"error": errors.New,

	"nl": func() string { return "\n" },

	"rmprefix": func(prefix, s string) string { return strings.TrimPrefix(s, prefix) },
	"rmsuffix": func(suffix, s string) string { return strings.TrimSuffix(s, suffix) },
	"trim":     func(cuts, s string) string { return strings.Trim(s, cuts) },
	"trimr":    func(cuts, s string) string { return strings.TrimLeft(s, cuts) },
	"triml":    func(cuts, s string) string { return strings.TrimRight(s, cuts) },
	"trimws":   strings.TrimSpace,
	"repeat":   func(count int, s string) string { return strings.Repeat(s, count) },

	"camelcase":  ToCamelCase,
	"pascalcase": ToPascalCase,
	"snakecase":  ToSnakeCase,

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
}

func copyDefaultTemplateFuncs(dst template.FuncMap) template.FuncMap {
	if dst == nil {
		dst = make(template.FuncMap, len(defaultTemplateFuncs))
	}

	for k, f := range defaultTemplateFuncs {
		dst[k] = f
	}

	return dst
}
