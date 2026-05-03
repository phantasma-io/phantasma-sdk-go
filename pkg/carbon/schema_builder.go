package carbon

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

// FieldType describes one public token-schema field declaration.
type FieldType struct {
	Name string `json:"name"`
	Type VMType `json:"type"`
}

// TokenSchemasJSON is the public JSON token-schema shape shared by the SDKs.
type TokenSchemasJSON struct {
	SeriesMetadata []FieldType `json:"seriesMetadata"`
	ROM            []FieldType `json:"rom"`
	RAM            []FieldType `json:"ram"`
}

// MarshalJSON encodes the field type name in the public SDK JSON shape.
func (f FieldType) MarshalJSON() ([]byte, error) {
	type fieldTypeJSON struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}

	typeName, err := VMTypeName(f.Type)
	if err != nil {
		return nil, err
	}
	return json.Marshal(fieldTypeJSON{Name: f.Name, Type: typeName})
}

// UnmarshalJSON decodes the field type name from the public SDK JSON shape.
func (f *FieldType) UnmarshalJSON(data []byte) error {
	var raw struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if raw.Name == "" {
		return fmt.Errorf("field name must be string")
	}
	if raw.Type == "" {
		return fmt.Errorf("field type must be string")
	}

	vmType, err := VMTypeFromString(raw.Type)
	if err != nil {
		return err
	}
	f.Name = raw.Name
	f.Type = vmType
	return nil
}

// ParseTokenSchemasJSON parses the public JSON token-schema shape.
func ParseTokenSchemasJSON(data string) (TokenSchemasJSON, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(data), &raw); err != nil {
		return TokenSchemasJSON{}, err
	}

	series, err := parseTokenSchemaFieldArray(raw, "seriesMetadata")
	if err != nil {
		return TokenSchemasJSON{}, err
	}
	rom, err := parseTokenSchemaFieldArray(raw, "rom")
	if err != nil {
		return TokenSchemasJSON{}, err
	}
	ram, err := parseTokenSchemaFieldArray(raw, "ram")
	if err != nil {
		return TokenSchemasJSON{}, err
	}

	return TokenSchemasJSON{
		SeriesMetadata: series,
		ROM:            rom,
		RAM:            ram,
	}, nil
}

// TokenSchemasFromJSON builds validated token schemas from the public JSON shape.
func TokenSchemasFromJSON(data string) (TokenSchemas, error) {
	fields, err := ParseTokenSchemasJSON(data)
	if err != nil {
		return TokenSchemas{}, err
	}
	return BuildTokenSchemasFromFields(fields.SeriesMetadata, fields.ROM, fields.RAM)
}

// MustTokenSchemasFromJSON builds token schemas from JSON and panics on invalid input.
func MustTokenSchemasFromJSON(data string) TokenSchemas {
	out, err := TokenSchemasFromJSON(data)
	if err != nil {
		panic(err)
	}
	return out
}

// BuildTokenSchemasFromFields builds validated token schemas from field declarations.
func BuildTokenSchemasFromFields(seriesMetadata []FieldType, rom []FieldType, ram []FieldType) (TokenSchemas, error) {
	seriesFields, err := fieldTypesToSchemas(append(standardSeriesFieldTypes(), seriesMetadata...))
	if err != nil {
		return TokenSchemas{}, err
	}
	romFields, err := fieldTypesToSchemas(append(standardNFTFieldTypes(), rom...))
	if err != nil {
		return TokenSchemas{}, err
	}
	ramFields, err := fieldTypesToSchemas(ram)
	if err != nil {
		return TokenSchemas{}, err
	}

	schemas := TokenSchemas{
		SeriesMetadata: VMStructSchema{
			Fields: seriesFields,
		},
		ROM: VMStructSchema{
			Fields: romFields,
		},
		RAM: VMStructSchema{
			Fields: ramFields,
			Flags:  VMStructFlagsNone,
		},
	}
	if len(ram) == 0 {
		schemas.RAM.Flags = VMStructFlagsDynamicExtras
	}
	if err := VerifyTokenSchemas(schemas); err != nil {
		return TokenSchemas{}, err
	}
	return schemas, nil
}

// SerializeTokenSchemasHex serializes token schemas to uppercase hex.
func SerializeTokenSchemasHex(schemas TokenSchemas) string {
	return strings.ToUpper(hex.EncodeToString(SerializeTokenSchemas(schemas)))
}

// BuildAndSerializeTokenSchemas serializes schemas or the default schemas when schemas is nil.
func BuildAndSerializeTokenSchemas(schemas *TokenSchemas) []byte {
	if schemas == nil {
		defaults := PrepareStandardTokenSchemas(false)
		return SerializeTokenSchemas(defaults)
	}
	return SerializeTokenSchemas(*schemas)
}

// VMTypeFromString converts the public schema type name to a Carbon VM type.
func VMTypeFromString(value string) (VMType, error) {
	trimmed := strings.TrimSpace(value)
	if vmType, ok := vmTypeNameMap[trimmed]; ok {
		return vmType, nil
	}
	return 0, fmt.Errorf("unknown VM type: %s", value)
}

// VMTypeName returns the canonical public schema type name.
func VMTypeName(vmType VMType) (string, error) {
	if name, ok := vmTypeCanonicalNames[vmType]; ok {
		return name, nil
	}
	return "", fmt.Errorf("unknown VM type: %d", vmType)
}

