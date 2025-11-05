package menu

import (
	"context"
	"fmt"
)

func Run(ctx context.Context, m Menu, d *Deps) {
	deps := d
	for {
		Draw(m)
		idx, err := ReadIndex(len(m.Items))
		if err != nil {
			fmt.Println("Неверный ввод")
			WaitEnter()
			fmt.Println()
			continue
		}

		// Берём выбранный пункт меню
		if idx < 1 || idx > len(m.Items) {
			fmt.Println("Неверный выбор")
			WaitEnter()
			fmt.Println()
			continue
		}
		item := m.Items[idx-1]
		key := item.Key
		title := item.Field

		if key == "exit" || key == "" {
			fmt.Println("Пока!")
			return
		}

		cmd := WithTiming(Command{
			Key:  key,
			Name: title,
			Run: func(ctx context.Context) error {
				return Execute(ctx, key, deps)
			},
		})

		if err := cmd.Run(ctx); err != nil {
			fmt.Println("Ошибка:", err)
		}

		WaitEnter()
		fmt.Println()
	}
}
