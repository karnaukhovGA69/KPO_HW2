package files

import (
	"os"
)

type Importer interface {
	parse(data []byte) ([]Row, error)
}

type BaseImporter struct {
	parser Importer
}

func (b BaseImporter) Import(path string) ([]Row, error) {
	bin, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	rows, err := b.parser.parse(bin)
	if err != nil {
		return nil, err
	}
	return rows, nil
}
