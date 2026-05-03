package carbon

import (
	"encoding/base64"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"
)

const (
	// StandardMetaID is the reserved Carbon metadata field for Phantasma IDs.
	StandardMetaID = "_i"
)

var (
	standardSeriesFields = []VMNamedVariableSchema{
		MustVMNamedVariableSchema(StandardMetaID, VMTypeInt256),
		MustVMNamedVariableSchema("mode", VMTypeInt8),
		MustVMNamedVariableSchema("rom", VMTypeBytes),
	}
	standardNFTFields = []VMNamedVariableSchema{
		MustVMNamedVariableSchema(StandardMetaID, VMTypeInt256),
		MustVMNamedVariableSchema("rom", VMTypeBytes),
	}
	standardMetadataFields = []VMNamedVariableSchema{
		MustVMNamedVariableSchema("name", VMTypeString),
		MustVMNamedVariableSchema("description", VMTypeString),
		MustVMNamedVariableSchema("imageURL", VMTypeString),
		MustVMNamedVariableSchema("infoURL", VMTypeString),
		MustVMNamedVariableSchema("royalties", VMTypeInt32),
	}
	iconDataURIPattern = regexp.MustCompile(`(?i)^data:image/(png|jpeg|webp);base64,`)
	base64Pattern      = regexp.MustCompile(`^[A-Za-z0-9+/]+={0,2}$`)
)

// MetadataField is one typed metadata value used by token and NFT ROM builders.
type MetadataField struct {
	Name  string
	Value any
}

// NowUnixMillis returns the current Unix timestamp in milliseconds.
func NowUnixMillis() int64 {
	return time.Now().UnixMilli()
}

// PrepareStandardTokenSchemas builds the standard series, ROM and RAM schemas.
func PrepareStandardTokenSchemas(sharedMetadata bool) TokenSchemas {
	seriesFields := cloneSchemaFields(standardSeriesFields)
	if sharedMetadata {
		seriesFields = append(seriesFields, cloneSchemaFields(standardMetadataFields)...)
	}

	romFields := cloneSchemaFields(standardNFTFields)
	if !sharedMetadata {
		romFields = append(romFields, cloneSchemaFields(standardMetadataFields)...)
	}

	return TokenSchemas{
		SeriesMetadata: VMStructSchema{Fields: seriesFields},
		ROM:            VMStructSchema{Fields: romFields},
		RAM:            VMStructSchema{Flags: VMStructFlagsDynamicExtras},
	}
}

// SerializeTokenSchemas serializes token schemas using Carbon encoding.
func SerializeTokenSchemas(schemas TokenSchemas) []byte {
	return Serialize(&schemas)
}

// VerifyTokenSchemas checks that required standard metadata fields are present.
func VerifyTokenSchemas(schemas TokenSchemas) error {
	if err := assertMetadataFields([]VMStructSchema{schemas.SeriesMetadata, schemas.ROM}, standardMetadataFields); err != nil {
		return err
	}
	if err := assertMetadataFields([]VMStructSchema{schemas.SeriesMetadata}, standardSeriesFields); err != nil {
		return err
	}
	if err := assertMetadataFields([]VMStructSchema{schemas.ROM}, standardNFTFields); err != nil {
		return err
	}
	return nil
}

// BuildTokenMetadata serializes validated token-level metadata.
func BuildTokenMetadata(fields map[string]string) ([]byte, error) {
	required := []string{"name", "icon", "url", "description"}
	if len(fields) < len(required) {
		return nil, fmt.Errorf("token metadata is mandatory")
	}
	for _, field := range required {
		value, ok := fields[field]
		if !ok || strings.TrimSpace(value) == "" {
			return nil, fmt.Errorf("token metadata is missing required field: %s", field)
		}
	}
	if err := validateIconDataURI(fields["icon"]); err != nil {
		return nil, err
	}

	structure := VMDynamicStruct{}
	for name, value := range fields {
		field, err := NewVMNamedDynamicVariable(name, VMTypeString, value)
		if err != nil {
			return nil, err
		}
		structure.Fields = append(structure.Fields, field)
	}
	return serializeBlobChecked(&structure)
}

