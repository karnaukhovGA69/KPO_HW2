package menu

import "context"

func BuildCommands(d Deps) map[string]Command {
	mk := func(key, name string) Command {
		return Command{
			Key:  key,
			Name: name,
			Run:  func(ctx context.Context) error { return Execute(ctx, key, &d) },
		}
	}

	return map[string]Command{
		"add_income":      mk("add_income", "Добавить доход"),
		"add_expense":     mk("add_expense", "Добавить расход"),
		"summary_30d":     mk("summary_30d", "Сводка за 30 дней"),
		"list_ops_30d":    mk("list_ops_30d", "Список операций за 30 дней"),
		"export_ops_csv":  mk("export_ops_csv", "Экспорт операций (CSV)"),
		"import_ops_csv":  mk("import_ops_csv", "Импорт операций (CSV)"),
		"export_ops_json": mk("export_ops_json", "Экспорт операций (JSON)"),
		"import_ops_json": mk("import_ops_json", "Импорт операций (JSON)"),
		"export_ops_yaml": mk("export_ops_yaml", "Экспорт операций (YAML)"),
		"import_ops_yaml": mk("import_ops_yaml", "Импорт операций (YAML)"),
		"add_category":    mk("add_category", "Создать категорию"),
		"list_categories": mk("list_categories", "Список категорий"),
		"exit":            mk("exit", "Выход"),
	}
}
