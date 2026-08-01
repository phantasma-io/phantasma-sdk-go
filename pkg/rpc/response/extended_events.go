package response

import "encoding/json"

// EventData is the payload of one extended event, typed by the kind of the event that carries it.
//
// The set of shapes is closed - only the types in this package implement the interface - so a type
// switch over them covers everything this build models:
//
//	switch data := event.Data.(type) {
//	case SpecialResolutionData:
//		for _, call := range data.Calls { /* call.Arguments is typed per method */ }
//	case MarketOrderData:
//		use(data.Price)
//	case UnknownEventData:
//		log(data.JSON)
//	}
//
// Decoding is total by design: an event kind this build does not model, and a modeled kind whose
// payload does not match its shape, both arrive as UnknownEventData with the JSON preserved
// verbatim. One unexpected event can therefore never fail a whole block answer, and an event whose
// Kind names a modeled shape while Data is UnknownEventData is how a consumer detects that the
// answering node is newer than this SDK.
//
// Numeric fields follow the wire exactly: chain amounts and big-integer ids travel as strings,
// while Carbon-side ids (carbonTokenId, carbonInstanceId, resolutionId) and timestamps are plain
// JSON numbers.
type EventData interface {
	isEventData()
}

// TokenCreateData is the payload of a TokenCreate extended event.
type TokenCreateData struct {
	Symbol        string `json:"symbol"`
	MaxSupply     string `json:"maxSupply"`
	Decimals      uint32 `json:"decimals"`
	IsNonFungible bool   `json:"isNonFungible"`
	CarbonTokenID uint64 `json:"carbonTokenId"`
	// Metadata values are rendered to strings by the node, so unlike the metadata of a token
	// response they are not VM values here. Keys arrive exactly as the chain stores them.
	Metadata map[string]string `json:"metadata"`
}

func (TokenCreateData) isEventData() {}

// TokenSeriesCreateData is the payload of a TokenSeriesCreate extended event.
type TokenSeriesCreateData struct {
	Symbol string `json:"symbol"`
	// SeriesID is the Phantasma series id, a big integer rendered as a string.
	SeriesID       string `json:"seriesId"`
	MaxMint        uint32 `json:"maxMint"`
	MaxSupply      uint32 `json:"maxSupply"`
	Owner          string `json:"owner"`
	CarbonTokenID  uint64 `json:"carbonTokenId"`
	CarbonSeriesID uint32 `json:"carbonSeriesId"`
	// Metadata values are rendered to strings by the node; keys arrive as the chain stores them.
	Metadata map[string]string `json:"metadata"`
}

func (TokenSeriesCreateData) isEventData() {}

// MarketOrderData is the payload of an OrderCreated, OrderCancelled or OrderFilled extended event.
// The three kinds share one shape; the carrying event's Kind tells them apart.
type MarketOrderData struct {
	BaseSymbol  string `json:"baseSymbol"`
	QuoteSymbol string `json:"quoteSymbol"`
	// TokenID is the Phantasma NFT id, a big integer rendered as a string.
	TokenID            string `json:"tokenId"`
	CarbonBaseTokenID  uint64 `json:"carbonBaseTokenId"`
	CarbonQuoteTokenID uint64 `json:"carbonQuoteTokenId"`
	CarbonInstanceID   uint64 `json:"carbonInstanceId"`
	Seller             string `json:"seller"`
	// Buyer repeats the seller on a cancel: that path has no buyer by definition and the payload
	// shape stays stable.
	Buyer     string `json:"buyer"`
	Price     string `json:"price"`
	EndPrice  string `json:"endPrice"`
	StartDate int64  `json:"startDate"`
	EndDate   int64  `json:"endDate"`
	// Type is the auction type name, for example "Fixed".
	Type string `json:"type"`
}

func (MarketOrderData) isEventData() {}

// UnknownEventData carries the payload of an event kind this build does not model, or of a modeled
// kind whose payload did not match its shape. The JSON is kept verbatim so nothing the node
// answered is ever lost.
type UnknownEventData struct {
	JSON json.RawMessage
}

func (UnknownEventData) isEventData() {}

// MarshalJSON writes the preserved payload back unchanged.
func (u UnknownEventData) MarshalJSON() ([]byte, error) {
	if len(u.JSON) == 0 {
		return []byte("null"), nil
	}
	return u.JSON, nil
}

// EventExResult describes one extended transaction event returned by current RPC nodes.
type EventExResult struct {
	Address  string `json:"address"`
	Contract string `json:"contract"`
	Kind     string `json:"kind"`
	// Data is nil only when the event carries no payload at all.
	Data EventData `json:"data,omitempty"`
}

// eventExResultWire is the envelope as it arrives; Data stays raw until Kind has been read, because
// the kind is what decides the payload shape.
type eventExResultWire struct {
	Address  string          `json:"address"`
	Contract string          `json:"contract"`
	Kind     string          `json:"kind"`
	Data     json.RawMessage `json:"data"`
}

func (e *EventExResult) UnmarshalJSON(data []byte) error {
	data, err := stripExactFields(data, "Kind", "Data")
	if err != nil {
		return err
	}

	var wire eventExResultWire
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}

	e.Address = wire.Address
	e.Contract = wire.Contract
	e.Kind = wire.Kind
	e.Data = DecodeEventData(wire.Kind, wire.Data)
	return nil
}

// DecodeEventData types a raw extended-event payload from the kind of the event that carries it.
//
// Exported because a consumer that reads blocks through its own transport still needs the same
// dispatch. An absent payload yields nil; everything else follows the totality rule documented on
// EventData.
func DecodeEventData(kind string, data json.RawMessage) EventData {
	if len(data) == 0 {
		return nil
	}

	switch kind {
	case "TokenCreate":
		return typedEventData[TokenCreateData](data)
	case "TokenSeriesCreate":
		return typedEventData[TokenSeriesCreateData](data)
	case "OrderCreated", "OrderCancelled", "OrderFilled":
		return typedEventData[MarketOrderData](data)
	case "SpecialResolution":
		return typedEventData[SpecialResolutionData](data)
	default:
		return UnknownEventData{JSON: data}
	}
}

// typedEventData decodes one modeled shape, keeping the raw payload when it does not match.
func typedEventData[T EventData](data json.RawMessage) EventData {
	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		return UnknownEventData{JSON: data}
	}
	return value
}