// MustBuildTokenMetadata serializes token metadata and panics on validation errors.
func MustBuildTokenMetadata(fields map[string]string) []byte {
	out, err := BuildTokenMetadata(fields)
	if err != nil {
		panic(err)
	}
	return out
}

// BuildSeriesInfo builds a series definition with standard metadata schemas.
func BuildSeriesInfo(phantasmaSeriesID *big.Int, maxMint uint32, maxSupply uint32, owner Bytes32) (SeriesInfo, error) {
	if err := requireBigInt("phantasmaSeriesID", phantasmaSeriesID); err != nil {
		return SeriesInfo{}, err
	}
	schemas := PrepareStandardTokenSchemas(false)
	metadata, err := BuildTokenSeriesMetadata(schemas.SeriesMetadata, phantasmaSeriesID, nil)
	if err != nil {
		return SeriesInfo{}, err
	}
	return SeriesInfo{
		MaxMint:   maxMint,
		MaxSupply: maxSupply,
		Owner:     owner,
		Metadata:  metadata,
		ROM:       VMStructSchema{},
		RAM:       VMStructSchema{},
	}, nil
}

// MustBuildSeriesInfo builds a series definition and panics on validation errors.
func MustBuildSeriesInfo(phantasmaSeriesID *big.Int, maxMint uint32, maxSupply uint32, owner Bytes32) SeriesInfo {
	out, err := BuildSeriesInfo(phantasmaSeriesID, maxMint, maxSupply, owner)
	if err != nil {
		panic(err)
	}
	return out
}

// BuildTokenSeriesMetadata serializes series metadata according to schema.
func BuildTokenSeriesMetadata(schema VMStructSchema, phantasmaSeriesID *big.Int, metadata []MetadataField) ([]byte, error) {
	if err := requireBigInt("phantasmaSeriesID", phantasmaSeriesID); err != nil {
		return nil, err
	}
	rom, err := metadataBytes(metadata, "rom")
	if err != nil {
		return nil, err
	}
	structure := VMDynamicStruct{}
	for _, field := range []struct {
		name   string
		vmType VMType
		value  any
	}{
		{name: StandardMetaID, vmType: VMTypeInt256, value: phantasmaSeriesID},
		{name: "mode", vmType: VMTypeInt8, value: int8(boolToInt(len(rom) > 0))},
		{name: "rom", vmType: VMTypeBytes, value: rom},
	} {
		dynamicField, err := NewVMNamedDynamicVariable(field.name, field.vmType, field.value)
		if err != nil {
			return nil, err
		}
		structure.Fields = append(structure.Fields, dynamicField)
	}
	if err := appendSchemaMetadataFields(&structure, schema, standardSeriesFields, metadata); err != nil {
		return nil, err
	}

	return writeDynamicStructWithSchema(&structure, &schema)
}

// MustBuildTokenSeriesMetadata serializes series metadata and panics on validation errors.
func MustBuildTokenSeriesMetadata(schema VMStructSchema, phantasmaSeriesID *big.Int, metadata []MetadataField) []byte {
	out, err := BuildTokenSeriesMetadata(schema, phantasmaSeriesID, metadata)
	if err != nil {
		panic(err)
	}
	return out
}

// BuildNFTRom serializes NFT ROM metadata according to schema.
func BuildNFTRom(schema VMStructSchema, phantasmaNFTID *big.Int, metadata []MetadataField) ([]byte, error) {
	if err := requireBigInt("phantasmaNFTID", phantasmaNFTID); err != nil {
		return nil, err
	}
	rom, err := metadataBytes(metadata, "rom")
	if err != nil {
		return nil, err
	}
	structure := VMDynamicStruct{}
	for _, field := range []struct {
		name   string
		vmType VMType
		value  any
	}{
		{name: StandardMetaID, vmType: VMTypeInt256, value: phantasmaNFTID},
		{name: "rom", vmType: VMTypeBytes, value: rom},
	} {
		dynamicField, err := NewVMNamedDynamicVariable(field.name, field.vmType, field.value)
		if err != nil {
			return nil, err
		}
		structure.Fields = append(structure.Fields, dynamicField)
	}
	if err := appendSchemaMetadataFields(&structure, schema, standardNFTFields, metadata); err != nil {
		return nil, err
	}

	return writeDynamicStructWithSchema(&structure, &schema)
}

