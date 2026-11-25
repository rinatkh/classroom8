# Контрольный запуск после исправлений — 25.11.2026

Go 1.22.12, linux/amd64, AMD EPYC 9V74 80-Core Processor, `-cpu=1`, `-benchtime=20ms`, `-count=3`. 112 сценариев × 3 повтора = **336 измерений**, все шесть групп завершились PASS. Это короткий обзорный замер в общей среде, не выделенное ядро; для выводов увеличьте длительность и повторы. Параллельно с этими измерениями другие Go tests, race и benchmark не запускались.

[Все агрегаты](summary.csv), [версия Go, команды и окружение шести запусков](environment.json). Медиана, среднее арифметическое, min/max и allocation находятся в разных столбцах. 896 числовых агрегатов независимо пересчитаны из raw через Python statistics. Raw сохранён с удалением только пробелов в концах строк; метрики и имена не изменены.

| Сценарий | Медиана ns/op | Min–max ns/op | B/op | allocs/op |
|---|---:|---:|---:|---:|
| `BenchmarkList/search-missing/slice/n=1000` | 298.7 | 296.1–300.8 | 0 | 0 |
| `BenchmarkList/search-missing/linked/n=1000` | 1151 | 1142–1167 | 0 | 0 |
| `BenchmarkStack/push-pop/slice/n=256` | 1266 | 1243–1414 | 0 | 0 |
| `BenchmarkStack/push-pop/linked/n=256` | 11088 | 10381–11723 | 6144 | 256 |
| `BenchmarkQueue/64-pairs/ring/n=1000` | 330 | 325.8–331.2 | 0 | 0 |
| `BenchmarkHash/lookup-last/map/n=1000` | 8.892 | 8.721–9.091 | 0 | 0 |
| `BenchmarkBuffer/accumulate/bytes-buffer/chunks=1000` | 8653 | 8604–8936 | 32752 | 10 |
| `BenchmarkBuffer/ring-cyclic-64-pairs/cap=1024` | 265.6 | 257.2–272.6 | 0 | 0 |
| `BenchmarkMemory/singly-build/n=1000` | 34439 | 32633–37844 | 24024 | 1001 |

Для отсутствующего значения при n=1000 отношение медиан linked/slice = 3.85× на этой машине. Это один готовый поиск, а не цена построения. Stack — вся серия Push/Pop; Queue/Ring — 64 пары; Buffer accumulate — все 1000 фрагментов. Их значения нельзя сравнивать как время одного метода.

Hash lookup использует уже подготовленную структуру. Hash build у slice не строит индекс. Ring bounded не хранит весь накапливаемый поток. `B/op` — выделения в измеряемом теле; ранее выделенные массивы и узлы продолжают занимать память. Retained heap здесь не измерялся. Условия каждого сценария — в [методике](../../benchmark-guide.md).

Воспроизведение:

```bash
make bench-all BENCHTIME=20ms BENCHCOUNT=3 BENCHCPU=1 BENCHVIEW=full
# Более длительное измерение на своём компьютере:
make bench-all BENCHTIME=300ms BENCHCOUNT=7 BENCHCPU=1
```

Raw-логи: [list](list-raw.txt), [stack](stack-raw.txt), [queue](queue-raw.txt), [hash](hash-raw.txt), [buffer](buffer-raw.txt), [memory](memory-raw.txt). Для форматирования одного лога: `go run ./cmd/benchreport -kind list -input docs/benchmarks/2026-11-25/list-raw.txt -count 3 -view full`. Исторические результаты 6 ноября сохранены отдельно и с новыми не усреднялись.
