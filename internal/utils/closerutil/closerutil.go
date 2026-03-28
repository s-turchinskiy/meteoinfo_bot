// Package closerutil Для graceful shutdown
package closerutil

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/s-turchinskiy/meteoinfo_bot/internal/utils/reflectutil"
)

type FuncClose func(ctx context.Context) error

type Closer struct {
	mu      sync.Mutex
	funcs   []FuncClose
	timeout time.Duration
	log     Logger
}

type Logger interface {
	Info(args ...any)
}

func New(timeout time.Duration, log Logger) *Closer {
	return &Closer{
		timeout: timeout,
		log:     log,
	}
}

func (c *Closer) Add(f FuncClose) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.funcs = append(c.funcs, f)
}

func (c *Closer) close(ctx context.Context) (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var (
		msgs     = make([]string, 0, len(c.funcs))
		complete = make(chan struct{}, 1)
	)

	go func() {
		for _, f := range c.funcs {
			fname := reflectutil.GetFunctionName(f)
			c.log.Info("stopping " + fname)
			err = f(ctx)
			c.log.Info("stopped " + fname)
			if err != nil {
				msgs = append(msgs, fmt.Sprintf("[!] %v", err))
			}
		}

		complete <- struct{}{}
	}()

	select {
	case <-complete:
	case <-ctx.Done():
		return fmt.Errorf("shutdown cancelled: %v", ctx.Err())
	}

	if len(msgs) > 0 {
		return fmt.Errorf(
			"shutdown finished with error(s): \n%s",
			strings.Join(msgs, "\n"),
		)
	}

	return nil
}

func (c *Closer) Shutdown() error {
	c.log.Info("shutting down server gracefully")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	err := c.close(shutdownCtx)
	if err != nil {
		return fmt.Errorf("closerutil: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
	c.log.Info("server was shutdown successfully")

	return nil
}
