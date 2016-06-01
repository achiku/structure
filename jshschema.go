package main

import (
	"errors"
	"log"

	"github.com/lestrrat/go-jshschema"
	"github.com/lestrrat/go-jsschema"
)

// JSONParse parses JSON Schema and returns Structure struct
func JSONParse(schemaPath string) ([]*Structure, error) {
	s, err := hschema.ReadFile(schemaPath)
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

func generateFields(s *schema.Schema, root *hschema.HyperSchema, depth int) (*Structure, error) {
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
