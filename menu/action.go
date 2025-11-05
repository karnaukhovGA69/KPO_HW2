package menu

import (
	"context"
	"fmt"
	"strings"
	"time"

	"main/domain"
	"main/facade"
	"main/files"
	"main/state"

	"github.com/shopspring/decimal"
)

func actionAddIncome(ctx context.Context, d *Deps) error {
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
	cat, err := d.CatRepo.Get(ctx, catID)
	if err != nil {
		return err
	}

	_, err = d.Op.AddIncome(ctx, facade.AddOpInput{
		AccountID:    d.AccountID,
		Amount:       amt,
		When:         when,
		CategoryName: cat.Name,
		Description:  desc,
	})
	if err != nil {
		return err
	}
	return printSummary(ctx, *d, "Доход добавлен.")
}

func actionAddExpense(ctx context.Context, d *Deps) error {
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
	cat, err := d.CatRepo.Get(ctx, catID)
	if err != nil {
		return err
	}

	_, err = d.Op.AddExpense(ctx, facade.AddOpInput{
		AccountID:    d.AccountID,
		Amount:       amt,
		When:         when,
		CategoryName: cat.Name,
		Description:  desc,
	})
	if err != nil {
		return err
	}
	return printSummary(ctx, *d, "Расход добавлен.")
}

func actionListOps30d(ctx context.Context, d *Deps) error {
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
}

func actionListOpsPeriod(ctx context.Context, d *Deps) error {
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
}

func actionSummary30d(ctx context.Context, d *Deps) error {
	return printSummary(ctx, *d, "")
}

func actionSummaryPeriod(ctx context.Context, d *Deps) error {
	from, _ := readDate("Дата ОТ")
	to, _ := readDate("Дата ДО")
	sum, err := d.Ana.Summary(ctx, d.AccountID, from, to)
	if err != nil {
		return err
	}
	fmt.Printf("Доход: %s | Расход: %s | Итого: %s\n",
		sum.Income.StringFixed(2), sum.Expense.StringFixed(2), sum.Net.StringFixed(2))
	return nil
}

func actionAddCategory(ctx context.Context, d *Deps) error {
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
}

func actionListCategories(ctx context.Context, d *Deps) error {
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
}

func actionRenameCategory(ctx context.Context, d *Deps) error {
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
}

func actionDeleteCategory(ctx context.Context, d *Deps) error {
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
}

func actionListAccounts(ctx context.Context, d *Deps) error {
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
}

func actionCreateAccount(ctx context.Context, d *Deps) error {
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
}

func actionSelectAccount(ctx context.Context, d *Deps) error {
	id, err := chooseAccount(ctx, d.AccRepo)
	if err != nil {
		return err
	}
	d.AccountID = id
	_ = state.SaveAccountID(string(d.AccountID))
	fmt.Println("Активный счёт обновлён:", id)
	return nil
}

