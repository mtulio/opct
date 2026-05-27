package cleaner

import (
	"testing"
)

func TestScanContentForLeaks_OpenShiftToken(t *testing.T) {
	content := []byte(`some log line
token: sha256~abcdefghijklmnopqrstuvwxyz01234567890ABCDEF
another line`)
	findings := ScanContentForLeaks("test.log", content)
	if len(findings) == 0 {
		t.Fatal("expected to find OpenShift User Token")
	}
	if findings[0].Pattern != "OpenShift User Token" {
		t.Errorf("expected pattern 'OpenShift User Token', got '%s'", findings[0].Pattern)
	}
	if findings[0].Line != 2 {
		t.Errorf("expected line 2, got %d", findings[0].Line)
	}
}

func TestScanContentForLeaks_AWSKey(t *testing.T) {
	content := []byte(`config:
  accessKeyId: AKIAIOSFODNN7EXAMPLE
  secretKey: something`)
	findings := ScanContentForLeaks("config.yaml", content)
	found := false
	for _, f := range findings {
		if f.Pattern == "AWS IAM Unique Identifier" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected to find AWS IAM Unique Identifier")
	}
}

func TestScanContentForLeaks_PrivateKey(t *testing.T) {
	content := []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA0Z3VS5JJcds3xfn/ygWyF9PbnGy0AHB7MhgHcTz6sE2I2yPB
anotherlineofbase64anotherlineofbase64anotherlineofbase64anotherlineofba
-----END RSA PRIVATE KEY-----`)
	findings := ScanContentForLeaks("key.pem", content)
	if len(findings) == 0 {
		t.Fatal("expected to find Private Key (PEM)")
	}
	if findings[0].Pattern != "Private Key (PEM)" {
		t.Errorf("expected pattern 'Private Key (PEM)', got '%s'", findings[0].Pattern)
	}
}

func TestScanContentForLeaks_GenericSecret(t *testing.T) {
	content := []byte(`DATABASE_PASSWORD=supersecretvalue123`)
	findings := ScanContentForLeaks("env.txt", content)
	found := false
	for _, f := range findings {
		if f.Pattern == "Generic Secret (key=value unquoted)" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected to find Generic Secret (key=value unquoted)")
	}
}

func TestScanContentForLeaks_GitHubToken(t *testing.T) {
	content := []byte(`token = "ghp_ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefgh12"`)
	findings := ScanContentForLeaks("config.json", content)
	found := false
	for _, f := range findings {
		if f.Pattern == "GitHub Personal Access Token" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected to find GitHub Personal Access Token")
	}
}

func TestScanContentForLeaks_NoFindings(t *testing.T) {
	content := []byte(`this is a normal log file
with no secrets or sensitive data
just regular cluster information
openshift version 4.22`)
	findings := ScanContentForLeaks("normal.log", content)
	if len(findings) != 0 {
		t.Errorf("expected no findings, got %d: %v", len(findings), findings)
	}
}

func TestScanContentForLeaks_BinarySkip(t *testing.T) {
	content := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00}
	findings := ScanContentForLeaks("image.png", content)
	if len(findings) != 0 {
		t.Errorf("expected no findings for binary file, got %d", len(findings))
	}
}

func TestScanContentForLeaks_EmptyContent(t *testing.T) {
	findings := ScanContentForLeaks("empty.txt", []byte{})
	if len(findings) != 0 {
		t.Errorf("expected no findings for empty content, got %d", len(findings))
	}
}

func TestScanContentForLeaks_GCPKey(t *testing.T) {
	content := []byte(`api_key: AIzaSyA1234567890abcdefghijklmnopqrstuv`)
	findings := ScanContentForLeaks("gcp.yaml", content)
	found := false
	for _, f := range findings {
		if f.Pattern == "GCP API Key" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected to find GCP API Key")
	}
}

func BenchmarkScanContentForLeaks(b *testing.B) {
	content := make([]byte, 100*1024)
	for i := range content {
		content[i] = byte('a' + (i % 26))
		if i%80 == 79 {
			content[i] = '\n'
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ScanContentForLeaks("benchmark.log", content)
	}
}
