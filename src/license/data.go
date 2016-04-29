package license

import (
	"encoding/json"
	"errors"
	"time"
)

// Data represents the data of a license entry. Specifically, it includes
// the protocol specific metadata and the custom license payload data.
type Data struct {
	Meta    *Metadata              `json:"meta"`
	Payload map[string]interface{} `json:"payload"`
}

// Metadata is the protocol specific metadata describing a license, including
// its unique identifier, valid date range and any child licenses it may
// issue.
type Metadata struct {
	ID          string               `json:"id"`
	ActivatesOn time.Time            `json:"activates"`
	ExpiresOn   time.Time            `json:"expires"`
	Pack        map[string]*Template `json:"pack"`
}

// Template represents a class of licenses as well as the number of licenses
// of that type which may be generated.
type Template struct {
	Count   int                    `json:"count"`
	Payload map[string]interface{} `json:"payload"`
}

// EncodeData is responsible for encoding a license Data structure into
// its binary form for transfer between nodes or further processing.
func EncodeData(data *Data) ([]byte, error) {
	if data == nil {
		return nil, errors.New("expected a non-nil license to be provided for encoding")
	}

	return json.Marshal(data)
}

// DecodeData is responsible for decoding a previously encoded, binary
// license data object, into its native format for further use.
func DecodeData(data []byte) (*Data, error) {
	d := Data{}

	err := json.Unmarshal(data, &d)
	if err != nil {
		return nil, err
	}

	return &d, nil
}
