package menu

import "fmt"

func Draw(m Menu) {
	fmt.Println("==== Меню ====")
	for i, it := range m.Items {
		fmt.Printf("%d) %s\n", i+1, it.Field)
	}
}