// MustBuildNFTRom serializes NFT ROM metadata and panics on validation errors.
func MustBuildNFTRom(schema VMStructSchema, phantasmaNFTID *big.Int, metadata []MetadataField) []byte {
	out, err := BuildNFTRom(schema, phantasmaNFTID, metadata)
	if err != nil {
		panic(err)
	}
	return out
}

// BuildPhantasmaNFTPublicMintSchema removes chain-owned fields from an NFT ROM schema.
func BuildPhantasmaNFTPublicMintSchema(nftRomSchema VMStructSchema) VMStructSchema {
	fields := make([]VMNamedVariableSchema, 0, len(nftRomSchema.Fields))
	for _, field := range nftRomSchema.Fields {
		if isPhantasmaNFTReservedField(field.Name.String()) {
			continue
		}
		fields = append(fields, field)
	}
	return VMStructSchema{
		Fields: fields,
		Flags:  nftRomSchema.Flags,
	}
}

// BuildPhantasmaNFTRom serializes public mint ROM metadata for Phantasma NFT minting.
func BuildPhantasmaNFTRom(nftRomSchema VMStructSchema, metadata []MetadataField) ([]byte, error) {
	if metadata == nil {
		return nil, fmt.Errorf("metadata is required")
	}
	for _, field := range metadata {
		if isPhantasmaNFTReservedField(field.Name) {
			return nil, fmt.Errorf("metadata field %q is reserved for chain-owned deterministic mint fields", field.Name)
		}
	}

	publicMintSchema := BuildPhantasmaNFTPublicMintSchema(nftRomSchema)
	structure := VMDynamicStruct{}
	if err := appendSchemaMetadataFields(&structure, publicMintSchema, nil, metadata); err != nil {
		return nil, err
	}

	return writeDynamicStructWithSchema(&structure, &publicMintSchema)
}

// MustBuildPhantasmaNFTRom serializes public mint ROM metadata and panics on validation errors.
func MustBuildPhantasmaNFTRom(nftRomSchema VMStructSchema, metadata []MetadataField) []byte {
	out, err := BuildPhantasmaNFTRom(nftRomSchema, metadata)
	if err != nil {
		panic(err)
	}
	return out
}

func isPhantasmaNFTReservedField(name string) bool {
	return strings.EqualFold(name, StandardMetaID) || strings.EqualFold(name, "rom")
}

func appendSchemaMetadataFields(structure *VMDynamicStruct, schema VMStructSchema, defaults []VMNamedVariableSchema, metadata []MetadataField) error {
	for _, fieldSchema := range schema.Fields {
		if hasDefaultField(defaults, fieldSchema.Name.String()) {
			continue
		}
		field, ok, err := findMetadataField(metadata, fieldSchema.Name.String())
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("metadata field %q is mandatory", fieldSchema.Name.String())
		}
		dynamicField, err := NewVMNamedDynamicVariable(fieldSchema.Name.String(), fieldSchema.Schema.Type, field.Value)
		if err != nil {
			return err
		}
		structure.Fields = append(structure.Fields, dynamicField)
	}
	return nil
}

func findMetadataField(fields []MetadataField, name string) (MetadataField, bool, error) {
	for _, field := range fields {
		if field.Name == name {
			return field, true, nil
		}
	}
	for _, field := range fields {
		if strings.EqualFold(field.Name, name) {
			return MetadataField{}, false, fmt.Errorf("metadata field %q provided in incorrect case as %q", name, field.Name)
		}
	}
	return MetadataField{}, false, nil
}

func metadataBytes(fields []MetadataField, name string) ([]byte, error) {
	field, ok, err := findMetadataField(fields, name)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	switch value := field.Value.(type) {
	case []byte:
		return value, nil
	case string:
		data, err := decodeFlexibleHex(value)
		if err != nil {
			return nil, err
		}
		return data, nil
	default:
		return nil, fmt.Errorf("metadata field %q must be bytes or hex string", name)
	}
}

