package main

import (
	"log"
	"strings"
	"unicode"
)

func ToCamelCase(s string) string {
	upNext := false
	length := 0
	return strings.Map(func(r rune) rune {
		if r == '_' || r == '-' || r == ' ' {
			upNext = length > 0
			return -1
		}
		if upNext {
			r = unicode.ToUpper(r)
		} else {
			r = unicode.ToLower(r)
		}

		length++
		upNext = false

		return r
	}, s)
}

func ToPascalCase(s string) string {
	upNext := false
	length := 0
	return strings.Map(func(r rune) rune {
		if r == '_' || r == '-' || r == ' ' {
			upNext = true
			return -1
		}
		if upNext || length == 0 {
			r = unicode.ToUpper(r)
		} else {
			r = unicode.ToLower(r)
		}

		length++
		upNext = false

		return r
	}, s)
}

func ToSnakeCase(sep string, s string) string {
	rsep := []rune(sep)
	out := make([]rune, 0, len(s))
	lastLower := false
	lastLetter := false
	for _, r := range s {
		if lastLower = unicode.IsLower(r); lastLower || r == '_' || r == '-' || r == ' ' {
			if !lastLower {
				if lastLetter {
					out = append(out, rsep...)
				}
				goto skip
			}
			goto next
		} else if unicode.IsUpper(r) {
			if lastLower {
				log.Println(r, "lastLower + upper(r)")
				out = append(out, rsep...)
			}
			r = unicode.ToLower(r)
		}
		lastLower = false

	next:
		out = append(out, r)
	skip:
		lastLower = false
		lastLetter = unicode.IsLetter(r)
	}
	return string(out)
}
