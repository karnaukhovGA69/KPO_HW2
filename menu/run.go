package menu

import (
	"context"
	"fmt"
)

func Run(ctx context.Context, m Menu, d *Deps) {
	deps := d // локальная копия, будем менять внутри
	for {
		Draw(m)
		// Если у тебя есть ReadChoice — используй его. Иначе оставь ReadIndex/KeyAt:
		idx, err := ReadIndex(len(m.Items))
		if err != nil {
			fmt.Println("Неверный ввод")
			WaitEnter()
			fmt.Println()
			continue
		}
		key := m.KeyAt(idx)
		if key == "exit" || key == "" {
			fmt.Println("Пока!")
			return
		}
		if err := Execute(ctx, key, deps); err != nil {
			fmt.Println("Ошибка:", err)
		}
		WaitEnter()
		fmt.Println()
	}
}
