# Реальные benchmark занятия 8

[Контрольный замер после исправлений 25.11.2026](2026-11-25/README.md): raw всех шести групп, 112 строк CSV, метаданные и независимая сверка агрегатов. Ниже сохранён исторический запуск 6 ноября, без смешивания результатов.

[Контрольный замер после исправлений 25.11.2026](2026-11-25/README.md): raw всех шести групп, 112 строк CSV, метаданные и независимая сверка агрегатов. Ниже сохранён исторический запуск 6 ноября, без смешивания результатов.

[Контрольный замер после исправлений 25.11.2026](2026-11-25/README.md): raw всех шести групп, 112 строк CSV, метаданные и независимая сверка агрегатов. Ниже сохранён исторический запуск 6 ноября, без смешивания результатов.

6 ноября 2026, Go 1.22.12 linux/amd64, **AMD EPYC 9V74 80-Core Processor**, CGO=1, -cpu=1. Полная матрица: **112 сценариев × 3 повтора = 336 измерений**, шесть групп PASS. Это короткий обзорный замер в shared container, без закреплённой частоты CPU и выделенного ядра.

```bash
make bench-all BENCHTIME=30ms BENCHCOUNT=3 BENCHCPU=1
# Для собственных выводов увеличьте длительность и повторы:
make bench-all BENCHTIME=300ms BENCHCOUNT=7 BENCHCPU=1
```

[Неизменённый сырой вывод](raw.txt), [окружение](environment.txt), [полный CSV](summary.csv), [методика](../benchmark-guide.md). Seed 7; runtime map может рандомизировать внутреннее размещение. Одинаковые входы воспроизводятся, точные ns/op воспроизводиться не обязаны.

## Список и срез

| Сценарий | ns/op  медиана | B/op | allocs/op |
|---|---:|---:|---:|
| `BenchmarkList/search-missing/slice/n=1000` | 324.80 | 0 | 0 |
| `BenchmarkList/search-missing/linked/n=1000` | 1 181.00 | 0 | 0 |
| `BenchmarkList/walk/slice/n=1000` | 369.20 | 0 | 0 |
| `BenchmarkList/walk/linked/n=1000` | 1 248.00 | 0 | 0 |
| `BenchmarkList/build-prepend/slice/n=1000` | 47 979.00 | 8192 | 1 |
| `BenchmarkList/build-prepend/linked/n=1000` | 31 998.00 | 24024 | 1001 |
| `BenchmarkList/middle-known-undo/slice/n=1000` | 88.72 | 0 | 0 |
| `BenchmarkList/middle-known-undo/linked/n=1000` | 45.03 | 24 | 1 |
| `BenchmarkList/middle-index-undo/slice/n=1000` | 89.76 | 0 | 0 |
| `BenchmarkList/middle-index-undo/linked/n=1000` | 613.20 | 24 | 1 |

Для отсутствующего значения n=1000 отношение linked/slice равно 3.64× на этой машине. При known handle связный список избегает сдвига; при index lookup сначала проходит связи. build-prepend — вся серия n вставок, middle-…-undo — вставка и обратное удаление, не одна вставка.

## Stack и Queue

| Сценарий | ns/op  медиана | B/op | allocs/op |
|---|---:|---:|---:|
| `BenchmarkStack/push-pop/slice/n=256` | 1 283.00 | 0 | 0 |
| `BenchmarkStack/push-pop/linked/n=256` | 10 999.00 | 6144 | 256 |
| `BenchmarkQueue/64-pairs/slice-shift/n=1000` | 5 486.00 | 0 | 0 |
| `BenchmarkQueue/64-pairs/linked/n=1000` | 3 234.00 | 1536 | 64 |
| `BenchmarkQueue/64-pairs/ring/n=1000` | 375.80 | 0 | 0 |

Stack op — 256 Push + 256 Pop. Queue op — 64 Enqueue/Dequeue при постоянной длине 1000. Ring заранее имеет capacity n+1 и не переполняется; это сравнение одной принятой FIFO-нагрузки, а не всех возможностей API. Срез стека прогрет до нужного cap до таймера.

