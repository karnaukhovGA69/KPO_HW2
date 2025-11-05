package menu

import (
	"context"
	"fmt"
)

func Execute(ctx context.Context, key string, d *Deps) error {
	switch key {
	case "add_income":
		if err := actionAddIncome(ctx, d); err != nil {
			return err
		}
	case "add_expense":
		if err := actionAddExpense(ctx, d); err != nil {
			return err
		}
	case "list_ops_30d":
		if err := actionListOps30d(ctx, d); err != nil {
			return err
		}
	case "list_ops_period":
		if err := actionListOpsPeriod(ctx, d); err != nil {
			return err
		}
	case "summary_30d":
		if err := actionSummary30d(ctx, d); err != nil {
			return err
		}
	case "summary_period":
		if err := actionSummaryPeriod(ctx, d); err != nil {
			return err
		}
	case "add_category":
		if err := actionAddCategory(ctx, d); err != nil {
			return err
		}
	case "list_categories":
		if err := actionListCategories(ctx, d); err != nil {
			return err
		}
	case "list_accounts":
		if err := actionListAccounts(ctx, d); err != nil {
			return err
		}
	case "create_account":
		if err := actionCreateAccount(ctx, d); err != nil {
			return err
		}
	case "select_account":
		if err := actionSelectAccount(ctx, d); err != nil {
			return err
		}
	case "delete_account":
		if err := actionDeleteAccount(ctx, d); err != nil {
			return err
		}
	case "export_ops_csv":
		if err := actionExportOpsCSV(ctx, d); err != nil {
			return err
		}
	case "import_ops_csv":
		if err := actionImportOpsCSV(ctx, d); err != nil {
			return err
		}
	case "edit_op_30d":
		if err := actionEditOp30d(ctx, d); err != nil {
			return err
		}
	case "delete_op_30d":
		if err := actionDeleteOp30d(ctx, d); err != nil {
			return err
		}
	case "export_ops_json":
		if err := actionExportOpsJSON(ctx, d); err != nil {
			return err
		}
	case "summary_cat_30d":
		if err := actionSummaryCat30d(ctx, d); err != nil {
			return err
		}
	case "summary_cat_period":
		if err := actionSummaryCatPeriod(ctx, d); err != nil {
			return err
		}
	case "import_ops_json":
		if err := actionImportOpsJSON(ctx, d); err != nil {
			return err
		}
	case "rename_account":
		if err := actionRenameAccount(ctx, d); err != nil {
			return err
		}
	case "rename_category":
		if err := actionRenameCategory(ctx, d); err != nil {
			return err
		}
	case "delete_category":
		if err := actionDeleteCategory(ctx, d); err != nil {
			return err
		}
	case "export_ops_yaml":
		if err := actionExportOpsYAML(ctx, d); err != nil {
			return err
		}
	case "import_ops_yaml":
		if err := actionImportOpsYAML(ctx, d); err != nil {
			return err
		}
	case "exit":
		return nil
	default:
		fmt.Println("Неизвестная команда:", key)
	}
	return nil
}
