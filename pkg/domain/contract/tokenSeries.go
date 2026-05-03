package contract

import (
	"math/big"

	"github.com/phantasma-io/phantasma-go/pkg/io"
)

type TokenSeriesMode uint

const (
	Unique     TokenSeriesMode = 1
	Duplicated TokenSeriesMode = 2
)

type TokenSeries struct {
	MintCount *big.Int
	MaxSupply *big.Int
	Mode      TokenSeriesMode
	Script    []byte
	ABI       ContractInterface
	ROM       []byte

	// Extra fields, not serializable
	SeriesID *big.Int
	Symbol   string
}

func (s *TokenSeries) Serialize(writer *io.BinWriter) {
	writer.WriteBigInteger(s.MintCount)
	writer.WriteBigInteger(s.MaxSupply)
	writer.WriteB(byte(s.Mode))
	writer.WriteVarBytes(s.Script)

	// Chain storage wraps the ABI bytes in a var-bytes field.
	bytes := io.Serialize[*ContractInterface](&s.ABI)
	writer.WriteVarBytes(bytes)

	writer.WriteVarBytes(s.ROM)
}

func (s *TokenSeries) Deserialize(reader *io.BinReader) {
	s.MintCount = reader.ReadBigInteger()
	s.MaxSupply = reader.ReadBigInteger()
	s.Mode = TokenSeriesMode(reader.ReadB())
	s.Script = reader.ReadVarBytes()

	// Chain storage wraps the ABI bytes in a var-bytes field.
	bytes := reader.ReadVarBytes()
	s.ABI = *io.Deserialize[*ContractInterface](bytes)

	s.ROM = reader.ReadVarBytes()
}

type TokenSeries_S struct {
	MintCount string
	MaxSupply string
	Mode      TokenSeriesMode
	Script    []byte
	ABI       ContractInterface
	ROM       []byte

	// Extra fields, not serializable
	SeriesID string
	Symbol   string
}

func (s *TokenSeries_S) Serialize(writer *io.BinWriter) {
	writer.WriteBigIntegerFromString(s.MintCount)
	writer.WriteBigIntegerFromString(s.MaxSupply)
	writer.WriteB(byte(s.Mode))
	writer.WriteVarBytes(s.Script)

	// Chain storage wraps the ABI bytes in a var-bytes field.
	bytes := io.Serialize[*ContractInterface](&s.ABI)
	writer.WriteVarBytes(bytes)

	writer.WriteVarBytes(s.ROM)
}

func (s *TokenSeries_S) Deserialize(reader *io.BinReader) {
	s.MintCount = reader.ReadBigIntegerToString()
	s.MaxSupply = reader.ReadBigIntegerToString()
	s.Mode = TokenSeriesMode(reader.ReadB())
	s.Script = reader.ReadVarBytes()

	// Chain storage wraps the ABI bytes in a var-bytes field.
	bytes := reader.ReadVarBytes()
	s.ABI = *io.Deserialize[*ContractInterface](bytes)

	s.ROM = reader.ReadVarBytes()
}
