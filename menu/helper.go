package menu

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"main/domain"
	"main/repo"
)

var stdin = bufio.NewReader(os.Stdin)

func readLine(prompt string) string {
	fmt.Print(prompt)
	s, _ := stdin.ReadString('\n')
	return strings.TrimSpace(s)
}

func readAmountAndDesc(prompt string) (decimal.Decimal, string, error) {
	for {
		raw := readLine(prompt)
		parts := strings.SplitN(raw, " ", 2)
		amtStr := parts[0]
		desc := ""
		if len(parts) == 2 {
			desc = strings.TrimSpace(parts[1])
		}
		amtStr = strings.ReplaceAll(amtStr, ",", ".")
		amt, err := decimal.NewFromString(amtStr)
		if err != nil {
			fmt.Println("Неверная сумма. Пример: 1500.00 Комментарий")
			continue
		}
		return amt, desc, nil
	}
}

func readDate(label string) (time.Time, error) {
	for {
		raw := readLine(label + " (YYYY-MM-DD, пусто = сегодня): ")
		if raw == "" {
			now := time.Now()
			return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local), nil
		}
		t, err := time.ParseInLocation("2006-01-02", raw, time.Local)
		if err != nil {
			fmt.Println("Формат даты неверный, ожидается YYYY-MM-DD")
			continue
		}

		t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
		now := time.Now().In(time.Local)
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
		if t.After(today) {
			fmt.Println("Дата в будущем не разрешена.")
			continue
		}
		return t, nil
	}
}

func readInt(prompt string) (int, error) {
	s := readLine(prompt)
	return strconv.Atoi(strings.TrimSpace(s))
}

func confirm(prompt string) bool {
	s := strings.ToLower(strings.TrimSpace(readLine(prompt + " [y/N]: ")))
	return s == "y" || s == "yes" || s == "д" || s == "да"
}

func chooseAccount(ctx context.Context, ar *repo.PgAccountRepo) (domain.AccountID, error) {
	accs, err := ar.List(ctx)
	if err != nil {
		return "", err
	}
	if len(accs) == 0 {
		return "", fmt.Errorf("нет счетов")
	}
	fmt.Println("=== Счета ===")
	for i, a := range accs {
		fmt.Printf("%d) %s | %s | %s\n", i+1, a.ID, a.Name, a.Balance.StringFixed(2))
	}
	n, err := readInt("Выбери №: ")
	if err != nil {
		return "", err
	}
	if n < 1 || n > len(accs) {
		return "", fmt.Errorf("неверный выбор")
	}
	return accs[n-1].ID, nil
}

func chooseAnyCategory(ctx context.Context, cr *repo.PgCategoryRepo) (domain.CategoryID, error) {
	cats, err := cr.List(ctx)
	if err != nil {
		return "", err
	}
	if len(cats) == 0 {
		return "", fmt.Errorf("нет категорий")
	}
	fmt.Println("=== Категории ===")
	for i, c := range cats {
		kind := "доход"
		if c.IsExpense() {
			kind = "расход"
		}
		fmt.Printf("%d) %s | %s [%s]\n", i+1, c.ID, c.Name, kind)
	}
	n, err := readInt("Выбери №: ")
	if err != nil {
		return "", err
	}
	if n < 1 || n > len(cats) {
		return "", fmt.Errorf("неверный выбор")
	}
	return cats[n-1].ID, nil
}

func chooseCategory(ctx context.Context, cr *repo.PgCategoryRepo, f domain.Factory, t domain.CategoryType) (domain.CategoryID, error) {
	cats, err := cr.List(ctx)
	if err != nil {
		return "", err
	}
	var opts []domain.Category
	for _, c := range cats {
		if c.Type == t {
			opts = append(opts, c)
		}
	}
	fmt.Println("=== Категории ===")
	for i, c := range opts {
		kind := "доход"
		if c.IsExpense() {
			kind = "расход"
		}
		fmt.Printf("%d) %s | %s [%s]\n", i+1, c.ID, c.Name, kind)
	}
	fmt.Println("0) + Новая категория")
	n, err := readInt("Выбери №: ")
	if err != nil {
		return "", err
	}
	if n == 0 {
		name := readLine("Название новой категории: ")
		c, err := f.NewCategory(name, t)
		if err != nil {
			return "", err
		}
		if err := cr.Create(ctx, c); err != nil {
			return "", err
		}
		return c.ID, nil
	}
	if n >= 1 && n <= len(opts) {
		return opts[n-1].ID, nil
	}
	return "", fmt.Errorf("неверный выбор")
}

