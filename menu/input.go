package menu

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func ReadIndex(count int) (int, error) {
	in := bufio.NewReader(os.Stdin)
	fmt.Printf("Выбор (1..%d): ", count)
	s, _ := in.ReadString('\n')
	s = strings.TrimSpace(s)
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 || n > count {
		return 0, fmt.Errorf("неверный ввод")
	}
	return n, nil
}

func WaitEnter() {
	fmt.Print("\nНажмите Enter для продолжения...")
	bufio.NewReader(os.Stdin).ReadString('\n')
}
