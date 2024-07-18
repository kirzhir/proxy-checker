package proxy

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

type Reader interface {
	Read(ctx context.Context, proxiesCh chan<- string) error
}

type FileReader struct {
	filename string
}

type StdinReader struct {
}

func NewFileReader(filename string) Reader {
	return &FileReader{filename: filename}
}

func NewStdinReader() *StdinReader {
	return &StdinReader{}
}

func (r *FileReader) Read(ctx context.Context, proxiesCh chan<- string) error {
	filename, err := expandPath(r.filename)
	if err != nil {
		return err
	}

	file, err := os.Open(filename)
	if err != nil {
		return err
	}

	reader := bufio.NewReader(file)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:

		}

		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}

		proxiesCh <- strings.TrimSpace(line)
	}

	return file.Close()
}

func (r *StdinReader) Read(ctx context.Context, proxiesCh chan<- string) error {
	scanner := bufio.NewScanner(os.Stdin)

	if _, err := fmt.Fprintln(os.Stdout, "Enter proxy address IP:PORT"); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:

		}

		if !scanner.Scan() {
			return scanner.Err()
		}

		line := strings.TrimSpace(scanner.Text())

		if line == "exit" {
			return nil
		}

		if line == "" {
			continue
		}

		proxiesCh <- line
	}
}

func expandPath(filename string) (string, error) {
	if strings.HasPrefix(filename, "~") {
		usr, err := user.Current()
		if err != nil {
			return "", err
		}
		return filepath.Join(usr.HomeDir, filename[1:]), nil
	}
	return filename, nil
}