## Hash Table

| Сценарий | ns/op  медиана | B/op | allocs/op |
|---|---:|---:|---:|
| `BenchmarkHash/build/custom/n=1000` | 79 008.00 | 85264 | 1011 |
| `BenchmarkHash/build/map/n=1000` | 89 143.00 | 122392 | 41 |
| `BenchmarkHash/build/slice/n=1000` | 6 679.00 | 24576 | 1 |
| `BenchmarkHash/lookup-last/custom/n=1000` | 13.27 | 0 | 0 |
| `BenchmarkHash/lookup-last/map/n=1000` | 9.56 | 0 | 0 |
| `BenchmarkHash/lookup-last/slice/n=1000` | 3 980.00 | 0 | 0 |
| `BenchmarkHash/insert-read-delete-triple/custom/n=1000` | 87.52 | 48 | 1 |
| `BenchmarkHash/insert-read-delete-triple/map/n=1000` | 34.74 | 0 | 0 |
| `BenchmarkHash/insert-read-delete-triple/slice/n=1000` | 3 416.00 | 0 | 0 |
| `BenchmarkHash/delete-restore-pair/custom/n=1000` | 79.82 | 48 | 1 |
| `BenchmarkHash/delete-restore-pair/map/n=1000` | 25.70 | 0 | 0 |
| `BenchmarkHash/delete-restore-pair/slice/n=1000` | 3 616.00 | 0 | 0 |

Lookup last — последний в исходном порядке среза, не самый дорогой bucket таблицы. Build создаёт все n пар: slice не строит индекс. Lookup исключает подготовку. Insert/read/delete — новый заведомо отсутствующий ключ, после тройки размер прежний. Delete/restore удаляет существующий ключ и возвращает пару; восстановление входит в таймер. Строки ключей имеют одинаковую ограниченную длину.

## Buffer и allocation

| Сценарий | ns/op  медиана | B/op | allocs/op |
|---|---:|---:|---:|
| `BenchmarkBuffer/accumulate/append/chunks=1000` | 9 050.00 | 62960 | 16 |
| `BenchmarkBuffer/accumulate/bytes-buffer/chunks=1000` | 7 796.00 | 32752 | 10 |
| `BenchmarkBuffer/accumulate/linear/chunks=1000` | 13 531.00 | 62992 | 17 |
| `BenchmarkBuffer/ring-cyclic-64-pairs/cap=1024` | 272.50 | 0 | 0 |
| `BenchmarkMemory/slice-build/n=1000` | 758.00 | 8192 | 1 |
| `BenchmarkMemory/singly-build/n=1000` | 34 575.00 | 24024 | 1001 |
| `BenchmarkMemory/std-doubly-build/n=1000` | 58 127.00 | 56048 | 2001 |

Accumulate собирает все фрагменты общего исходного набора в новом буфере. Ring-cyclic — отдельная модель ограниченного FIFO с 64 парами, она не сохраняет весь поток и не является эквивалентной заменой накопления. container/list — reference с двусвязными узлами и interface-значениями; это другой API и состав узла.

## Границы вывода

- ns/op относится ко всему именованному телу (серии/паре/тройке), не всегда к одному методу.
- B/op и allocs/op — выделения за время тела, не весь удерживаемый heap. Для prepared ring backing array создан до таймера. Retained heap отдельно не измерялся.
- Interface/closure, прогрев кэша, копирование и allocator влияют на константы. Сравнивайте одинаковые сценарии и размеры. Короткие наносекундные тела особенно шумны.
- Отношение медиан A/B означает, во сколько раз B быстрее A в данном сценарии на этой машине. Оно не доказывает Big O. Асимптотику выводим из операций и предпосылок.
- Структура выбирается по контракту: индексы, handle, порядок, bounded capacity, поток, ключи. В этих измерениях нет универсального победителя.
