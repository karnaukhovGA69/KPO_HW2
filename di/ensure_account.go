package di

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"main/domain"
	"main/repo"
	"main/state"
)

func ensureActiveAccount(ctx context.Context, accRepo *repo.PgAccountRepo, f domain.Factory) (domain.AccountID, string, error) {
	accs, err := accRepo.List(ctx)
	if err != nil {
		return "", "", err
	}

	if len(accs) == 0 {
		acc, err := f.NewBankAccount("Основной")
		if err != nil {
			return "", "", err
		}
		if err := accRepo.Create(ctx, acc); err != nil {
			return "", "", err
		}
		_ = state.SaveAccountID(string(acc.ID))
		return acc.ID, acc.Name, nil
	}

	if saved, err := state.LoadAccountID(); err == nil && saved != "" {
		for _, a := range accs {
			if string(a.ID) == saved {
				return a.ID, a.Name, nil
			}
		}
	}

	fmt.Println("=== Выберите счёт ===")
	for i, a := range accs {
		fmt.Printf("%d) %s | %s\n", i+1, a.Name, a.Balance.StringFixed(2))
	}
	fmt.Println("0) Создать новый счёт")

	choice := readLineSimple("Ваш выбор: ")
	n, _ := strconv.Atoi(choice)

	if n == 0 {
		name := readLineSimple("Имя нового счёта: ")
		acc, err := f.NewBankAccount(name)
		if err != nil {
			return "", "", err
		}
		if err := accRepo.Create(ctx, acc); err != nil {
			return "", "", err
		}
		_ = state.SaveAccountID(string(acc.ID))
		return acc.ID, acc.Name, nil
	}

	if n >= 1 && n <= len(accs) {
		_ = state.SaveAccountID(string(accs[n-1].ID))
		return accs[n-1].ID, accs[n-1].Name, nil
	}

	_ = state.SaveAccountID(string(accs[0].ID))
	return accs[0].ID, accs[0].Name, nil
}

func readLineSimple(prompt string) string {
	fmt.Print(prompt)
	in := bufio.NewReader(os.Stdin)
	s, _ := in.ReadString('\n')
	return strings.TrimSpace(s)
}