func parseTokenSchemaFieldArray(raw map[string]json.RawMessage, key string) ([]FieldType, error) {
	body, ok := raw[key]
	if !ok {
		return nil, fmt.Errorf("%s must be an array", key)
	}
	var items []json.RawMessage
	if err := json.Unmarshal(body, &items); err != nil {
		return nil, fmt.Errorf("%s must be an array", key)
	}

	fields := make([]FieldType, len(items))
	for i, item := range items {
		if err := json.Unmarshal(item, &fields[i]); err != nil {
			return nil, fmt.Errorf("%s field %d invalid: %w", key, i, err)
		}
		field := fields[i]
		if field.Name == "" {
			return nil, fmt.Errorf("%s field name must be string", key)
		}
		if _, ok := vmTypeCanonicalNames[field.Type]; !ok {
			return nil, fmt.Errorf("unknown VM type: %d", field.Type)
		}
	}
	return fields, nil
}

func fieldTypesToSchemas(fields []FieldType) ([]VMNamedVariableSchema, error) {
	out := make([]VMNamedVariableSchema, len(fields))
	for i, field := range fields {
		schema, err := NewVMNamedVariableSchema(field.Name, field.Type)
		if err != nil {
			return nil, err
		}
		out[i] = schema
	}
	return out, nil
}

func standardSeriesFieldTypes() []FieldType {
	return schemaFieldsToFieldTypes(standardSeriesFields)
}

func standardNFTFieldTypes() []FieldType {
	return schemaFieldsToFieldTypes(standardNFTFields)
}

func schemaFieldsToFieldTypes(fields []VMNamedVariableSchema) []FieldType {
	out := make([]FieldType, len(fields))
	for i, field := range fields {
		out[i] = FieldType{Name: field.Name.String(), Type: field.Schema.Type}
	}
	return out
}

var vmTypeNameMap = map[string]VMType{
	"Dynamic":       VMTypeDynamic,
	"Array":         VMTypeArray,
	"Bytes":         VMTypeBytes,
	"Struct":        VMTypeStruct,
	"Int8":          VMTypeInt8,
	"Int16":         VMTypeInt16,
	"Int32":         VMTypeInt32,
	"Int64":         VMTypeInt64,
	"Int256":        VMTypeInt256,
	"Bytes16":       VMTypeBytes16,
	"Bytes32":       VMTypeBytes32,
	"Bytes64":       VMTypeBytes64,
	"String":        VMTypeString,
	"Array_Dynamic": VMTypeArrayDynamic,
	"Array_Bytes":   VMTypeArrayBytes,
	"Array_Struct":  VMTypeArrayStruct,
	"Array_Int8":    VMTypeArrayInt8,
	"Array_Int16":   VMTypeArrayInt16,
	"Array_Int32":   VMTypeArrayInt32,
	"Array_Int64":   VMTypeArrayInt64,
	"Array_Int256":  VMTypeArrayInt256,
	"Array_Bytes16": VMTypeArrayBytes16,
	"Array_Bytes32": VMTypeArrayBytes32,
	"Array_Bytes64": VMTypeArrayBytes64,
	"Array_String":  VMTypeArrayString,
	"ArrayDynamic":  VMTypeArrayDynamic,
	"ArrayBytes":    VMTypeArrayBytes,
	"ArrayStruct":   VMTypeArrayStruct,
	"ArrayInt8":     VMTypeArrayInt8,
	"ArrayInt16":    VMTypeArrayInt16,
	"ArrayInt32":    VMTypeArrayInt32,
	"ArrayInt64":    VMTypeArrayInt64,
	"ArrayInt256":   VMTypeArrayInt256,
	"ArrayBytes16":  VMTypeArrayBytes16,
	"ArrayBytes32":  VMTypeArrayBytes32,
	"ArrayBytes64":  VMTypeArrayBytes64,
	"ArrayString":   VMTypeArrayString,
}

var vmTypeCanonicalNames = map[VMType]string{
	VMTypeDynamic:      "Dynamic",
	VMTypeArray:        "Array",
	VMTypeBytes:        "Bytes",
	VMTypeStruct:       "Struct",
	VMTypeInt8:         "Int8",
	VMTypeInt16:        "Int16",
	VMTypeInt32:        "Int32",
	VMTypeInt64:        "Int64",
	VMTypeInt256:       "Int256",
	VMTypeBytes16:      "Bytes16",
	VMTypeBytes32:      "Bytes32",
	VMTypeBytes64:      "Bytes64",
	VMTypeString:       "String",
	VMTypeArrayBytes:   "Array_Bytes",
	VMTypeArrayStruct:  "Array_Struct",
	VMTypeArrayInt8:    "Array_Int8",
	VMTypeArrayInt16:   "Array_Int16",
	VMTypeArrayInt32:   "Array_Int32",
	VMTypeArrayInt64:   "Array_Int64",
	VMTypeArrayInt256:  "Array_Int256",
	VMTypeArrayBytes16: "Array_Bytes16",
	VMTypeArrayBytes32: "Array_Bytes32",
	VMTypeArrayBytes64: "Array_Bytes64",
	VMTypeArrayString:  "Array_String",
}
