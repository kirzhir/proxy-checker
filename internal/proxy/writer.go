package proxy

import (
	"bufio"
	"context"
	"fmt"
	"os"
)

type Writer interface {
	Write(ctx context.Context, proxiesCh <-chan string) error
}

type FileWriter struct {
	filename string
}

type StdoutWriter struct{}

func NewStdoutWriter() Writer {
	return &StdoutWriter{}
}

func NewFileWriter(filename string) *FileWriter {
	return &FileWriter{filename: filename}
}

func (w *FileWriter) Write(ctx context.Context, proxiesCh <-chan string) error {
	filename, err := expandPath(w.filename)
	if err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	for {
		select {
		case <-ctx.Done():
			for proxy := range proxiesCh {
				if _, err = writer.WriteString(proxy + "\n"); err != nil {
					return fmt.Errorf("failed to write remaining data to file: %w", err)
				}
			}

			return writer.Flush()
		case proxy, ok := <-proxiesCh:
			if !ok {
				return writer.Flush()
			}

			if _, err = writer.WriteString(proxy + "\n"); err != nil {
				return fmt.Errorf("failed to write to file: %w", err)
			}
		}
	}
}

func (w *StdoutWriter) Write(ctx context.Context, proxiesCh <-chan string) error {
	for {
		select {
		case <-ctx.Done():
			for proxy := range proxiesCh {
				if _, err := fmt.Fprintln(os.Stdout, proxy); err != nil {
					return err
				}
			}

			return ctx.Err()
		case proxy, ok := <-proxiesCh:
			if !ok {
				return nil
			}

			if _, err := fmt.Fprintln(os.Stdout, proxy); err != nil {
				return err
			}
		}
	}
}
