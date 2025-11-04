package menu

import (
	"context"
	"fmt"
)

// Run — крутит меню до "exit", вызывает обработчики и ждёт Enter.
func Run(ctx context.Context, m Menu, d Deps) {
	for {
		Draw(m)
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
		if err := Execute(ctx, key, d); err != nil {
			fmt.Println("Ошибка:", err)
		}
		WaitEnter()
		fmt.Println()
	}
}
