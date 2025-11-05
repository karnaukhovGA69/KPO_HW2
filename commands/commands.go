package commands

import (
	"context"
	"fmt"
	"os"
	"time"

	"main/menu"
)

// ===== Command паттерн =====

type Command interface {
	Name() string
	Execute(ctx context.Context, deps *menu.Deps) error
}

// Реализация команды поверх обычной функции-хендлера
type FuncCommand struct {
	name string
	fn   func(ctx context.Context, deps *menu.Deps) error
}

func NewFuncCommand(name string, fn func(ctx context.Context, deps *menu.Deps) error) FuncCommand {
	return FuncCommand{name: name, fn: fn}
}

func (c FuncCommand) Name() string { return c.name }
func (c FuncCommand) Execute(ctx context.Context, deps *menu.Deps) error {
	return c.fn(ctx, deps)
}

// ===== Декоратор тайминга для команд =====

type TimedCommand struct {
	inner Command
	// можно расширить: writer, репозиторий метрик и т.п.
}

func NewTimed(inner Command) TimedCommand { return TimedCommand{inner: inner} }

func (t TimedCommand) Name() string { return t.inner.Name() }
func (t TimedCommand) Execute(ctx context.Context, deps *menu.Deps) error {
	start := time.Now()
	err := t.inner.Execute(ctx, deps)
	dur := time.Since(start)

	// 1) лог в консоль
	status := "OK"
	if err != nil {
		status = "ERR"
	}
	fmt.Printf("⏱  %s: %s (%v)\n", t.inner.Name(), status, dur)

	// 2) лёгкий файл-лог (без падения при ошибках файловой системы)
	if f, e := os.OpenFile("timings.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); e == nil {
		// формат: 2025-11-05T12:34:56Z;CommandName;OK/ERR;123ms
		_, _ = fmt.Fprintf(f, "%s;%s;%s;%v\n", time.Now().UTC().Format(time.RFC3339), t.inner.Name(), status, dur)
		_ = f.Close()
	}
	return err
}

// ===== Декоратор тайминга для "старых" хендлеров =====

type Handler = func(ctx context.Context, deps *menu.Deps) error

func WithTiming(name string, h Handler) Handler {
	cmd := NewFuncCommand(name, h)
	timed := NewTimed(cmd)
	return func(ctx context.Context, deps *menu.Deps) error {
		return timed.Execute(ctx, deps)
	}
}