func chooseCategoryOptional(ctx context.Context, cr *repo.PgCategoryRepo, f domain.Factory,
	t domain.CategoryType, current domain.CategoryID, allowEmpty bool,
) (domain.CategoryID, error) {
	cats, err := cr.List(ctx)
	if err != nil {
		return "", err
	}
	var opts []domain.Category
	for _, c := range cats {
		if c.Type == t {
			opts = append(opts, c)
		}
	}
	fmt.Println("=== Категории ===")
	for i, c := range opts {
		mark := " "
		if c.ID == current {
			mark = "*"
		}
		kind := "доход"
		if c.IsExpense() {
			kind = "расход"
		}
		fmt.Printf("%d) %s %s | %s [%s]\n", i+1, mark, c.ID, c.Name, kind)
	}
	if allowEmpty {
		fmt.Println("0) Оставить без изменений")
	} else {
		fmt.Println("0) + Новая категория")
	}
	n, err := readInt("Выбери №: ")
	if err != nil {
		return "", err
	}
	if n == 0 {
		if allowEmpty {
			return current, nil
		}
		name := readLine("Название новой категории: ")
		c, err := f.NewCategory(name, t)
		if err != nil {
			return "", err
		}
		if err := cr.Create(ctx, c); err != nil {
			return "", err
		}
		return c.ID, nil
	}
	if n >= 1 && n <= len(opts) {
		return opts[n-1].ID, nil
	}
	return "", fmt.Errorf("неверный выбор")
}

func chooseOperation(ctx context.Context, or *repo.PgOperationRepo, cr *repo.PgCategoryRepo, acc domain.AccountID, from, to time.Time) (domain.OperationID, error) {
	list, err := or.ListByAccount(ctx, acc, from, to)
	if err != nil {
		return "", err
	}
	if len(list) == 0 {
		return "", fmt.Errorf("операций не найдено")
	}
	fmt.Println("=== Операции ===")
	for i, o := range list {
		typ := "доход"
		if o.IsExpense() {
			typ = "расход"
		}
		catName := ""
		if o.Category != "" {
			if c, err := cr.Get(ctx, o.Category); err == nil {
				catName = c.Name
			}
		}
		fmt.Printf("%d) %s | %-6s | %8s | %-14s | %s\n",
			i+1, o.Date.Format("2006-01-02"), typ, o.Amount.StringFixed(2), catName, o.Description)
	}
	n, err := readInt("Выбери № операции: ")
	if err != nil {
		return "", err
	}
	if n < 1 || n > len(list) {
		return "", fmt.Errorf("неверный выбор")
	}
	return list[n-1].ID, nil
}

func readType() domain.CategoryType {
	for {
		raw := strings.ToLower(readLine("Тип категории (1=расход, 2=доход): "))
		switch raw {
		case "1", "расход", "expense", "e":
			return domain.CatExpense
		case "2", "доход", "income", "i":
			return domain.CatIncome
		default:
			fmt.Println("Выберите 1 или 2")
		}
	}
}

func readTypeOptional(def domain.OperationType) domain.OperationType {
	for {
		raw := strings.ToLower(readLine(fmt.Sprintf(
			"Тип операции (1=расход, 2=доход, пусто = %s): ",
			map[domain.OperationType]string{
				domain.OpExpense: "расход",
				domain.OpIncome:  "доход",
			}[def],
		)))
		switch raw {
		case "":
			return def
		case "1", "расход", "expense", "e":
			return domain.OpExpense
		case "2", "доход", "income", "i":
			return domain.OpIncome
		default:
			fmt.Println("Выберите 1 или 2, либо пусто")
		}
	}
}

func readAmountOptional(label string, def decimal.Decimal) (decimal.Decimal, error) {
	for {
		raw := readLine(fmt.Sprintf("%s (пусто = %s): ", label, def.StringFixed(2)))
		if strings.TrimSpace(raw) == "" {
			return def, nil
		}
		raw = strings.ReplaceAll(raw, ",", ".")
		d, err := decimal.NewFromString(raw)
		if err != nil {
			fmt.Println("Неверная сумма")
			continue
		}
		return d, nil
	}
}

func readDateOptional(def time.Time) (time.Time, error) {
	for {
		raw := readLine(fmt.Sprintf("Дата (YYYY-MM-DD, пусто = %s): ", def.Format("2006-01-02")))
		if strings.TrimSpace(raw) == "" {
			return def, nil
		}
		t, err := time.ParseInLocation("2006-01-02", raw, time.Local)
		if err != nil {
			fmt.Println("Формат даты неверный")
			continue
		}
		t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
		now := time.Now().In(time.Local)
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
		if t.After(today) {
			fmt.Println("Дата в будущем не разрешена.")
			continue
		}
		return t, nil
	}
}

func ensureCategory(ctx context.Context, cr *repo.PgCategoryRepo, f domain.Factory, name string, t domain.CategoryType) (domain.CategoryID, error) {
	cats, err := cr.List(ctx)
	if err != nil {
		return "", err
	}
	lname := strings.ToLower(strings.TrimSpace(name))
	for _, c := range cats {
		if c.Type == t && strings.ToLower(strings.TrimSpace(c.Name)) == lname {
			return c.ID, nil
		}
	}
	c, err := f.NewCategory(name, t)
	if err != nil {
		return "", err
	}
	if err := cr.Create(ctx, c); err != nil {
		return "", err
	}
	return c.ID, nil
}
