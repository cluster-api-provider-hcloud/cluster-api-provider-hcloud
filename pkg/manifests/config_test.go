package manifests

import (
	"bufio"
	"bytes"
	"testing"
)

func TestEvaluateJsonnet(t *testing.T) {
	path := "../../manifests/config-extvar.jsonnet"
	var buf bytes.Buffer
	err := evaluateJsonnet(
		bufio.NewWriter(&buf),
		path,
		sampleParameters().ExtVar(),
	)

	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	/*
		if buf.Len() == 0 {
			t.Error("no YAML data received")
		}
	*/
}
