package receipt

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"time"

	"github.com/mr-tron/base58"
)

type Receipt struct {
	Version    string   `json:"version"`
	ProviderPK string   `json:"provider_pk"`
	Model      string   `json:"model"`
	PromptHash string   `json:"prompt_hash"`
	OutputHash string   `json:"output_hash"`
	TokensIn   int      `json:"tokens_in"`
	TokensOut  int      `json:"tokens_out"`
	StartISO   string   `json:"start_iso"`
	EndISO     string   `json:"end_iso"`
	DurationMs int64    `json:"duration_ms"`
	Epoch      int64    `json:"epoch"`
	Seq        int64    `json:"seq"`
	PrevHash   string   `json:"prev_hash"`
	Canary     Canary   `json:"canary"`
	Rate       RateInfo `json:"rate"`
	ReceiptID  string   `json:"receipt_id"`
}

type Canary struct {
	ID     string `json:"id"`
	Passed bool   `json:"passed"`
}

type RateInfo struct {
	Throttle  bool `json:"throttle"`
	Truncated bool `json:"truncated"`
}

type SignedReceipt struct {
	Receipt   Receipt `json:"receipt"`
	Signature string  `json:"signature"`
}

func CanonicalizeJSON(v interface{}) ([]byte, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}

	return json.Marshal(sortKeys(m))
}

func sortKeys(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := m[k]
		switch val := v.(type) {
		case map[string]interface{}:
			result[k] = sortKeys(val)
		case []interface{}:
			for i, item := range val {
				if mapItem, ok := item.(map[string]interface{}); ok {
					val[i] = sortKeys(mapItem)
				}
			}
			result[k] = val
		default:
			result[k] = v
		}
	}

	return result
}

func HashData(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func GenerateReceiptID(hash string) string {
	decoded, _ := hex.DecodeString(hash)
	if len(decoded) >= 16 {
		return base58.Encode(decoded[:16])
	}
	return base58.Encode(decoded)
}

func NewReceipt(providerPK, model, promptHash, outputHash string, tokensIn, tokensOut int, start, end time.Time) *Receipt {
	epoch := start.UTC().Unix() / 86400

	receipt := &Receipt{
		Version:    "1.0.0",
		ProviderPK: providerPK,
		Model:      model,
		PromptHash: promptHash,
		OutputHash: outputHash,
		TokensIn:   tokensIn,
		TokensOut:  tokensOut,
		StartISO:   start.UTC().Format(time.RFC3339),
		EndISO:     end.UTC().Format(time.RFC3339),
		DurationMs: end.Sub(start).Milliseconds(),
		Epoch:      epoch,
		Seq:        time.Now().UnixNano(),
		PrevHash:   "",
		Canary: Canary{
			ID:     "",
			Passed: true,
		},
		Rate: RateInfo{
			Throttle:  false,
			Truncated: false,
		},
	}

	canonical, _ := CanonicalizeJSON(receipt)
	hash := HashData(canonical)
	receipt.ReceiptID = GenerateReceiptID(hash)

	return receipt
}
