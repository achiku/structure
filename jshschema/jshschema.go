package jshschema

import (
	"encoding/json"
	"io"
	"log"
	"regexp"
	"sort"
	"strings"

	"github.com/lestrrat/go-jsref"
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

var intToWordMap = []string{
	"zero",
	"one",
	"two",
	"three",
	"four",
	"five",
	"six",
	"seven",
	"eight",
	"nine",
}

const (
	FormatDateTime Format = "date-time"
	FormatEmail    Format = "email"
	FormatHostname Format = "hostname"
	FormatIPv4     Format = "ipv4"
	FormatIPv6     Format = "ipv6"
	FormatURI      Format = "uri"
)

type ItemSpec struct {
	TupleMode bool
	Schemas   SchemaList
}
type DependencyMap struct {
	Names   map[string][]string
	Schemas map[string]*Schema
}
type PrimitiveType int
type PrimitiveTypes []PrimitiveType
type Format string

type Number struct {
	Val         float64
	Initialized bool
}

type Integer struct {
	Val         int
	Initialized bool
}

type Bool struct {
	Val         bool
	Default     bool
	Initialized bool
}

const (
	UnspecifiedType PrimitiveType = iota
	NullType
	IntegerType
	StringType
	ObjectType
	ArrayType
	BooleanType
	NumberType
)

type SchemaList []*Schema

type Schema struct {
	ID          string             `json:"id,omitempty"`
	Title       string             `json:"title,omitempty"`
	Description string             `json:"description,omitempty"`
	Default     interface{}        `json:"default,omitempty"`
	Type        []string           `json:"type,omitempty"`
	SchemaRef   string             `json:"$schema,omitempty"`
	Definitions map[string]*Schema `json:"definitions,omitempty"`
	Reference   string             `json:"$ref,omitempty"`
	Format      Format             `json:"format,omitempty"`

	// NumericValidations
	MultipleOf       Number `json:"multipleOf,omitempty"`
	Minimum          Number `json:"minimum,omitempty"`
	Maximum          Number `json:"maximum,omitempty"`
	ExclusiveMinimum Bool   `json:"exclusiveMinimum,omitempty"`
	ExclusiveMaximum Bool   `json:"exclusiveMaximum,omitempty"`

	// StringValidation
	MaxLength Integer `json:"maxLength,omitempty"`
	MinLength Integer `json:"minLength,omitempty"`

	// ArrayValidations
	AdditionalItems *AdditionalItems
	Items           *ItemSpec
	MinItems        Integer
	MaxItems        Integer
	UniqueItems     Bool

	// ObjectValidations
	MaxProperties        Integer                    `json:"maxProperties,omitempty"`
	MinProperties        Integer                    `json:"minProperties,omitempty"`
	Required             []string                   `json:"required,omitempty"`
	Dependencies         DependencyMap              `json:"dependencies,omitempty"`
	Properties           map[string]*Schema         `json:"properties,omitempty"`
	AdditionalProperties *AdditionalProperties      `json:"additionalProperties,omitempty"`
	PatternProperties    map[*regexp.Regexp]*Schema `json:"patternProperties,omitempty"`

	Enum  []interface{} `json:"enum,omitempty"`
	AllOf SchemaList    `json:"allOf,omitempty"`
	AnyOf SchemaList    `json:"anyOf,omitempty"`
	OneOf SchemaList    `json:"oneOf,omitempty"`
	Not   *Schema       `json:"not,omitempty"`
}

type AdditionalItems struct {
	*Schema
}

type AdditionalProperties struct {
	*Schema
}

// Generate creates struct
func Generate(input io.Reader, pkgName string) ([]byte, error) {
	var schema Schema
	if err := json.NewDecoder(input).Decode(&schema); err != nil {
		return nil, err
	}
	for k, v := range schema.Properties {
		log.Println(k)
		log.Println(v.Reference)
		res := jsref.New()
		tmp, err := res.Resolve(schema, v.Reference)
		if err != nil {
			log.Println(err)
		}
		a := tmp.(*Schema)
		for col, def := range a.Definitions {
			log.Println(col)
			log.Println(def)
		}
	}
	return nil, nil
}

func generateTypes(obj map[string]interface{}, depth int) string {
	keys := make([]string, 0, len(obj))
	for key := range obj {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := obj[key]

		//If a nested value, recurse
		switch value := value.(type) {
		case []map[string]interface{}:
			generateTypes(value[0], depth+1)
		case map[string]interface{}:
			generateTypes(value, depth+1)
		}

		res := jsref.New()
		switch value := value.(type) {
		case string:
			if key == "$ref" {
				s, err := res.Resolve(obj, value)
				if err != nil {
					log.Print(err)
				}
				log.Printf("%v", s)
			}
			log.Println(strings.Repeat(" ", depth) + key + "->" + value)
		case map[string]interface{}:
			log.Println(strings.Repeat(" ", depth) + key)
		}
	}
	return ""
}
