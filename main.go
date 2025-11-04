package main

import (
	"context"
	"fmt"
	"os"

	"main/di"
	"main/menu"
)

func main() {
	ctx := context.Background()
	if os.Getenv("DATABASE_URL") == "" {
		fmt.Println("ERROR: set DATABASE_URL")
		return
	}

	c, err := di.Build(ctx)
	if err != nil {
		panic(err)
	}

	// Достаём собранное приложение из контейнера и запускаем меню
	if err := c.Invoke(func(app *di.App) {
		menu.Run(ctx, app.Menu, &app.Deps)
	}); err != nil {
		panic(err)
	}
}
