package schemas

import (
	"encoding/json"
	"fmt"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// Validators is a compiled set of per-tool request+response validators.
type Validators struct {
	request  map[string]*jsonschema.Schema
	response map[string]*jsonschema.Schema
}

// CompileValidators builds the validators from a loaded Catalog.
func CompileValidators(cat *Catalog) (*Validators, error) {
	v := &Validators{
		request:  map[string]*jsonschema.Schema{},
		response: map[string]*jsonschema.Schema{},
	}
	for tool, raw := range cat.Tools {
		req, err := compileOne(tool+".request", raw.RequestJSON)
		if err != nil {
			return nil, fmt.Errorf("%s.request: %w", tool, err)
		}
		resp, err := compileOne(tool+".response", raw.ResponseJSON)
		if err != nil {
			return nil, fmt.Errorf("%s.response: %w", tool, err)
		}
		v.request[tool] = req
		v.response[tool] = resp
	}
	return v, nil
}

func compileOne(name string, raw []byte) (*jsonschema.Schema, error) {
	// Unmarshal JSON bytes into any to pass to AddResource
	var doc any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, err
	}

	c := jsonschema.NewCompiler()
	if err := c.AddResource(name, doc); err != nil {
		return nil, err
	}
	return c.Compile(name)
}

// ValidateRequest checks the payload against the named tool's request schema.
func (v *Validators) ValidateRequest(tool string, payload any) error {
	s, ok := v.request[tool]
	if !ok {
		return fmt.Errorf("unknown tool: %s", tool)
	}
	return s.Validate(payload)
}

// ValidateResponse checks the payload against the named tool's response schema.
func (v *Validators) ValidateResponse(tool string, payload any) error {
	s, ok := v.response[tool]
	if !ok {
		return fmt.Errorf("unknown tool: %s", tool)
	}
	return s.Validate(payload)
}

// AsAny turns []byte JSON into the any tree jsonschema expects.
func AsAny(b []byte) (any, error) {
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return nil, err
	}
	return v, nil
}
