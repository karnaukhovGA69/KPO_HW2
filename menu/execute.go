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
	"main/files"
	"main/repo"
	"main/service"
	"main/state"
)

func Execute(ctx context.Context, key string, d *Deps) error {
	switch key {

	// ===== ОПЕРАЦИИ =====
	case "add_income":
		amt, desc, err := readAmountAndDesc("Сумма дохода (например 1500.00): ")
		if err != nil {
			return err
		}
		when, err := readDate("Дата дохода")
		if err != nil {
			return err
		}
		catID, err := chooseCategory(ctx, d.CatRepo, d.Factory, domain.CatIncome)
		if err != nil {
			return err
		}
		opSvc := service.NewOperationService(d.Pool, d.Factory)
		if _, err := opSvc.ApplyOperation(ctx, domain.OpIncome, d.AccountID, amt, when, catID, desc); err != nil {
			return err
		}
		return printSummary(ctx, *d, "Доход добавлен.")

	case "add_expense":
		amt, desc, err := readAmountAndDesc("Сумма расхода (например 250.55): ")
		if err != nil {
			return err
		}
		when, err := readDate("Дата расхода")
		if err != nil {
			return err
		}
		catID, err := chooseCategory(ctx, d.CatRepo, d.Factory, domain.CatExpense)
		if err != nil {
			return err
		}
		opSvc := service.NewOperationService(d.Pool, d.Factory)
		if _, err := opSvc.ApplyOperation(ctx, domain.OpExpense, d.AccountID, amt, when, catID, desc); err != nil {
			return err
		}
		return printSummary(ctx, *d, "Расход добавлен.")

	case "list_ops_30d":
		from, to := time.Now().AddDate(0, 0, -30), time.Now()
		list, err := d.OpsRepo.ListByAccount(ctx, d.AccountID, from, to)
		if err != nil {
			return err
		}
		if len(list) == 0 {
			fmt.Println("Операций нет за период")
			return nil
		}
		fmt.Println("=== Операции за 30 дней ===")
		for _, o := range list {
			typ := "доход"
			if o.IsExpense() {
				typ = "расход"
			}
			fmt.Printf("%s | %-6s | %8s | %s\n",
				o.Date.Format("2006-01-02"), typ, o.Amount.StringFixed(2), o.Description)
		}
		return nil

	case "list_ops_period":
		from, _ := readDate("Дата ОТ")
		to, _ := readDate("Дата ДО")
		list, err := d.OpsRepo.ListByAccount(ctx, d.AccountID, from, to)
		if err != nil {
			return err
		}
		if len(list) == 0 {
			fmt.Println("Операций нет за период")
			return nil
		}
		fmt.Println("=== Операции ===")
		for _, o := range list {
			typ := "доход"
			if o.IsExpense() {
				typ = "расход"
			}
			fmt.Printf("%s | %-6s | %8s | %s\n",
				o.Date.Format("2006-01-02"), typ, o.Amount.StringFixed(2), o.Description)
		}
		return nil

	// ===== СВОДКИ =====
	case "summary_30d":
		return printSummary(ctx, *d, "")

	case "summary_period":
		from, _ := readDate("Дата ОТ")
		to, _ := readDate("Дата ДО")
		an := service.NewAnalyticsService(d.OpsRepo)
		sum, err := an.SummaryByPeriod(ctx, d.AccountID, from, to)
		if err != nil {
			return err
		}
		fmt.Printf("Доход: %s | Расход: %s | Итого: %s\n",
			sum.Income.StringFixed(2), sum.Expense.StringFixed(2), sum.Net.StringFixed(2))
		return nil

	// ===== КАТЕГОРИИ =====
	case "add_category":
		name := readLine("Название категории: ")
		t := readType()
		c, err := d.Factory.NewCategory(name, t)
		if err != nil {
			return err
		}
		if err := d.CatRepo.Create(ctx, c); err != nil {
			return err
		}
		fmt.Println("Категория создана.")
		return nil

	case "list_categories":
		cats, err := d.CatRepo.List(ctx)
		if err != nil {
			return err
		}
		if len(cats) == 0 {
			fmt.Println("Категорий нет")
			return nil
		}
		fmt.Println("=== Категории ===")
		for _, c := range cats {
			kind := "доход"
			if c.IsExpense() {
				kind = "расход"
			}
			fmt.Printf("- %s [%s]\n", c.Name, kind)
		}
		return nil

	// ===== СЧЕТА =====
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

	case "create_account":
		name := readLine("Имя счёта: ")
		acc, err := d.Factory.NewBankAccount(name)
		if err != nil {
			return err
		}
		if err := d.AccRepo.Create(ctx, acc); err != nil {
			return err
		}
		fmt.Println("Счёт создан:", acc.Name, acc.ID)
		if confirm("Сделать активным?") {
			d.AccountID = acc.ID
			_ = state.SaveAccountID(string(d.AccountID))
			fmt.Println("Активный счёт обновлён.")
		}
		return nil

	case "select_account":
		id, err := chooseAccount(ctx, d.AccRepo)
		if err != nil {
			return err
		}
		d.AccountID = id
		_ = state.SaveAccountID(string(d.AccountID))
		fmt.Println("Активный счёт обновлён:", id)
		return nil

	case "delete_account":
		id, err := chooseAccount(ctx, d.AccRepo)
		if err != nil {
			return err
		}
		if !confirm("Точно удалить счёт и его операции?") {
			return nil
		}
		if err := d.AccRepo.Delete(ctx, id); err != nil {
			return err
		}
		fmt.Println("Счёт удалён.")
		if id == d.AccountID {
			accs, _ := d.AccRepo.List(ctx)
			if len(accs) > 0 {
				d.AccountID = accs[0].ID
				_ = state.SaveAccountID(string(d.AccountID))
				fmt.Println("Новый активный счёт:", accs[0].Name)
			} else {
				d.AccountID = ""
				_ = state.SaveAccountID("")
				fmt.Println("Нет активного счёта.")
			}
		}
		return nil

	// ===== ЭКСПОРТ/ИМПОРТ =====
	case "export_ops_csv":
		path := readLine("Путь к файлу (напр. ops.csv): ")
		if path == "" {
			path = "ops.csv"
		}
		from, to := time.Now().AddDate(0, 0, -30), time.Now()
		if err := files.ExportOperationsCSV(ctx, d.OpsRepo, d.CatRepo, d.AccountID, from, to, path); err != nil {
			return err
		}
		fmt.Println("Экспортировано в", path)
		return nil

	case "import_ops_csv":
		path := readLine("Путь к CSV для импорта: ")
		if path == "" {
			fmt.Println("Файл не указан")
			return nil
		}
		rows, err := files.ImportOperationsCSV(path)
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
		return printSummary(ctx, *d, fmt.Sprintf("Импортировано операций: %d.", len(rows)))
	case "edit_op_30d":
		from, to := time.Now().AddDate(0, 0, -30), time.Now()
		// выберем операцию
		opID, err := chooseOperation(ctx, d.OpsRepo, d.CatRepo, d.AccountID, from, to)
		if err != nil {
			return err
		}

		// загрузим текущие поля
		old, err := d.OpsRepo.Get(ctx, opID)
		if err != nil {
			return err
		}

		// спросим новые значения (пусто = оставить)
		newType := readTypeOptional(old.Type)
		newAmt, err := readAmountOptional("Сумма", old.Amount)
		if err != nil {
			return err
		}
		newDate, err := readDateOptional(old.Date)
		if err != nil {
			return err
		}

		// категория: если тип поменялся — обязателен выбор новой; иначе можно Enter оставить
		var newCat domain.CategoryID
		if newType != old.Type {
			newCat, err = chooseCategoryOptional(ctx, d.CatRepo, d.Factory,
				domain.CategoryType(newType), old.Category, false)
			if err != nil {
				return err
			}
		} else {
			newCat, err = chooseCategoryOptional(ctx, d.CatRepo, d.Factory,
				domain.CategoryType(newType), old.Category, true)
			if err != nil {
				return err
			}
		}

		// описание — простое (пусто = оставить)
		desc := readLine(fmt.Sprintf("Описание (пусто = оставить: %q): ", old.Description))
		if desc == "" {
			desc = old.Description
		}

		// применяем через сервис
		opSvc := service.NewOperationService(d.Pool, d.Factory)
		if err := opSvc.UpdateOperation(ctx, opID, newType, newAmt, newDate, newCat, desc); err != nil {
			return err
		}
		return printSummary(ctx, *d, "Операция обновлена.")
	case "delete_op_30d":
		from, to := time.Now().AddDate(0, 0, -30), time.Now()
		opID, err := chooseOperation(ctx, d.OpsRepo, d.CatRepo, d.AccountID, from, to)
		if err != nil {
			return err
		}
		if !confirm("Удалить выбранную операцию?") {
			return nil
		}
		opSvc := service.NewOperationService(d.Pool, d.Factory)
		if err := opSvc.RemoveOperation(ctx, opID); err != nil {
			return err
		}
		return printSummary(ctx, *d, "Операция удалена.")
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
	case "summary_cat_30d":
		from, to := time.Now().AddDate(0, 0, -30), time.Now()
		an := service.NewAnalyticsService(d.OpsRepo)
		list, err := an.ByCategory(ctx, d.AccountID, from, to)
		if err != nil {
			return err
		}
		if len(list) == 0 {
			fmt.Println("Нет данных за период")
			return nil
		}
		fmt.Println("=== Сводка по категориям (30 дней) ===")
		for _, cs := range list {
			tag := "доход"
			if cs.Type == domain.CatExpense {
				tag = "расход"
			}
			fmt.Printf("%-20s [%s]  Доход: %8s  Расход: %8s  Итого: %8s\n",
				cs.Name, tag, cs.Income.StringFixed(2), cs.Expense.StringFixed(2), cs.Net.StringFixed(2))
		}
		return nil

	case "summary_cat_period":
		from, _ := readDate("Дата ОТ")
		to, _ := readDate("Дата ДО")
		an := service.NewAnalyticsService(d.OpsRepo)
		list, err := an.ByCategory(ctx, d.AccountID, from, to)
		if err != nil {
			return err
		}
		if len(list) == 0 {
			fmt.Println("Нет данных за период")
			return nil
		}
		fmt.Println("=== Сводка по категориям ===")
		for _, cs := range list {
			tag := "доход"
			if cs.Type == domain.CatExpense {
				tag = "расход"
			}
			fmt.Printf("%-20s [%s]  Доход: %8s  Расход: %8s  Итого: %8s\n",
				cs.Name, tag, cs.Income.StringFixed(2), cs.Expense.StringFixed(2), cs.Net.StringFixed(2))
		}
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
		return printSummary(ctx, *d, fmt.Sprintf("Импортировано операций: %d.", len(rows)))
	case "rename_account":
		newName := readLine("Новое имя активного счёта: ")
		if strings.TrimSpace(newName) == "" {
			fmt.Println("Имя пустое")
			return nil
		}
		if err := d.AccRepo.UpdateName(ctx, d.AccountID, newName); err != nil {
			return err
		}
		fmt.Println("Счёт переименован.")
		return nil

	// ПЕРЕИМЕНОВАТЬ КАТЕГОРИЮ
	case "rename_category":
		catID, err := chooseAnyCategory(ctx, d.CatRepo)
		if err != nil {
			return err
		}
		newName := readLine("Новое имя категории: ")
		if strings.TrimSpace(newName) == "" {
			fmt.Println("Имя пустое")
			return nil
		}
		if err := d.CatRepo.UpdateName(ctx, catID, newName); err != nil {
			return err
		}
		fmt.Println("Категория переименована.")
		return nil

	// УДАЛИТЬ КАТЕГОРИЮ (если нет операций)
	case "delete_category":
		catID, err := chooseAnyCategory(ctx, d.CatRepo)
		if err != nil {
			return err
		}
		has, err := d.CatRepo.HasOperations(ctx, catID)
		if err != nil {
			return err
		}
		if has {
			fmt.Println("Нельзя удалить: в категории есть операции. Сначала удалите/перенесите операции.")
			return nil
		}
		if !confirm("Точно удалить категорию?") {
			return nil
		}
		if err := d.CatRepo.Delete(ctx, catID); err != nil {
			return err
		}
		fmt.Println("Категория удалена.")
		return nil
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
		return printSummary(ctx, *d, fmt.Sprintf("Импортировано операций: %d.", len(rows)))

	// ===== ПРОЧЕЕ =====
	case "exit", "":
		return nil

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
func readDate(prompt string) (time.Time, error) {
	for {
		s := strings.TrimSpace(readLine(prompt + " (YYYY-MM-DD, пусто = сегодня): "))
		if s == "" {
			return time.Now(), nil
		}
		dt, err := time.Parse("2006-01-02", s)
		if err == nil {
			return dt, nil
		}
		fmt.Println("Неверный формат даты, пример: 2025-11-04")
	}
}

func chooseCategory(ctx context.Context, cr *repo.PgCategoryRepo, f domain.Factory, t domain.CategoryType) (domain.CategoryID, error) {
	list, err := cr.List(ctx)
	if err != nil {
		return "", err
	}
	var opts []domain.Category
	for _, c := range list {
		if c.Type == t {
			opts = append(opts, c)
		}
	}

	fmt.Println("=== Категории ===")
	for i, c := range opts {
		fmt.Printf("%d) %s\n", i+1, c.Name)
	}
	fmt.Println("0) Создать новую")

	nStr := readLine("Выбор: ")
	n, _ := strconv.Atoi(nStr)
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
		return "", fmt.Errorf("счетов нет")
	}
	fmt.Println("=== Счета ===")
	for i, a := range accs {
		fmt.Printf("%d) %s | %s | %s\n", i+1, a.ID, a.Name, a.Balance.StringFixed(2))
	}
	nStr := readLine("Выбор: ")
	n, _ := strconv.Atoi(nStr)
	if n < 1 || n > len(accs) {
		return "", fmt.Errorf("неверный выбор")
	}
	return accs[n-1].ID, nil
}
func readType() domain.CategoryType {
	for {
		s := strings.TrimSpace(readLine("Тип (-1=расход, 1=доход): "))
		if s == "-1" {
			return domain.CatExpense
		}
		if s == "1" {
			return domain.CatIncome
		}
		fmt.Println("Введите -1 или 1")
	}
}
func chooseOperation(ctx context.Context, ops *repo.PgOperationRepo, cats *repo.PgCategoryRepo, accID domain.AccountID, from, to time.Time) (domain.OperationID, error) {
	list, err := ops.ListByAccount(ctx, accID, from, to)
	if err != nil {
		return "", err
	}
	if len(list) == 0 {
		return "", fmt.Errorf("Операций нет за период")
	}

	// кэш категорий
	cmap := map[domain.CategoryID]string{}
	getCat := func(id domain.CategoryID) string {
		if n, ok := cmap[id]; ok {
			return n
		}
		c, err := cats.Get(ctx, id)
		if err != nil {
			return ""
		}
		cmap[id] = c.Name
		return c.Name
	}

	fmt.Println("=== Операции ===")
	for i, o := range list {
		typ := "доход"
		if o.IsExpense() {
			typ = "расход"
		}
		fmt.Printf("%d) %s | %-6s | %8s | %-12s | %s\n",
			i+1, o.Date.Format("2006-01-02"), typ, o.Amount.StringFixed(2), getCat(o.Category), o.Description)
	}
	nStr := readLine("Выберите номер операции: ")
	n, _ := strconv.Atoi(nStr)
	if n < 1 || n > len(list) {
		return "", fmt.Errorf("неверный выбор")
	}
	return list[n-1].ID, nil
}
func chooseAnyCategory(ctx context.Context, cr *repo.PgCategoryRepo) (domain.CategoryID, error) {
	cats, err := cr.List(ctx)
	if err != nil {
		return "", err
	}
	if len(cats) == 0 {
		return "", fmt.Errorf("категорий нет")
	}
	fmt.Println("=== Категории ===")
	for i, c := range cats {
		tag := "доход"
		if c.IsExpense() {
			tag = "расход"
		}
		fmt.Printf("%d) %s [%s]\n", i+1, c.Name, tag)
	}
	nStr := readLine("Выбор: ")
	n, _ := strconv.Atoi(nStr)
	if n < 1 || n > len(cats) {
		return "", fmt.Errorf("неверный выбор")
	}
	return cats[n-1].ID, nil
}
func readAmountOptional(prompt string, def decimal.Decimal) (decimal.Decimal, error) {
	s := strings.TrimSpace(readLine(fmt.Sprintf("%s (пусто = %s): ", prompt, def.StringFixed(2))))
	if s == "" {
		return def, nil
	}
	amt, err := decimal.NewFromString(s)
	if err != nil {
		return decimal.Zero, fmt.Errorf("некорректная сумма: %w", err)
	}
	return amt.Round(2), nil
}

// опциональный ввод типа (-1/1) (пусто = оставить как есть)
func readTypeOptional(def domain.OperationType) domain.OperationType {
	for {
		s := strings.TrimSpace(readLine(fmt.Sprintf("Тип (-1=расход, 1=доход, пусто=%d): ", int(def))))
		if s == "" {
			return def
		}
		if s == "-1" {
			return domain.OpExpense
		}
		if s == "1" {
			return domain.OpIncome
		}
		fmt.Println("Введите -1, 1 или пусто")
	}
}

// опциональный ввод даты (пусто = оставить как есть)
func readDateOptional(def time.Time) (time.Time, error) {
	s := strings.TrimSpace(readLine(fmt.Sprintf("Дата (YYYY-MM-DD, пусто=%s): ", def.Format("2006-01-02"))))
	if s == "" {
		return def, nil
	}
	dt, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}, fmt.Errorf("неверный формат даты")
	}
	return dt, nil
}

// выбрать категорию по типу; если allowEmpty=true и пользователь нажал пусто — вернуть defID
func chooseCategoryOptional(ctx context.Context, cr *repo.PgCategoryRepo, f domain.Factory, t domain.CategoryType, defID domain.CategoryID, allowEmpty bool) (domain.CategoryID, error) {
	list, err := cr.List(ctx)
	if err != nil {
		return "", err
	}
	var opts []domain.Category
	for _, c := range list {
		if c.Type == t {
			opts = append(opts, c)
		}
	}

	fmt.Println("=== Категории ===")
	for i, c := range opts {
		fmt.Printf("%d) %s\n", i+1, c.Name)
	}
	if allowEmpty {
		fmt.Println("Enter) Оставить без изменений")
	}
	fmt.Println("0) Создать новую")

	s := readLine("Выбор: ")
	if s == "" && allowEmpty {
		return defID, nil
	}
	n, _ := strconv.Atoi(s)
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
