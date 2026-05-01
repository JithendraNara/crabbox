package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestCoordinatorMachineIDAcceptsStringOrNumber(t *testing.T) {
	for name, input := range map[string]string{
		"string": `{"id":"i-123","labels":{}}`,
		"number": `{"id":128694755,"labels":{}}`,
	} {
		t.Run(name, func(t *testing.T) {
			var machine CoordinatorMachine
			if err := json.Unmarshal([]byte(input), &machine); err != nil {
				t.Fatal(err)
			}
			if machine.ID == "" {
				t.Fatalf("machine ID was empty")
			}
		})
	}
}

func TestSplitCurlResponseParsesTrailingStatus(t *testing.T) {
	body, status, err := splitCurlResponse([]byte("{\"ok\":true}\n200"))
	if err != nil {
		t.Fatal(err)
	}
	if status != 200 {
		t.Fatalf("status = %d, want 200", status)
	}
	if string(body) != `{"ok":true}` {
		t.Fatalf("body = %q", body)
	}
}

func TestDecodeCoordinatorResponseCanReadTextBody(t *testing.T) {
	var buf bytes.Buffer
	if err := decodeCoordinatorResponse("GET", "/v1/runs/run_1/logs", 200, strings.NewReader("hello"), &buf); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "hello" {
		t.Fatalf("body=%q", buf.String())
	}
}

func TestCurlConfigKeepsBearerTokenInConfig(t *testing.T) {
	client := CoordinatorClient{BaseURL: "https://example.test", Token: "secret-token"}
	config, cleanup, err := client.curlConfig("POST", "/v1/leases", []byte(`{"leaseID":"cbx"}`), true)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	for _, want := range []string{
		`url = "https://example.test/v1/leases"`,
		`request = "POST"`,
		`header = "Authorization: Bearer secret-token"`,
		`header = "Content-Type: application/json"`,
		`data-binary = "@`,
	} {
		if !strings.Contains(config, want) {
			t.Fatalf("config missing %q:\n%s", want, config)
		}
	}
	bodyPath := curlConfigValueForTest(t, config, "data-binary")
	bodyPath = strings.TrimPrefix(bodyPath, "@")
	if _, err := os.Stat(bodyPath); err != nil {
		t.Fatalf("body file missing: %v", err)
	}
}

func curlConfigValueForTest(t *testing.T, config, key string) string {
	t.Helper()
	prefix := key + " = "
	for _, line := range strings.Split(config, "\n") {
		if strings.HasPrefix(line, prefix) {
			var value string
			if err := json.Unmarshal([]byte(strings.TrimPrefix(line, prefix)), &value); err != nil {
				t.Fatal(err)
			}
			return value
		}
	}
	t.Fatalf("config key %q missing:\n%s", key, config)
	return ""
}
