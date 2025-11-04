package menu

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"main/domain"
	"main/files"
	"main/repo"
	"main/service"
)

// Deps — всё, что нужно экзекьютору для действий меню.
type Deps struct {
	Pool    *pgxpool.Pool
	Factory domain.Factory

	AccRepo *repo.PgAccountRepo
	CatRepo *repo.PgCategoryRepo
	OpsRepo *repo.PgOperationRepo

	AccountID domain.AccountID // с каким счётом работаем
}

// Execute — выполняет действие по ключу меню.
func Execute(ctx context.Context, key string, d Deps) error {
	switch key {
	case "add_income":
		amt, desc, err := readAmountAndDesc("Сумма дохода (например 1500.00): ")
		if err != nil {
			return err
		}
		catID, err := ensureCategory(ctx, d.CatRepo, d.Factory, "Зарплата", domain.CatIncome)
		if err != nil {
			return err
		}
		opSvc := service.NewOperationService(d.Pool, d.Factory)
		if _, err := opSvc.ApplyOperation(ctx, domain.OpIncome, d.AccountID, amt, time.Now(), catID, desc); err != nil {
			return err
		}
		return printSummary(ctx, d, "Доход добавлен.")

	case "list_accounts":
		accs, err := d.AccRepo.List(ctx)
		if err != nil {
			return err
		}
		if len(accs) == 0 {
			fmt.Println("Счетов нет")
			return nil
		}
		fmt.Println("=== Счета ===")
		for _, a := range accs {
			fmt.Printf("- %s | %s | %s\n", a.ID, a.Name, a.Balance.StringFixed(2))
		}
		return nil

	case "add_expense":
		amt, desc, err := readAmountAndDesc("Сумма расхода (например 250.55): ")
		if err != nil {
			return err
		}
		catID, err := ensureCategory(ctx, d.CatRepo, d.Factory, "Кафе", domain.CatExpense)
		if err != nil {
			return err
		}
		opSvc := service.NewOperationService(d.Pool, d.Factory)
		if _, err := opSvc.ApplyOperation(ctx, domain.OpExpense, d.AccountID, amt, time.Now(), catID, desc); err != nil {
			return err
		}
		return printSummary(ctx, d, "Расход добавлен.")

	case "summary_30d":
		return printSummary(ctx, d, "")

	case "exit", "":
		os.Exit(0)
		return nil
	case "export_ops_json":
		path := readLine("Путь к файлу (напр. ops.json): ")
		if path == "" {
			path = "ops.json"
		}
		from, to := time.Now().AddDate(0, 0, -30), time.Now()
		if err := files.ExportOperationsJSON(ctx, d.OpsRepo, d.CatRepo, d.AccountID, from, to, path); err != nil {
			return err
		}
		fmt.Println("Экспортировано в", path)
		return nil

	case "import_ops_json":
		path := readLine("Путь к JSON для импорта: ")
		if path == "" {
			fmt.Println("Файл не указан")
			return nil
		}
		rows, err := files.ImportOperationsJSON(path)
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			fmt.Println("Нет записей для импорта")
			return nil
		}
		opSvc := service.NewOperationService(d.Pool, d.Factory)
		for _, r := range rows {
			catID, err := ensureCategory(ctx, d.CatRepo, d.Factory, r.Category, domain.CategoryType(r.Type))
			if err != nil {
				return err
			}
			if _, err := opSvc.ApplyOperation(ctx, domain.OperationType(r.Type), d.AccountID, r.Amount, r.Date, catID, r.Description); err != nil {
				return err
			}
		}
		return printSummary(ctx, d, fmt.Sprintf("Импортировано операций: %d.", len(rows)))

	case "export_ops_yaml":
		path := readLine("Путь к файлу (напр. ops.yaml): ")
		if path == "" {
			path = "ops.yaml"
		}
		from, to := time.Now().AddDate(0, 0, -30), time.Now()
		if err := files.ExportOperationsYAML(ctx, d.OpsRepo, d.CatRepo, d.AccountID, from, to, path); err != nil {
			return err
		}
		fmt.Println("Экспортировано в", path)
		return nil

	case "import_ops_yaml":
		path := readLine("Путь к YAML для импорта: ")
		if path == "" {
			fmt.Println("Файл не указан")
			return nil
		}
		rows, err := files.ImportOperationsYAML(path)
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			fmt.Println("Нет записей для импорта")
			return nil
		}
		opSvc := service.NewOperationService(d.Pool, d.Factory)
		for _, r := range rows {
			catID, err := ensureCategory(ctx, d.CatRepo, d.Factory, r.Category, domain.CategoryType(r.Type))
			if err != nil {
				return err
			}
			if _, err := opSvc.ApplyOperation(ctx, domain.OperationType(r.Type), d.AccountID, r.Amount, r.Date, catID, r.Description); err != nil {
				return err
			}
		}
		return printSummary(ctx, d, fmt.Sprintf("Импортировано операций: %d.", len(rows)))
	default:
		fmt.Println("Неизвестный пункт")
		return nil
	}
}

// ===== helpers (приватные для пакета menu) =====

func readAmountAndDesc(prompt string) (decimal.Decimal, string, error) {
	amtStr := readLine(prompt)
	amt, err := decimal.NewFromString(strings.TrimSpace(amtStr))
	if err != nil {
		return decimal.Zero, "", fmt.Errorf("некорректная сумма: %w", err)
	}
	desc := readLine("Описание (необязательно): ")
	return amt.Round(2), strings.TrimSpace(desc), nil
}

func readLine(prompt string) string {
	fmt.Print(prompt)
	in := bufio.NewReader(os.Stdin)
	s, _ := in.ReadString('\n')
	return strings.TrimSpace(s)
}

func ensureCategory(ctx context.Context, cr *repo.PgCategoryRepo, f domain.Factory, name string, t domain.CategoryType) (domain.CategoryID, error) {
	list, err := cr.List(ctx)
	if err != nil {
		return "", err
	}
	for _, c := range list {
		if strings.EqualFold(c.Name, name) && c.Type == t {
			return c.ID, nil
		}
	}
	cat, err := f.NewCategory(name, t)
	if err != nil {
		return "", err
	}
	if err := cr.Create(ctx, cat); err != nil {
		return "", err
	}
	return cat.ID, nil
}

func printSummary(ctx context.Context, d Deps, prefix string) error {
	if prefix != "" {
		fmt.Println(prefix)
	}
	got, err := d.AccRepo.Get(ctx, d.AccountID)
	if err != nil {
		return err
	}
	fmt.Printf("Баланс: %s\n", got.Balance.StringFixed(2))

	an := service.NewAnalyticsService(d.OpsRepo)
	from := time.Now().AddDate(0, 0, -30)
	to := time.Now()
	sum, err := an.SummaryByPeriod(ctx, d.AccountID, from, to)
	if err != nil {
		return err
	}
	fmt.Printf("Доход: %s | Расход: %s | Итого: %s\n",
		sum.Income.StringFixed(2),
		sum.Expense.StringFixed(2),
		sum.Net.StringFixed(2),
	)
	return nil
}
