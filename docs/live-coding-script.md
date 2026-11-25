# Live coding — от связей к операциям

Сценарий рассчитан на 90 минут. Подборка с проверенными ссылками — [короткие визуализации](short-videos.md); подробные паузы локальных сцен — [visualizations.md](visualizations.md). Код classroom готов; преподаватель набирает только выделенные фрагменты в существующих файлах, затем возвращает рабочий вариант. Никаких временных проектов.

## 0–18: задача, память и выбор структуры

0–5: напомнить из занятия 7 поиск по значению в slice и map. Вопрос: «Если данные одинаковые, зачем менять способ хранения?» Предсказать цену вставки в начало.

5–12: сравнить непрерывные слоты slice и связи Node/Next, затем singly/doubly. Открыть локальную сцену list, сделать один InsertAfter. Обсудить известный handle против необходимости Find/At; физический адрес отличается от индекса.

12–18: кратко LIFO/FIFO, head/tail и память узлов. Не набирать все структуры с нуля: код и tests уже готовы. На уроке меняем только несколько строк Reverse, очистку ссылок и индексы Ring. Остальные реализации читаем и запускаем.

## 18–30: список

1. `make run-01_linkedlist`: ожидаются insert=[1 2 9 3], reverse=[3 9 2], doubly=[1 3].
2. Открыть internal/linkedlist/singly.go: Node, head/tail/size, Append. Предсказать оба случая пустой/непустой.
3. В InsertAfter показать порядок new.Next=at.Next, at.Next=new. Handle уже найден: изменили две связи O(1). Поиск через Find/At обсуждается отдельно.
4. В Reverse набрать сохранение next и три переназначения. После каждого шага остановить визуализацию reverse.
5. `go test ./internal/linkedlist -run 'TestListContracts/(two_reverse|duplicates|extremes)|TestReverseTrace' -v`.
6. HasCycle готовый: два nil-guard, сравнение адресов; набирать второй алгоритм с нуля не требуется.

## 30–42: Stack и Queue

Сначала одна короткая демонстрация №3 **или** №4, затем №6 из подборки; спросить, чем отличаются края удаления. `make run-02_stack run-03_queue`. Показать Pop со стиранием хвоста, затем сравнить Linked.Prepend/PopFront и Queue.Append/PopFront. Набрать только очистку слота и объяснить отличие `len--` от освобождения capacity. `go test ./internal/stack ./internal/queue -run 'TestClearsReferences|TestMixedSequences' -v`.

## 42–54: Buffer и Ring

Открыть [Buffer](buffer-visualization.html): 3–4 шага до Read/Reset, остановиться на сохранённой capacity. Затем `make run-04_buffer`: read="Go ", unread="API", изменённая копия "XPI" не меняет original="API". Стандартный bytes.Buffer.Bytes имеет другой alias-контракт.

Показать локальную сцену ring либо GIF №8 (18 с; его full/empty отличается от нашей модели size). Открыть Ring.Enqueue/Dequeue: набрать check full, запись, modulo, size; показать где очищается слот. На схеме ring capacity=3: full (0,0,3), после dequeue/enqueue (1,1,3), после drain (1,1,0). `make run-05_ringbuffer`; `go test ./internal/ringbuffer -run 'TestRingModel/cap=3$|TestReferenceCleanup/cap=3$' -v`.

## 54–64: Hash Table

MP4 №11: 12 с — chaining, 23 с — probing; либо интерактив №9. Показываем один вариант, а не всю подборку. `make run-06_hashtable`. До четвёртого ключа capacity=4, size=3, load=0.75; после — 8 и size=4. Set("a",9) обновляет запись без нового size. Показать готовые Get/Set/resize; не набирать все методы в десять минут. `go test ./internal/hashtable -run 'TestTableOperations|TestResize/n=100$' -v`.

## 64–77: короткий benchmark

```bash
make bench-list BENCHSIZE=1000 BENCHOP=search-missing BENCHTIME=20ms BENCHCOUNT=3
make bench-list BENCHSIZE=1000 BENCHOP=middle-known-undo BENCHTIME=20ms BENCHCOUNT=3
make bench-list BENCHSIZE=1000 BENCHOP=middle-index-undo BENCHTIME=20ms BENCHCOUNT=3
```

Перед измерением закрыть autoplay/видео. Показать отдельные таблицы по `n`, медиану, min/max и allocation. Одна op undo — обе операции, а не одна вставка. Сравнить известный handle и поиск по индексу. Итог: наблюдаемое ускорение зависит от контракта; benchmark не доказывает Big O.

Остальные группы подготовлены в [benchmarks/](benchmarks/README.md). `make bench-list-raw` показывает привычный testing.B; `BENCHVIEW=full` раскрывает скрытые компактным режимом операции. Разобрать, почему `b.N`, `n` и `-count` — разные величины. `B/op` не означает весь retained heap; ring с 0 B/op использует массив, выделенный заранее.

## 77–87: самостоятельная практика

Выбрать одну задачу, не все 19: Reverse, Valid Parentheses или wrap-around CircularQueue. Запустить в homework `make starter-check`, затем соответствующий `make run-linkedlist` / `make run-stack` / `make run-ringbuffer`: actual рядом с expected. Затем `make test-<section>` — настоящий acceptance. Спросить, какой инвариант должен сохраниться после каждой операции.

Дополнительно дома: Floyd cycle, merge sorted lists, MinStack, очередь через два стека, собственная HashTable, ByteBuffer, анализ и benchmark. Внешние визуализации необязательны; все упражнения компилируются без сети и аккаунтов.

## 87–90: вывод и CI

Спросить: «Что меняется при известном handle? Зачем Ring хранит size? Чем B/op отличается от capacity? Почему Hash не гарантирует O(1) в худшем случае?»

Показать `make help` и workflow обоих репозиториев. Classroom: `make ci` выполняем до урока, а не параллельно с benchmark. Homework: starter зелёный, functional matrix до реализации красная; `make test-benchmark` проверяет отдельную 19-ю задачу. Подробности — [CI](ci.md). Итог: структура выбирается по операциям, ограничениям памяти и контракту задачи; универсально лучшей структуры нет.
