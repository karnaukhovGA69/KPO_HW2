package menu

import "fmt"

// Draw — напечатать меню нумерованным списком (для выбора цифрой).
func Draw(m Menu) {
	fmt.Println("==== Меню ====")
	for i, it := range m.Items {
		fmt.Printf("%d) %s\n", i+1, it.Field)
	}
}
