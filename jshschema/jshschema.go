package jshschema

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"unicode"

	"github.com/lestrrat/go-jsschema"
)

// commonInitialisms is a set of common initialisms.
// Only add entries that are highly unlikely to be non-initialisms.
// For instance, "ID" is fine (Freudian code is rare), but "AND" is not.
var commonInitialisms = map[string]bool{
	"API":   true,
	"ASCII": true,
	"CPU":   true,
	"CSS":   true,
	"DNS":   true,
	"EOF":   true,
	"GUID":  true,
	"HTML":  true,
	"HTTP":  true,
	"HTTPS": true,
	"ID":    true,
	"IP":    true,
	"JSON":  true,
	"LHS":   true,
	"QPS":   true,
	"RAM":   true,
	"RHS":   true,
	"RPC":   true,
	"SLA":   true,
	"SMTP":  true,
	"SSH":   true,
	"TLS":   true,
	"TTL":   true,
	"UI":    true,
	"UID":   true,
	"UUID":  true,
	"URI":   true,
	"URL":   true,
	"UTF8":  true,
	"VM":    true,
	"XML":   true,
}

// Structure represents Go struct
type Structure struct {
	Name         string
	Depth        int
	Fields       map[string]string
	ChildStructs []*Structure
}

// String formatted struct
func (s *Structure) String(isTopLevel bool) string {
	var str string
	if isTopLevel {
		str = fmt.Sprintf("type")
	}
	str = fmt.Sprintf("%s %s struct {\n", str, s.Name)
	for f, t := range s.Fields {
		str = fmt.Sprintf("%s %s%s %s\n", str, strings.Repeat(" ", s.Depth*2), f, t)
	}
	for _, c := range s.ChildStructs {
		str = fmt.Sprintf("%s %s", str, c.String(false))
	}
	str = fmt.Sprintf("%s}\n\n ", str)
	return str
}

// Parse parses JSON Schema and returns Structure struct
func Parse(schemaPath string) ([]*Structure, error) {
	s, err := schema.ReadFile(schemaPath)
	if err != nil {
		return nil, err
	}
	var structures []*Structure
	for structName, def := range s.Properties {
		structure, err := generateFields(def, s, 0)
		if err != nil {
			return nil, err
		}
		structure.Name = fmtFieldName(structName)
		structures = append(structures, structure)
	}
	return structures, nil
}

func generateFields(s *schema.Schema, root *schema.Schema, depth int) (*Structure, error) {
	depth = depth + 1
	if depth > 10 {
		return nil, errors.New("the number of recursion exceeds 10")
	}
	st := &Structure{
		Depth:  depth,
		Fields: map[string]string{},
	}
	res, err := s.Resolve(nil)
	if err != nil {
		log.Fatal(err)
	}
	for fieldName, def := range res.Definitions {
		if def.Reference != "" {
			r, err := def.Resolve(nil)
			if err != nil {
				log.Fatal(err)
			}
			st.Fields[fmtFieldName(fieldName)] = typeForValue(r.Type)[0]
		} else if containsInt(schema.ObjectType, def.Type) {
			childStruct, err := generateFields(def, root, depth)
			childStruct.Name = fmtFieldName(fieldName)
			if err != nil {
				return nil, err
			}
			st.ChildStructs = append(st.ChildStructs, childStruct)
		} else {
			st.Fields[fmtFieldName(fieldName)] = typeForValue(def.Type)[0]
		}
	}

	for fieldName, def := range res.Properties {
		if def.Reference != "" {
			r, err := def.Resolve(nil)
			if err != nil {
				log.Fatal(err)
			}
			st.Fields[fmtFieldName(fieldName)] = typeForValue(r.Type)[0]
		} else if containsInt(schema.ObjectType, def.Type) {
			childStruct, err := generateFields(def, root, depth)
			childStruct.Name = fmtFieldName(fieldName)
			if err != nil {
				return nil, err
			}
			st.ChildStructs = append(st.ChildStructs, childStruct)
		} else {
			st.Fields[fmtFieldName(fieldName)] = typeForValue(def.Type)[0]
		}
	}
	return st, nil
}

func typeForValue(types schema.PrimitiveTypes) []string {
	var goTypes []string
	for _, t := range types {
		switch t {
		case schema.StringType:
			goTypes = append(goTypes, "string")
		case schema.IntegerType:
			goTypes = append(goTypes, "int")
		case schema.NumberType:
			goTypes = append(goTypes, "int")
		case schema.BooleanType:
			goTypes = append(goTypes, "bool")
		case schema.ObjectType:
			goTypes = append(goTypes, "struct")
		case schema.ArrayType:
			goTypes = append(goTypes, "struct")
		}
	}
	return goTypes
}

func containsInt(s schema.PrimitiveType, l schema.PrimitiveTypes) bool {
	for _, i := range l {
		if i == s {
			return true
		}
	}
	return false
}

// fmtFieldName formats a string as a struct key
//
// Example:
// 	fmtFieldName("foo_id")
// Output: FooID
func fmtFieldName(s string) string {
	name := lintFieldName(s)
	runes := []rune(name)
	for i, c := range runes {
		ok := unicode.IsLetter(c) || unicode.IsDigit(c)
		if i == 0 {
			ok = unicode.IsLetter(c)
		}
		if !ok {
			runes[i] = '_'
		}
	}
	return string(runes)
}

func lintFieldName(name string) string {
	// Fast path for simple cases: "_" and all lowercase.
	if name == "_" {
		return name
	}

	for len(name) > 0 && name[0] == '_' {
		name = name[1:]
	}

	allLower := true
	for _, r := range name {
		if !unicode.IsLower(r) {
			allLower = false
			break
		}
	}
	if allLower {
		runes := []rune(name)
		if u := strings.ToUpper(name); commonInitialisms[u] {
			copy(runes[0:], []rune(u))
		} else {
			runes[0] = unicode.ToUpper(runes[0])
		}
		return string(runes)
	}

	// Split camelCase at any lower->upper transition, and split on underscores.
	// Check each word for common initialisms.
	runes := []rune(name)
	w, i := 0, 0 // index of start of word, scan
	for i+1 <= len(runes) {
		eow := false // whether we hit the end of a word

		if i+1 == len(runes) {
			eow = true
		} else if runes[i+1] == '_' {
			// underscore; shift the remainder forward over any run of underscores
			eow = true
			n := 1
			for i+n+1 < len(runes) && runes[i+n+1] == '_' {
				n++
			}

			// Leave at most one underscore if the underscore is between two digits
			if i+n+1 < len(runes) && unicode.IsDigit(runes[i]) && unicode.IsDigit(runes[i+n+1]) {
				n--
			}

			copy(runes[i+1:], runes[i+n+1:])
			runes = runes[:len(runes)-n]
		} else if unicode.IsLower(runes[i]) && !unicode.IsLower(runes[i+1]) {
			// lower->non-lower
			eow = true
		}
		i++
		if !eow {
			continue
		}

		// [w,i) is a word.
		word := string(runes[w:i])
		if u := strings.ToUpper(word); commonInitialisms[u] {
			// All the common initialisms are ASCII,
			// so we can replace the bytes exactly.
			copy(runes[w:], []rune(u))

		} else if strings.ToLower(word) == word {
			// already all lowercase, and not the first word, so uppercase the first character.
			runes[w] = unicode.ToUpper(runes[w])
		}
		w = i
	}
	return string(runes)
}
