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
		os.Exit(1)
	}

	app, err := di.Build(ctx)
	if err != nil {
		panic(err)
	}

	menu.Run(ctx, app.Menu, &app.Deps)
}
