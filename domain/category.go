package domain

import (
	"errors"
	"strings"
)

var (
	ErrEmptyCategoryID     = errors.New("category id is empty")
	ErrEmptyCategoryName   = errors.New("category name is empty")
	ErrInvalidCategoryType = errors.New("invalid category type")
)

type Category struct {
	ID   CategoryID   `json:"id"   yaml:"id"`
	Type CategoryType `json:"type" yaml:"type"` // 1=income(доход), -1=expense(траты)
	Name string       `json:"name" yaml:"name"`
}

func (c Category) Sign() int {
	if c.IsExpense() {
		return -1
	}
	return 1
}

func (c Category) Validate() error {
	if strings.TrimSpace(string(c.ID)) == "" {
		return ErrEmptyCategoryID
	}
	if strings.TrimSpace(c.Name) == "" {
		return ErrEmptyCategoryName
	}
	if c.Type != CatIncome && c.Type != CatExpense {
		return ErrInvalidCategoryType
	}
	return nil
}

func (c Category) IsIncome() bool  { return c.Type == CatIncome }
func (c Category) IsExpense() bool { return c.Type == CatExpense }

func (c *Category) Rename(name string) error {
	if c == nil {
		return errors.New("nil receiver: Category")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrEmptyCategoryName
	}
	c.Name = name
	return nil
}

func (c *Category) SetType(t CategoryType) error {
	if c == nil {
		return errors.New("nil receiver: Category")
	}
	if t != CatIncome && t != CatExpense {
		return ErrInvalidCategoryType
	}
	c.Type = t
	return nil
}
