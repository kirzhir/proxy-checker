package proxy

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func TestFileWriter_Write(t *testing.T) {
	filename := "testdata/write_proxies.txt"
	proxies := []string{"127.0.0.1:8080", "192.168.0.1:3128"}

	writer := NewFileWriter(filename)
	proxiesCh := make(chan string, len(proxies))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Write proxies to the channel
	go func() {
		defer close(proxiesCh)
		for _, proxy := range proxies {
			proxiesCh <- proxy
		}
	}()

	err := writer.Write(ctx, proxiesCh)
	if err != nil {
		t.Fatalf("failed to write proxies to file: %v", err)
	}

	// Verify the content of the file
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	readProxies := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(readProxies) != len(proxies) {
		t.Fatalf("expected %d proxies, got %d", len(proxies), len(readProxies))
	}

	for i, p := range readProxies {
		if p != proxies[i] {
			t.Errorf("expected %s, got %s", proxies[i], p)
		}
	}
}

func TestStdoutWriter_Write(t *testing.T) {
	proxies := []string{"127.0.0.1:8080", "192.168.0.1:3128"}

	writer := NewStdoutWriter()
	proxiesCh := make(chan string, len(proxies))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	go func() {
		defer close(proxiesCh)
		for _, proxy := range proxies {
			proxiesCh <- proxy
		}
	}()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	defer r.Close()
	defer w.Close()

	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()
	os.Stdout = w

	writeErrCh := make(chan error)
	go func() {
		writeErrCh <- writer.Write(ctx, proxiesCh)
	}()

	var buf bytes.Buffer
	doneCh := make(chan struct{})
	go func() {
		_, err := io.Copy(&buf, r)
		if err != nil {
			t.Errorf("failed to read from pipe: %v", err)
		}
		close(doneCh)
	}()

	err = <-writeErrCh
	if err != nil {
		t.Fatalf("failed to write proxies to stdout: %v", err)
	}

	w.Close()
	<-doneCh

	output := strings.TrimSpace(buf.String())
	readProxies := strings.Split(output, "\n")
	if len(readProxies) != len(proxies) {
		t.Fatalf("expected %d proxies, got %d", len(proxies), len(readProxies))
	}

	for i, p := range readProxies {
		if p != proxies[i] {
			t.Errorf("expected %s, got %s", proxies[i], p)
		}
	}
}
