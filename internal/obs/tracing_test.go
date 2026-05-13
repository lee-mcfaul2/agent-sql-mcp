package obs

import (
	"context"
	"testing"
)

func TestSetupTracing_NoEndpointIsNoop(t *testing.T) {
	shutdown, err := SetupTracing(context.Background(), "", "test")
	if err != nil {
		t.Fatal(err)
	}
	if err := shutdown(context.Background()); err != nil {
		t.Errorf("noop shutdown should not error: %v", err)
	}
}
