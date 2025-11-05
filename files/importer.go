package files

import (
	"os"
)

// Template Method pattern for importers.
// Общая структура и шаблон процесса импорта.
type Importer interface {
	parse(data []byte) ([]Row, error)
}

// BaseImporter — каркасный тип, реализующий шаблон метода Import.
type BaseImporter struct {
	parser Importer
}

// Import — общий метод для всех импортёров (CSV, JSON, YAML).
// Шаги: чтение файла → вызов конкретного парсера → возврат []Row.
func (b BaseImporter) Import(path string) ([]Row, error) {
	bin, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	rows, err := b.parser.parse(bin)
	if err != nil {
		return nil, err
	}
	// можно добавить общую пост-валидацию
	return rows, nil
}
