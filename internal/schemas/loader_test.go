package schemas

import (
	"testing"
)

func validBytes() map[string][]byte {
	make := func() []byte {
		return []byte(`{"type":"object"}`)
	}
	m := map[string][]byte{}
	for _, t := range Tools {
		m[t+".request.json"] = make()
		m[t+".response.json"] = make()
	}
	return m
}

func TestLoad_OK(t *testing.T) {
	cat, err := LoadFromBytes(validBytes())
	if err != nil {
		t.Fatal(err)
	}
	if len(cat.Tools) != 5 {
		t.Errorf("expected 5 tools, got %d", len(cat.Tools))
	}
	if len(cat.Digest) != 64 {
		t.Errorf("digest should be 64 hex chars, got %q", cat.Digest)
	}
}

func TestLoad_DigestStable(t *testing.T) {
	a, _ := LoadFromBytes(validBytes())
	b, _ := LoadFromBytes(validBytes())
	if a.Digest != b.Digest {
		t.Errorf("digest unstable across loads: %s vs %s", a.Digest, b.Digest)
	}
}

func TestLoad_DigestChangesOnContentDiff(t *testing.T) {
	a := validBytes()
	b := validBytes()
	b["search_customer.request.json"] = []byte(`{"type":"object","title":"x"}`)
	ca, _ := LoadFromBytes(a)
	cb, _ := LoadFromBytes(b)
	if ca.Digest == cb.Digest {
		t.Errorf("digest should differ when content differs")
	}
}

func TestLoad_MissingFileFails(t *testing.T) {
	m := validBytes()
	delete(m, "search_customer.request.json")
	if _, err := LoadFromBytes(m); err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoad_InvalidJSONFails(t *testing.T) {
	m := validBytes()
	m["search_customer.request.json"] = []byte(`not json`)
	if _, err := LoadFromBytes(m); err == nil {
		t.Fatal("expected error for invalid json")
	}
}