func actionDeleteAccount(ctx context.Context, d *Deps) error {
	id, err := chooseAccount(ctx, d.AccRepo)
	if err != nil {
		return err
	}
	if !confirm("Удалить счёт и его операции?") {
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
}

func actionRenameAccount(ctx context.Context, d *Deps) error {
	newName := readLine("Новое имя активного счёта: ")
	if strings.TrimSpace(newName) == "" {
		fmt.Println("Имя пустое")
		return nil
	}
	if err := d.Acc.Rename(ctx, d.AccountID, newName); err != nil {
		return err
	}
	fmt.Println("Счёт переименован.")
	return nil
}

func actionExportOpsCSV(ctx context.Context, d *Deps) error {
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
}

func actionExportOpsJSON(ctx context.Context, d *Deps) error {
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
}

func actionExportOpsYAML(ctx context.Context, d *Deps) error {
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
}

func actionImportOpsCSV(ctx context.Context, d *Deps) error {
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
	for _, r := range rows {
		in := facade.AddOpInput{
			AccountID:    d.AccountID,
			Amount:       r.Amount,
			When:         r.Date,
			CategoryName: r.Category,
			Description:  r.Description,
		}
		if r.Type >= 0 {
			_, err = d.Op.AddIncome(ctx, in)
		} else {
			_, err = d.Op.AddExpense(ctx, in)
		}
		if err != nil {
			return err
		}
	}
	return printSummary(ctx, *d, fmt.Sprintf("Импортировано операций: %d.", len(rows)))
}

func actionImportOpsJSON(ctx context.Context, d *Deps) error {
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
	for _, r := range rows {
		in := facade.AddOpInput{
			AccountID:    d.AccountID,
			Amount:       r.Amount,
			When:         r.Date,
			CategoryName: r.Category,
			Description:  r.Description,
		}
		if r.Type >= 0 {
			_, err = d.Op.AddIncome(ctx, in)
		} else {
			_, err = d.Op.AddExpense(ctx, in)
		}
		if err != nil {
			return err
		}
	}
	return printSummary(ctx, *d, fmt.Sprintf("Импортировано операций: %d.", len(rows)))
}

func actionImportOpsYAML(ctx context.Context, d *Deps) error {
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
	for _, r := range rows {
		in := facade.AddOpInput{
			AccountID:    d.AccountID,
			Amount:       r.Amount,
			When:         r.Date,
			CategoryName: r.Category,
			Description:  r.Description,
		}
		if r.Type >= 0 {
			_, err = d.Op.AddIncome(ctx, in)
		} else {
			_, err = d.Op.AddExpense(ctx, in)
		}
		if err != nil {
			return err
		}
	}
	return printSummary(ctx, *d, fmt.Sprintf("Импортировано операций: %d.", len(rows)))
}

func actionEditOp30d(ctx context.Context, d *Deps) error {
	from, to := time.Now().AddDate(0, 0, -30), time.Now()
	opID, err := chooseOperation(ctx, d.OpsRepo, d.CatRepo, d.AccountID, from, to)
	if err != nil {
		return err
	}

	old, err := d.OpsRepo.Get(ctx, opID)
	if err != nil {
		return err
	}

	newType := readTypeOptional(old.Type)
	newAmt, err := readAmountOptional("Сумма", old.Amount)
	if err != nil {
		return err
	}
	newDate, err := readDateOptional(old.Date)
	if err != nil {
		return err
	}

	var newAmtPtr *decimal.Decimal
	if !newAmt.Equal(old.Amount) {
		v := newAmt
		newAmtPtr = &v
	}
	var newDatePtr *time.Time
	if !newDate.Equal(old.Date) {
		v := newDate
		newDatePtr = &v
	}

	var newCatNamePtr *string
	if newType != old.Type {
		newCatID, err := chooseCategoryOptional(ctx, d.CatRepo, d.Factory,
			domain.CategoryType(newType), old.Category, false)
		if err != nil {
			return err
		}
		c, err := d.CatRepo.Get(ctx, newCatID)
		if err != nil {
			return err
		}
		n := c.Name
		newCatNamePtr = &n
	} else {
		newCatID, err := chooseCategoryOptional(ctx, d.CatRepo, d.Factory,
			domain.CategoryType(newType), old.Category, true)
		if err != nil {
			return err
		}
		c, err := d.CatRepo.Get(ctx, newCatID)
		if err != nil {
			return err
		}
		n := c.Name
		newCatNamePtr = &n
	}

	var forced *domain.CategoryType
	if newType != old.Type {
		switch newType {
		case domain.OpIncome:
			t := domain.CatIncome
			forced = &t
		case domain.OpExpense:
			t := domain.CatExpense
			forced = &t
		}
	}

	op, err := d.Op.Edit(ctx, facade.EditOpInput{
		OperationID: old.ID,
		NewAmount:   newAmtPtr,
		NewWhen:     newDatePtr,
		NewCategory: newCatNamePtr,
		NewDesc:     strPtrOrNil(readLine(fmt.Sprintf("Описание (пусто = оставить: %q): ", old.Description))),
		ForcedType:  forced,
	})
	if err != nil {
		return err
	}
	fmt.Printf("Операция обновлена: %s  %s  %s  %s\n",
		op.ID, op.Date.Format("2006-01-02"), op.Amount.StringFixed(2), op.Description)
	return printSummary(ctx, *d, "")
}

func actionDeleteOp30d(ctx context.Context, d *Deps) error {
	from, to := time.Now().AddDate(0, 0, -30), time.Now()
	opID, err := chooseOperation(ctx, d.OpsRepo, d.CatRepo, d.AccountID, from, to)
	if err != nil {
		return err
	}
	if !confirm("Удалить выбранную операцию?") {
		return nil
	}
	if err := d.Op.Delete(ctx, opID); err != nil {
		return err
	}
	return printSummary(ctx, *d, "Операция удалена.")
}
