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

func NewReader(in string) Reader {
	if in == "stdin" {
		return NewStdinReader()
	}

	return NewFileReader(in)
}

func NewFileReader(filename string) Reader {
	return &FileReader{filename: filename}
}

func NewStdinReader() *StdinReader {
	return &StdinReader{}
}

func (r *FileReader) Read(ctx context.Context, proxiesCh chan<- string) error {
	defer close(proxiesCh)

	filename, err := expandPath(r.filename)
	if err != nil {
		return err
	}

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	var line string
	reader := bufio.NewReader(file)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:

		}

		if line, err = reader.ReadString('\n'); err == io.EOF {
			break
		}

		proxiesCh <- strings.TrimSpace(line)
	}

	return nil
}

func (r *StdinReader) Read(ctx context.Context, proxiesCh chan<- string) error {
	scanner := bufio.NewScanner(os.Stdin)

	if _, err := fmt.Fprintln(os.Stdout, "Enter proxy address IP:PORT"); err != nil {
		return err
	}

	errCh := make(chan error)

	go func() {
		defer close(proxiesCh)
		defer close(errCh)

		var line string
		for {
			if !scanner.Scan() {
				errCh <- scanner.Err()
				break
			}

			if line = strings.TrimSpace(scanner.Text()); line == "exit" {
				break
			}

			if line == "" {
				continue
			}

			proxiesCh <- line
		}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return err
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
