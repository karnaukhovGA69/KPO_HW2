# KPO_HW2 — консольный учёт финансов (Go)

Мини‑приложение для учёта доходов/расходов со следующими фичами:

- операции по счёту: доход/расход, редактирование, удаление;
- категории (доход/расход): создание, переименование, удаление;
- аналитика: сводка за 30 дней и за произвольный период, разбивка по категориям;
- импорт/экспорт операций в **CSV/JSON/YAML**;
- **DI** (uber/dig), фасады для фич‑сценариев, **замер времени сценариев**, **кэш категорий** (Proxy);
- PostgreSQL для хранения.

> Проект ориентирован на демонстрацию архитектурных практик (GoF, SOLID/GRASP) в живом приложении.

---

## Содержание

- [Быстрый старт](#быстрый-старт)
- [Технологии](#технологии)
- [Структура проекта](#структура-проекта)
- [Как это работает (пайплайн сценария)](#как-это-работает-пайплайн-сценария)
- [Доменная модель и инварианты](#доменная-модель-и-инварианты)
- [Паттерны и практики](#паттерны-и-практики)
- [Меню и порядок пунктов](#меню-и-порядок-пунктов)
- [Импорт/экспорт: форматы](#импортэкспорт-форматы)
- [Аналитика](#аналитика)
- [Замер времени сценариев](#замер-времени-сценариев)
- [Конфигурация и переменные окружения](#конфигурация-и-переменные-окружения)
- [Схема БД и DDL](#схема-бд-и-ddl)
- [Тестирование (минимальный план)](#тестирование-минимальный-план)
- [FAQ / Траблшутинг](#faq--траблшутинг)
- [Расширение (как добавить новый формат)](#расширение-как-добавить-новый-формат)
- [Roadmap](#roadmap)
- [Лицензия](#лицензия)

---

## Быстрый старт

Требования: **Go 1.21+**, **PostgreSQL 13+**.

1. Настрой БД (см. [Схема БД и DDL](#схема-бд-и-ddl)).
2. Укажи строку подключения, например:
   ```bash
   export DATABASE_URL='postgres://user:pass@localhost:5433/finance?sslmode=disable'
   ```
3. (Опционально) путь к меню:
   ```bash
   export MENU_PATH=menu/menu.json
   ```
4. Запуск:
   ```bash
   go run .
   ```
5. Выбери активный счёт и работай через меню.

---

## Технологии

- Язык: **Go**
- БД: **PostgreSQL**, драйвер/пул: **pgx/pgxpool**
- DI: **uber/dig**
- Денежная арифметика: **shopspring/decimal**
- Форматы импорта/экспорта: **encoding/csv**, **encoding/json**, **gopkg.in/yaml.v3**

---

## Структура проекта

```
.
├── di/
│   └── di_run.go                  # DI-композиция (dig), сборка App
├── db/
│   └── ...                        # подключение к PostgreSQL (pgxpool)
├── domain/
│   └── ...                        # чистые доменные типы/инварианты/Factory
├── repo/
│   ├── account_pg.go              # PgAccountRepo
│   ├── category_pg.go             # PgCategoryRepo
│   ├── operation_pg.go            # PgOperationRepo
│   └── cached_category_repo.go    # Proxy-кэш над категориями (in-memory)
├── facade/
│   ├── ports.go                   # интерфейсы (порты) для фасадов
│   ├── account_facade.go
│   ├── category_facade.go
│   ├── operation_facade.go
│   └── analytics_facade.go
├── files/
│   ├── importer.go                # Template Method: общий каркас импорта
│   ├── exporter.go                # Strategy: общий каркас экспорта
│   ├── csv_ops.go                 # CSV: Encoder/Importer
│   ├── json.ops.go                # JSON: Encoder/Importer
│   └── yaml_ops.go                # YAML: Encoder/Importer
├── menu/
│   ├── run.go                     # основной цикл меню (Command+Timing)
│   ├── execute.go                 # диспетчер кейсов
│   ├── action.go                  # обработчики действий (через фасады)
│   ├── command.go                 # Command + Decorator (тайминг)
│   └── menu.json                  # конфигурация пунктов меню
└── main.go
```

---

## Как это работает (пайплайн сценария)

```
menu.json → menu.Run → (пользователь выбирает пункт) → Execute(ctx,key,deps)
         → (обёртка Command + WithTiming) → вызов action*
         → facade.* → repo.* → db → результат в консоль
```

- **Меню** читается из `menu/menu.json`.
- **Каждый пункт** оборачивается в декоратор тайминга (`WithTiming`) — лог в консоль и `timings.log`.
- Обработчики в `menu/action.go` вызывают **фасады** (входная точка фичи).
- Фасады работают с **репозиториями** и **доменной моделью**. Категории идут через **кэш‑прокси**.

---

## Доменная модель и инварианты

- **BankAccount**: `ID`, `Name`, `Balance` (decimal).  
  Инвариант: баланс не может уйти в минус; методы `Credit/Debit` валидируют операции.
- **Category**: `ID`, `Name`, `Type` (`CatIncome`/`CatExpense`).
- **Operation**: `ID`, `Type` (`OpIncome`/`OpExpense`), `AccountID`, `Amount`, `Date`, `CategoryID`, `Description`.
- **Factory**: централизованное создание доменных объектов (валидации).

---

## Паттерны и практики

### GoF

- **Facade**
  - `facade.OperationFacade` — добавление/редактирование/удаление операций, синхронизация баланса, авто‑создание категорий по имени.
  - `facade.AccountFacade` — CRUD, пересчёт баланса.
  - `facade.CategoryFacade` — CRUD категорий с проверками.
  - `facade.AnalyticsFacade` — `Summary` и `BreakdownByCategory`.
- **Command + Decorator**
  - `menu.Command` + `WithTiming` — обёртка всех сценариев меню, лог в `timings.log`.
- **Template Method** (импорт)
  - `files.BaseImporter.Import()` — общий алгоритм: чтение файла → `parse` → `[]Row`.
  - `CSVImporter/JSONImporter/YAMLImporter.parse` — своя «начинка».
- **Strategy** (экспорт)
  - `files.ExportOperations(..., enc Encoder)` + `CSVEncoder/JSONEncoder/YAMLEncoder` — сериализация в разные форматы.
- **Proxy**
  - `repo.CachedCategoryRepo` — кэширует `List/Get` категорий с инвалидацией при изменениях.

### SOLID/GRASP

- **DIP**: фасады завязаны на интерфейсы (`facade.CategoryRepo`), не на конкретные PG‑типы.
- **Low Coupling**: меню знает только про **один** метод фасада на сценарий.
- **High Cohesion**: бизнес‑правила в `domain/` и `facade/`.
- **SRP**: импортеры/энкодеры/обработчики не смешивают ответственности.

### DI

- **uber/dig** в `di/di_run.go`: сборка `App`, провайдинг `*pgxpool.Pool`, фабрики, репы, фасады, меню.
- В фасады подставляется **кэш категорий**; в `Deps.CatRepo` оставлен `PgCategoryRepo` для совместимости.

---

## Меню и порядок пунктов

Рекомендованный «человечный» порядок (`menu/menu.json`):

```json
[
	{ "field": "Добавить доход", "key": "add_income" },
	{ "field": "Добавить расход", "key": "add_expense" },
	{ "field": "Редактировать операцию (30 дней)", "key": "edit_op_30d" },
	{ "field": "Удалить операцию (за 30 дней)", "key": "delete_op_30d" },
	{ "field": "Список операций за 30 дней", "key": "list_ops_30d" },

	{ "field": "Сводка за 30 дней", "key": "summary_30d" },
	{ "field": "Сводка по категориям (30 дней)", "key": "summary_cat_30d" },
	{ "field": "Сводка по категориям (период)", "key": "summary_cat_period" },

	{ "field": "Экспорт операций (CSV)", "key": "export_ops_csv" },
	{ "field": "Импорт операций (CSV)", "key": "import_ops_csv" },
	{ "field": "Экспорт операций (JSON)", "key": "export_ops_json" },
	{ "field": "Импорт операций (JSON)", "key": "import_ops_json" },
	{ "field": "Экспорт операций (YAML)", "key": "export_ops_yaml" },
	{ "field": "Импорт операций (YAML)", "key": "import_ops_yaml" },

	{ "field": "Создать категорию", "key": "add_category" },
	{ "field": "Список категорий", "key": "list_categories" },
	{ "field": "Переименовать категорию", "key": "rename_category" },
	{ "field": "Удалить категорию", "key": "delete_category" },

	{ "field": "Список счетов", "key": "list_accounts" },
	{ "field": "Переименовать активный счёт", "key": "rename_account" },
	{ "field": "Создать новый счёт", "key": "create_account" },

	{ "field": "Выход", "key": "exit" }
]
```

---

## Импорт/экспорт: форматы

### CSV

Заголовок обязателен:

```
type,amount,date,category,description
```

Строки:

```
1,123.45,2025-01-15,Зарплата,Премия
-1,50.00,2025-01-16,Еда,Обед
```

- `type`: `1` — доход, `-1` — расход
- `amount`: десятичное число
- `date`: `YYYY-MM-DD`
- `category`: имя категории
- `description`: строка (опционально)

### JSON

```json
[
	{
		"type": 1,
		"amount": "123.45",
		"date": "2025-01-15",
		"category": "Зарплата",
		"description": "Премия"
	},
	{
		"type": -1,
		"amount": "50.00",
		"date": "2025-01-16",
		"category": "Еда",
		"description": "Обед"
	}
]
```

### YAML

```yaml
- type: 1
  amount: '123.45'
  date: '2025-01-15'
  category: 'Зарплата'
  description: 'Премия'
- type: -1
  amount: '50.00'
  date: '2025-01-16'
  category: 'Еда'
  description: 'Обед'
```

**Особенность доменной модели:** баланс счёта не может уйти в минус.  
Если счёт пустой и первая строка импорта — расход, строка может не пройти с ошибкой `insufficient funds`.

---

## Аналитика

`facade.AnalyticsFacade` предоставляет:

- **Summary** — `Income`, `Expense`, `Net` за период;
- **BreakdownByCategory** — суммы по категориям отдельно для доходов и расходов.  
  В меню они отображаются в объединённом виде (доход/расход/итого на категорию).

---

## Замер времени сценариев

Каждый запуск пункта меню проходит через `WithTiming(Command)`:

- В консоли:  
  `⏱ import_ops_csv (OK): 220ms` или `⏱ import_ops_csv (ERR): 4.962s`
- В файл `timings.log` (одна строка на сценарий):  
  `2025-11-05T12:34:56Z;import_ops_csv;OK;220ms`

Это покрывает требование «статистика времени сценариев».

---

## Конфигурация и переменные окружения

- `DATABASE_URL` — строка подключения к PostgreSQL (используется в `db.Connect`).
  Пример:
  ```bash
  export DATABASE_URL='postgres://user:pass@localhost:5433/finance?sslmode=disable'
  ```
- `MENU_PATH` — путь к `menu.json` (по умолчанию `menu/menu.json`).

---

## Схема БД и DDL

Минимально необходимая схема (UUID можно заменить на TEXT, если так реализовано в репозитории):

```sql
-- accounts
CREATE TABLE IF NOT EXISTS accounts (
  id        UUID PRIMARY KEY,
  name      TEXT NOT NULL,
  balance   NUMERIC(18,2) NOT NULL DEFAULT 0
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_accounts_name ON accounts(name);

-- categories
CREATE TABLE IF NOT EXISTS categories (
  id        UUID PRIMARY KEY,
  name      TEXT NOT NULL,
  type      INT  NOT NULL  -- 1 = CatIncome, -1 = CatExpense (или иное кодирование, как в domain)
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_categories_name ON categories(name);
CREATE INDEX IF NOT EXISTS ix_categories_type ON categories(type);

-- operations
CREATE TABLE IF NOT EXISTS operations (
  id           UUID PRIMARY KEY,
  account_id   UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
  type         INT  NOT NULL,     -- 1 = OpIncome, -1 = OpExpense
  amount       NUMERIC(18,2) NOT NULL CHECK (amount >= 0),
  date         DATE NOT NULL,
  category_id  UUID NOT NULL REFERENCES categories(id),
  description  TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS ix_operations_account_date ON operations(account_id, date);
CREATE INDEX IF NOT EXISTS ix_operations_category      ON operations(category_id);
CREATE INDEX IF NOT EXISTS ix_operations_type          ON operations(type);
```

> ⚠️ Убедись, что кодирование `type`/`category.type` соответствует твоему `domain` (например, `1/-1`).  
> Баланс счёта обновляется фасадом при создании/редактировании/удалении операций.

---

## Тестирование (минимальный план)

### files (импорт/экспорт)

- **Import**: дать фикстуры `ops.csv/json/yaml` → `ImportOperations*` → сравнить с ожидаемым `[]Row`.
- **Export**: использовать фейковые репозитории, вызвать `ExportOperations*`, распарсить результат обратно и сверить.

Скетч (псевдо‑код):

```go
func TestImportCSV(t *testing.T) {
  rows, err := files.ImportOperationsCSV("testdata/ops.csv")
  require.NoError(t, err)
  require.Len(t, rows, 3)
  require.Equal(t, 1, rows[0].Type) // ...
}
```

### facade.AnalyticsFacade

- Сгенерировать набор операций в фейковом репозитории и проверить `Income/Expense/Net` + агрегации по категориям.

### facade.OperationFacade

- **AddIncome/AddExpense**: проверка корректировки баланса и создания категорий по имени.
- **Edit**: изменение суммы корректирует баланс (в плюс/минус) в зависимости от типа операции.
- **Delete**: после добавления метода в repo — удаление операции и откат баланса.

---

## FAQ / Траблшутинг

**Импорт падает с `insufficient funds`.**  
Счёт пуст, а первая строка — расход. Решения: пополнить счёт, переставить доходы раньше расходов (в пределах даты), либо импортировать частями.

---

---

## Сообщения об ошибках и что они означают

Ниже — самые частые сообщения и как их решать.

- `insufficient funds`  
  Баланс счёта меньше суммы расхода (или вы уменьшили доход/увеличили расход при редактировании так, что баланс ушёл бы в минус).  
  **Что делать:** пополните счёт, импортируйте строки так, чтобы доходы шли раньше расходов в ту же дату, либо импортируйте частями.

- `category is required`  
  При добавлении/импорте не указано имя категории.  
  **Что делать:** заполните поле `category` (CSV/JSON/YAML) или введите имя при добавлении операции.

- `unknown operation type` / неправильный `type` в файле  
  В импорте допустимы только `1` (доход) и `-1` (расход).  
  **Что делать:** проверьте колонку/поле `type` в файле импорта.

- `operation service not wired: cannot delete`  
  В DI не подставлен `OperationService` для удаления операций.  
  **Что делать:** убедитесь, что в `di/di_run.go` в `Invoke` передаётся `opSvc *service.OperationService` и прокидывается в `facade.OperationFacade{ OpSvc: opSvc }`.

- `Delete not implemented ...`  
  Старая версия фасада без реализации удаления.  
  **Что делать:** обновите `facade/operation_facade.go` на актуальную реализацию (удаление через `OperationService.RemoveOperation`).

- Ошибки импорта файлов  
  - `open <path>: no such file or directory` — неверный путь к файлу. Укажите корректный путь.  
  - `csv: wrong number of fields` / «нет заголовков» — в CSV обязателен заголовок `type,amount,date,category,description`.  
  - `invalid date` / `parsing time ...` — формат даты должен быть `YYYY-MM-DD`.  
  - `invalid amount` — сумма должна быть десятичным числом, разделитель — точка (например, `123.45`).  
  - `JSON parse error` / `YAML parse error` — проверьте валидность синтаксиса и соответствие схемам из README.

- Ошибки БД (PostgreSQL)  
  - `duplicate key value violates unique constraint` — уже существует запись с таким именем (например, категория/счёт). Измените имя.  
  - `foreign key violation` — ссылка на несуществующий счёт/категорию. Проверьте порядок действий/данных.  
  - `permission denied` — нет прав на запись/чтение файла экспорта/лога. Запустите с правами или поменяйте путь.

- Ошибки меню/конфигурации  
  - «Пункт меню не найден» / «key not found»: ключ в `menu.json` не совпадает с `case` в `menu/execute.go`. Синхронизируйте ключи.  
  - `timings.log: permission denied`: нет прав на запись лога. Укажите доступный каталог или запустите с правами.

- Сообщения, которые не являются ошибкой
  - «Операций нет за период» — данных нет, это не ошибка.  
  - «Нет активного счёта» — создайте или выберите счёт в меню.

> Если получаете непонятное сообщение — запишите точный текст и шаги воспроизведения. Это поможет быстро локализовать проблему.
