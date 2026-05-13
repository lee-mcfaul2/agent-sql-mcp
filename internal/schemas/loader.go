package schemas

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
)

// RawSchema holds the request and response schema bytes for one tool.
type RawSchema struct {
	RequestJSON  []byte
	ResponseJSON []byte
}

// Tools is the canonical list expected; the loader complains if any are missing.
var Tools = []string{
	"search_customer",
	"lookup_customer",
	"list_orders",
	"get_order",
}

// Catalog holds the loaded raw schemas + the computed digest.
type Catalog struct {
	Tools  map[string]RawSchema
	Digest string // hex(sha256(canonical))
}

// LoadFromBytes accepts pre-read bytes per file (filename -> bytes).
// Production wiring passes the embed.FS content; tests pass synthetic maps.
func LoadFromBytes(files map[string][]byte) (*Catalog, error) {
	cat := &Catalog{Tools: map[string]RawSchema{}}
	for _, tool := range Tools {
		req, ok := files[tool+".request.json"]
		if !ok {
			return nil, fmt.Errorf("missing schema file: %s.request.json", tool)
		}
		resp, ok := files[tool+".response.json"]
		if !ok {
			return nil, fmt.Errorf("missing schema file: %s.response.json", tool)
		}
		if err := verifyJSON(req); err != nil {
			return nil, fmt.Errorf("%s.request.json: %w", tool, err)
		}
		if err := verifyJSON(resp); err != nil {
			return nil, fmt.Errorf("%s.response.json: %w", tool, err)
		}
		cat.Tools[tool] = RawSchema{RequestJSON: req, ResponseJSON: resp}
	}
	digest, err := computeDigest(cat)
	if err != nil {
		return nil, err
	}
	cat.Digest = digest
	return cat, nil
}

func verifyJSON(b []byte) error {
	var v any
	return json.Unmarshal(b, &v)
}

func computeDigest(cat *Catalog) (string, error) {
	type entry struct {
		Tool string
		Req  []byte
		Resp []byte
	}
	entries := make([]entry, 0, len(cat.Tools))
	for k, v := range cat.Tools {
		req, err := canonicalJSON(v.RequestJSON)
		if err != nil {
			return "", err
		}
		resp, err := canonicalJSON(v.ResponseJSON)
		if err != nil {
			return "", err
		}
		entries = append(entries, entry{Tool: k, Req: req, Resp: resp})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Tool < entries[j].Tool })

	var buf bytes.Buffer
	for _, e := range entries {
		buf.WriteString(e.Tool)
		buf.WriteByte('\n')
		buf.Write(e.Req)
		buf.WriteByte('\n')
		buf.Write(e.Resp)
		buf.WriteByte('\n')
	}
	sum := sha256.Sum256(buf.Bytes())
	return hex.EncodeToString(sum[:]), nil
}

func canonicalJSON(in []byte) ([]byte, error) {
	var v any
	if err := json.Unmarshal(in, &v); err != nil {
		return nil, err
	}
	// json.Marshal sorts map keys for map[string]any.
	return json.Marshal(v)
}
