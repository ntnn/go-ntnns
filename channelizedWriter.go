package ntnns

import (
	"context"
	"fmt"
	"io"
	"sync"
)

var _ io.Writer = (*ChannelizedWriter)(nil)

var (
	ChannelizedWriterQueueSize = 100

	ErrChannelizedWriterClosed = fmt.Errorf("closed")
)

type ChannelizedWriter struct {
	ctx        context.Context
	writer     io.Writer
	writeQueue chan []byte
	err        error
	wg         *sync.WaitGroup
	done       bool
}

func NewChannelizedWriter(ctx context.Context, w io.Writer, wg *sync.WaitGroup) *ChannelizedWriter {
	if ctx == nil {
		ctx = context.Background()
	}

	cw := new(ChannelizedWriter)
	cw.ctx = ctx
	cw.writer = w
	cw.writeQueue = make(chan []byte, ChannelizedWriterQueueSize)
	cw.wg = wg

	go cw.worker()

	return cw
}

func (cw *ChannelizedWriter) Close() {
	cw.err = ErrChannelizedWriterClosed
	close(cw.writeQueue)
}

func (cw *ChannelizedWriter) Done() bool {
	return cw.done
}

func (cw *ChannelizedWriter) Write(p []byte) (n int, err error) {
	if err := cw.ctx.Err(); err != nil {
		return 0, err
	}

	if cw.err != nil {
		return 0, cw.err
	}

	cw.writeQueue <- p
	return len(p), nil
}

func (cw *ChannelizedWriter) Error() error {
	if err := cw.ctx.Err(); err != nil {
		return err
	}
	return cw.err
}

func (cw *ChannelizedWriter) worker() {
	if cw.wg != nil {
		defer cw.wg.Done()
	}
	defer func() { cw.done = true }()

	for b := range cw.writeQueue {
		if err := cw.ctx.Err(); err != nil {
			cw.err = err
			return
		}

		if _, err := cw.writer.Write(b); err != nil {
			cw.err = err
			return
		}
	}
}
