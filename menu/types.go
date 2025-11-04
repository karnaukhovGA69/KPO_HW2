package menu

type Item struct {
	Key   string `json:"key"`   // строковый ключ действия
	Field string `json:"field"` // текст для вывода
}

type Menu struct {
	Items []Item
}

// KeyAt — вернуть ключ по номеру пункта (1..N), "" если мимо.
func (m Menu) KeyAt(index int) string {
	if index < 1 || index > len(m.Items) {
		return ""
	}
	return m.Items[index-1].Key
}
