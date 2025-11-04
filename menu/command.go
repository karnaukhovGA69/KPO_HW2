package menu

import (
	"context"
	"fmt"
	"time"
)

type Command struct {
	Key  string
	Name string
	Run  func(ctx context.Context) error
}

func WithTiming(c Command) Command {
	return Command{
		Key:  c.Key,
		Name: c.Name,
		Run: func(ctx context.Context) error {
			start := time.Now()
			err := c.Run(ctx)
			fmt.Printf("Время выполнения: %s\n", time.Since(start).Round(time.Millisecond))
			return err
		},
	}
}
