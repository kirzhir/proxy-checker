package proxy

import (
	"context"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"testing"
	"time"
)

func TestFileReader_Read(t *testing.T) {

	filepath := "testdata/inventory.txt"
	proxies := []string{"127.0.0.1:8080", "192.168.0.1:3128"}

	reader := NewFileReader(filepath)
	proxiesCh := make(chan string, len(proxies))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	go func() {
		err := reader.Read(ctx, proxiesCh)
		if err != nil {
			t.Fatalf("failed to read proxies: %v", err)
		}

		close(proxiesCh)
	}()

	readProxies := []string{}
	for p := range proxiesCh {
		readProxies = append(readProxies, p)
	}

	if len(readProxies) != len(proxies) {
		t.Fatalf("expected %d proxies, got %d", len(proxies), len(readProxies))
	}

	for i, p := range readProxies {
		if p != proxies[i] {
			t.Errorf("expected %s, got %s", proxies[i], p)
		}
	}
}

func TestStdinReader_Read(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	defer r.Close()

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	os.Stdin = r

	input := "127.0.0.1:8080\n192.168.0.1:3128\nexit\n"
	go func() {
		defer w.Close()
		if _, err := io.WriteString(w, input); err != nil {
			t.Fatalf("failed to write to pipe: %v", err)
		}
	}()

	reader := NewStdinReader()
	proxiesCh := make(chan string, 2)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err = reader.Read(ctx, proxiesCh)
	if err != nil {
		t.Fatalf("failed to read from stdin: %v", err)
	}
	close(proxiesCh)

	expectedProxies := []string{"127.0.0.1:8080", "192.168.0.1:3128"}
	readProxies := []string{}
	for p := range proxiesCh {
		readProxies = append(readProxies, p)
	}

	if len(readProxies) != len(expectedProxies) {
		t.Fatalf("expected %d proxies, got %d", len(expectedProxies), len(readProxies))
	}

	for i, p := range readProxies {
		if p != expectedProxies[i] {
			t.Errorf("expected %s, got %s", expectedProxies[i], p)
		}
	}
}

func TestExpandPath(t *testing.T) {
	usr, err := user.Current()
	if err != nil {
		t.Fatalf("failed to get current user: %v", err)
	}

	homeDir := usr.HomeDir
	tests := []struct {
		input    string
		expected string
	}{
		{"~/test", filepath.Join(homeDir, "test")},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
	}

	for _, test := range tests {
		result, err := expandPath(test.input)
		if err != nil {
			t.Fatalf("failed to expand path: %v", err)
		}

		if result != test.expected {
			t.Errorf("expected %s, got %s", test.expected, result)
		}
	}
}
