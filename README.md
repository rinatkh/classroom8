# Классная работа 8 — структуры данных в Go

Продолжаем занятие 7: Big O, срез, map, тесты и честные benchmark теперь применяем к способам хранения данных. Параметры типов из занятия 6 позволяют использовать те же структуры с int и указателями. Новых архитектурных слоёв и backend-приложения здесь нет. Go 1.22, стандартная библиотека, независимый модуль.

## Структура и запуск

| Раздел | Код | Команда |
|---|---|---|
| Singly/Doubly Linked List | internal/linkedlist | make run-01_linkedlist |
| Stack: slice и list | internal/stack | make run-02_stack |
| Queue: slice shift и list | internal/queue | make run-03_queue |
| Linear Buffer и bytes.Buffer | internal/buffer | make run-04_buffer |
| Circular Buffer, reject-on-full | internal/ringbuffer | make run-05_ringbuffer |
| Hash Table с цепочками | internal/hashtable | make run-06_hashtable |
| JSON состояний реальных структур | internal/demo | make run-07_visualizations |
| Измерения | internal/benchmark, internal/dataset | make bench-all |

```bash
make run-all
make ci
make bench-list
make bench-stack
make bench-queue
make bench-hash
make bench-buffer
make bench-memory
make bench-all BENCHTIME=100ms BENCHCOUNT=3 BENCHCPU=1
```

`cmd` печатает учебный результат; это полезный вывод демонстрации. Внутри benchmark нет Println. Нулевая List/Stack/Queue/Linear/Table готова к работе; Ring создаётся через New с положительной capacity. Эти структуры не синхронизированы: для общего доступа горутин нужны механизмы занятия 6. Обычные тесты и race запускаются отдельно от измерений.

## Как читать код

Сначала контракт и инварианты в комментариях, затем изменение связей/индексов, после — тесты и сложность. Удаление очищает ненужные ссылки. List/Doubly/Table и заполненные контейнеры не копируем: структура содержит указатели или общий backing array. Handle узла отличается от индекса: найденная связь меняется O(1), её поиск может стоить O(n).

## Материалы

- [Операции, сложность и память](docs/structures.md).
- [Сценарий на 90 минут и live coding](docs/live-coding-script.md).
- [Локальная интерактивная визуализация](docs/visualizations.html): открывается как файл, без сети и сервера. Выбор сцены, шаги и autoplay. Данные встроены из [Go JSON](docs/trace-data.json).
- [Короткие видео и визуализации структур данных](docs/short-videos.md), [локальные сцены и паузы](docs/visualizations.md).
- [Линейный Buffer: чтение, рост и Reset](docs/buffer-visualization.html), [данные реального Go-кода](docs/buffer-trace-data.json).
- [Методика benchmark](docs/benchmark-guide.md) и [реальные результаты](docs/benchmarks/README.md).
- [Матрица тестов и проверяемые инварианты](docs/test-matrix.md).
- [Преемственность курса](docs/continuity.md). Корневой `lesson8_additional_reading.md` остаётся в комплекте занятия; ссылки этого README работают и в отдельном репозитории.
- [GitHub Actions: запуск и диагностика CI](docs/ci.md).

`make coverage` измеряет statements, не доказывает корректность. CI включает fmt, vet, unit/integration, race, coverage-check (90%) и сборку семи примеров плюс benchreport. Артефакты bin/ и coverage.out не входят в исходники.

## Понятный результат benchmark

`make bench-list` запускает настоящие `testing.B` и печатает отдельные таблицы для каждого `n`. По умолчанию в таблице **медиана трёх прогонов**, min/max времени, `B/op`, `allocs/op` и смысл одной операции. Среднее арифметическое отдельно находится в CSV. `b.N` — число вызовов измеряемого тела, а не длина данных и не число повторов `-count`.

```bash
make help
make bench-list BENCHSIZE=1000 BENCHOP=search-missing
make bench-list BENCHVIEW=full
make bench-hash BENCHSIZE=10000 BENCHOP=lookup-last
make bench-buffer BENCHSIZE=1000 BENCHOP=accumulate
make bench-buffer BENCHSIZE=32 BENCHOP=ring-cyclic-64-pairs
make bench-list-raw BENCHTIME=20ms BENCHCOUNT=3
make bench-smoke
```

`BENCHSIZE` выбирает имеющийся размер `n`, число `chunks` или `cap`; не создаёт новый набор данных. Все 112 сценариев остаются в raw/CSV, даже если компактная таблица скрывает дополнительные операции. Для одного запуска в `bin/benchmarks/<group>-…/` сохраняются `raw.txt`, `summary.csv`, `report.md`, `environment.json`. Ошибка запуска или неполный raw приводят к ненулевому exit code. `1x` проверяет ветки, а не быстродействие.

Одна op Stack — вся серия `n Push + n Pop`, Queue — 64 пары, Buffer accumulate — сбор всех фрагментов. `B/op` означает выделения, не весь размер структуры. Смотрите [границы измерения](docs/benchmark-guide.md), прежде чем сравнивать значения. Увеличивайте `BENCHTIME` и `BENCHCOUNT` для серьёзных выводов.
