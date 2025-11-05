package menu

import (
	"context"
	"fmt"
	"os"
	"time"
)

type Command struct {
	Key  string
	Name string
	Run  func(ctx context.Context) error
}

// Декоратор тайминга
func WithTiming(c Command) Command {
	return Command{
		Key:  c.Key,
		Name: c.Name,
		Run: func(ctx context.Context) error {
			start := time.Now()
			err := c.Run(ctx)
			dur := time.Since(start).Round(time.Millisecond)

			status := "OK"
			if err != nil {
				status = "ERR"
			}
			fmt.Printf("⏱ %s (%s): %s\n", c.Key, status, dur)

			if f, e := os.OpenFile("timings.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); e == nil {
				_, _ = fmt.Fprintf(f, "%s;%s;%s;%s\n", time.Now().UTC().Format(time.RFC3339), c.Key, status, dur)
				_ = f.Close()
			}
			return err
		},
	}
}
