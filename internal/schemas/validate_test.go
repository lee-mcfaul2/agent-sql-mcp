package schemas

import (
	"testing"
)

func realBytes(t *testing.T) map[string][]byte {
	t.Helper()
	return map[string][]byte{
		"search_customer.request.json": []byte(`{
			"type":"object",
			"properties":{"name":{"type":"string"},"email":{"type":"string"}},
			"anyOf":[{"required":["name"]},{"required":["email"]}],
			"additionalProperties":false
		}`),
		"search_customer.response.json":   []byte(`{"type":"object"}`),
		"lookup_customer.request.json":    []byte(`{"type":"object","required":["customer_id"],"properties":{"customer_id":{"type":"integer"}}}`),
		"lookup_customer.response.json":   []byte(`{"type":"object"}`),
		"list_orders.request.json":        []byte(`{"type":"object","required":["customer_id"],"properties":{"customer_id":{"type":"integer"}}}`),
		"list_orders.response.json":       []byte(`{"type":"object"}`),
		"list_transactions.request.json":  []byte(`{"type":"object","required":["customer_id"],"properties":{"customer_id":{"type":"integer"}}}`),
		"list_transactions.response.json": []byte(`{"type":"object"}`),
		"get_order.request.json":          []byte(`{"type":"object","required":["order_id"],"properties":{"order_id":{"type":"integer"}}}`),
		"get_order.response.json":         []byte(`{"type":"object"}`),
		"list_all_customers.request.json":    []byte(`{"type":"object"}`),
		"list_all_customers.response.json":   []byte(`{"type":"object"}`),
		"list_all_orders.request.json":       []byte(`{"type":"object"}`),
		"list_all_orders.response.json":      []byte(`{"type":"object"}`),
		"list_all_transactions.request.json": []byte(`{"type":"object"}`),
		"list_all_transactions.response.json":[]byte(`{"type":"object"}`),
	}
}

func TestValidate_RequestOK(t *testing.T) {
	cat, _ := LoadFromBytes(realBytes(t))
	v, err := CompileValidators(cat)
	if err != nil {
		t.Fatal(err)
	}
	payload, _ := AsAny([]byte(`{"name":"Alice"}`))
	if err := v.ValidateRequest("search_customer", payload); err != nil {
		t.Fatalf("expected pass, got %v", err)
	}
}

func TestValidate_RequestFailsAnyOf(t *testing.T) {
	cat, _ := LoadFromBytes(realBytes(t))
	v, _ := CompileValidators(cat)
	payload, _ := AsAny([]byte(`{}`))
	if err := v.ValidateRequest("search_customer", payload); err == nil {
		t.Fatal("expected failure: anyOf requires name or email")
	}
}

func TestValidate_UnknownTool(t *testing.T) {
	cat, _ := LoadFromBytes(realBytes(t))
	v, _ := CompileValidators(cat)
	if err := v.ValidateRequest("nope", map[string]any{}); err == nil {
		t.Fatal("expected unknown-tool error")
	}
}

func TestValidate_LookupRequiresCustomerID(t *testing.T) {
	cat, _ := LoadFromBytes(realBytes(t))
	v, _ := CompileValidators(cat)
	payload, _ := AsAny([]byte(`{}`))
	if err := v.ValidateRequest("lookup_customer", payload); err == nil {
		t.Fatal("expected required-field error")
	}
}
