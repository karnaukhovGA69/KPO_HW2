-- 1. Дедуп категорий: оставляем по одному id на (type, lower(btrim(name)))
WITH d AS (
  SELECT id, type, name,
         lower(btrim(name)) AS norm_name,
         row_number() OVER (PARTITION BY type, lower(btrim(name)) ORDER BY id) AS rn
  FROM categories
),
canon AS (
  SELECT type, norm_name, id
  FROM d WHERE rn = 1
)
-- Перевешиваем операции на канонический id
UPDATE operations o
SET category_id = c.id
FROM d dup
JOIN canon c ON c.type = dup.type AND c.norm_name = dup.norm_name
WHERE o.category_id = dup.id AND dup.rn > 1;

-- Удаляем лишние категории
DELETE FROM categories c
USING d
WHERE c.id = d.id AND d.rn > 1;

-- 2. Уникальность для категорий
ALTER TABLE categories
  ADD CONSTRAINT uq_categories_type_name UNIQUE (type, name);

