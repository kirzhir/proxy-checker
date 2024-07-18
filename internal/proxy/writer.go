package proxy

import (
	"context"
	"fmt"
	"os"
)

type Writer interface {
	Write(ctx context.Context, proxiesCh chan<- string) error
}

type StdoutWriter struct{}

func NewStdoutWriter() *StdoutWriter {
	return &StdoutWriter{}
}

func (w *StdoutWriter) Write(ctx context.Context, proxiesCh <-chan string) error {
	for {
		select {
		case <-ctx.Done():
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