func serializeBlobChecked(blob Blob) (out []byte, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			out = nil
			err = fmt.Errorf("carbon serialization failed: %v", recovered)
		}
	}()
	return Serialize(blob), nil
}

func writeDynamicStructWithSchema(structure *VMDynamicStruct, schema *VMStructSchema) (out []byte, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			out = nil
			err = fmt.Errorf("metadata serialization failed: %v", recovered)
		}
	}()

	w := NewWriter()
	if !structure.WriteWithSchema(schema, w) {
		return nil, fmt.Errorf("metadata does not match schema")
	}
	return w.Bytes(), nil
}

func hasDefaultField(defaults []VMNamedVariableSchema, name string) bool {
	for _, field := range defaults {
		if field.Name.String() == name {
			return true
		}
	}
	return false
}

func cloneSchemaFields(fields []VMNamedVariableSchema) []VMNamedVariableSchema {
	out := make([]VMNamedVariableSchema, len(fields))
	copy(out, fields)
	return out
}

func validateIconDataURI(icon string) error {
	candidate := strings.TrimSpace(icon)
	if candidate == "" || !iconDataURIPattern.MatchString(candidate) {
		return fmt.Errorf("token metadata icon must be a base64-encoded data URI (PNG, JPEG, or WebP)")
	}
	payload := candidate[strings.Index(candidate, ",")+1:]
	if payload == "" || len(payload)%4 != 0 || !base64Pattern.MatchString(payload) {
		return fmt.Errorf("token metadata icon payload is not valid base64")
	}
	decoded, err := base64.StdEncoding.DecodeString(payload)
	if err != nil || len(decoded) == 0 {
		return fmt.Errorf("token metadata icon payload is not valid base64")
	}
	if strings.TrimRight(base64.StdEncoding.EncodeToString(decoded), "=") != strings.TrimRight(payload, "=") {
		return fmt.Errorf("token metadata icon payload is not valid base64")
	}
	return nil
}

func decodeFlexibleHex(value string) ([]byte, error) {
	value = strings.TrimPrefix(strings.TrimPrefix(value, "0x"), "0X")
	if len(value)%2 != 0 {
		value = "0" + value
	}
	out := make([]byte, len(value)/2)
	for i := 0; i < len(out); i++ {
		var b byte
		for j := 0; j < 2; j++ {
			c := value[i*2+j]
			switch {
			case c >= '0' && c <= '9':
				b = b<<4 | (c - '0')
			case c >= 'a' && c <= 'f':
				b = b<<4 | (c - 'a' + 10)
			case c >= 'A' && c <= 'F':
				b = b<<4 | (c - 'A' + 10)
			default:
				return nil, fmt.Errorf("invalid hex byte")
			}
		}
		out[i] = b
	}
	return out, nil
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func requireBigInt(name string, value *big.Int) error {
	if value == nil {
		return fmt.Errorf("%s is required", name)
	}
	return nil
}

func assertMetadataFields(schemas []VMStructSchema, fields []VMNamedVariableSchema) error {
	for _, expected := range fields {
		expectedName := expected.Name.String()
		var caseMismatch *VMNamedVariableSchema
		for schemaIndex := range schemas {
			for fieldIndex := range schemas[schemaIndex].Fields {
				actual := &schemas[schemaIndex].Fields[fieldIndex]
				actualName := actual.Name.String()
				if actualName == expectedName {
					if actual.Schema.Type != expected.Schema.Type {
						return fmt.Errorf("type mismatch for %s field", expectedName)
					}
					caseMismatch = nil
					goto found
				}
				if caseMismatch == nil && strings.EqualFold(actualName, expectedName) {
					caseMismatch = actual
				}
			}
		}
		if caseMismatch != nil {
			return fmt.Errorf("case mismatch for %s field", expectedName)
		}
		return fmt.Errorf("mandatory metadata field not found: %s", expectedName)
	found:
	}
	return nil
}
