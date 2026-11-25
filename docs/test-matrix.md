# Матрица тестов Classroom 8

Каждая строка ниже проверяет ключевые операции не менее чем на 10 именованных входах или последовательностях. Один сценарий может проверять несколько связанных методов: например, Enqueue, Front, Dequeue и Len после каждого шага. Это не означает, что каждый getter имеет отдельные 10 функций Test.

Общие 15 входов: nil, empty, zero, positive, negative, two_ascending, two_reverse, equal_pair, duplicates, negatives, sorted, reverse, mixed, extremes, many. Проверки строят независимый oracle на срезе, а не повторяют реализацию контейнера.

| Пакет и ключевые функции | Именованные семейства | Что доказывают проверки |
|---|---|---|
| linkedlist: Append, Prepend, Find, At, InsertAfter, DeleteAfter, RemoveFirst, PopFront, Reverse, Values; Head/Tail/Len | TestListContracts: 15 входов на каждую основную операцию | Порядок, первый duplicate, неверные индексы, узлы/owner, длина, tail.Next, независимость Values, пустой/одиночный список |
| linkedlist: InsertAfter/DeleteAfter с неверным handle | TestHandleRejection: 11 случаев | nil, чужой, удалённый и несвязанный узел не повреждают список |
| linkedlist: сырые Reverse и HasCycle | Reverse входит в 15 общих случаев; TestHasCycle: 13 случаев | Сохранение адресов/значений, разворот связей; nil, self-cycle, разные входы в цикл, повторные Value без цикла |
| linkedlist: ReverseTrace | TestReverseTrace: 15 входов | n+1 состояний того же цикла Reverse, связи обработанного префикса, current/previous, неизменность значений |
| linkedlist: Doubly Append/Prepend/Find/Remove, Values и metadata | TestDoublyContracts: 15 входов на каждую основную операцию | Прямой/обратный обход, Prev/Next, owner, повторное удаление, очистка отсоединённого узла |
| stack: Slice/Linked Push, Pop, Peek, Len; IsEmpty | TestContracts: 2 реализации × 15 входов; TestIsEmpty: 15 | LIFO, пустой результат с bool, ноль и отрицательные, повторное использование |
| stack: ValidParentheses | TestValidParentheses: 14 случаев | Вложенность, перекрёстные/неполные скобки, неизвестный символ, пустая строка |
| stack: смешанные операции и очистка ссылок | 12 seed × 1000 шагов; 12 длин | Oracle после каждого шага; Pop очищает неиспользуемый слот сразу, capacity может сохраняться |
| queue: Slice/Linked Enqueue, Dequeue, Front, Len | TestContracts: 2 реализации × 15 входов | FIFO, пустое состояние, повторения, отрицательные, drain и reuse |
| queue: смешанные операции и очистка ссылок | 12 seed × 1000 шагов; 12 длин | Oracle после каждого шага; сдвиг и очистка последнего использованного слота |
| ringbuffer: New | TestCapacity: 12 значений | Отрицательная/нулевая ёмкость, capacity=1, малые и большие ёмкости |
| ringbuffer: Enqueue, Dequeue, Front, Snapshot, Len/Cap/IsEmpty/IsFull | TestRingModel: 12 capacity × 1000 шагов | FIFO, full-reject без изменения данных, empty, многократный wrap, tail=(head+len)%cap, независимость Snapshot |
| ringbuffer: очистка ссылок | TestReferenceCleanup: 12 capacity, по 5 циклов | Каждый освобождённый слот сразу nil, после drain все слоты пусты; zero-value отдельно |
| buffer: Write, Read, Bytes, Reset, Len/Cap | TestLinearContract: 14 строк; TestWriteReadCycles: 12 размеров × 500 циклов | Reader/Writer против bytes.Buffer, nil/empty, UTF-8 как байты, NUL, частичное чтение/EOF, compaction, копии входа/снимка, reset |
| hashtable: HashString | TestHashString: 15 ключей | Независимый oracle hash/fnv, включая пустой, Unicode, NUL и длинный ключ |
| hashtable: Set/Get/Delete/Len/Snapshot | TestTableOperations: 15 ключей × 2 hasher | Принудительные коллизии, update без дубликата, нулевые значения, отсутствие, цепочки, число/расположение записей |
| hashtable: resize, New, Capacity/LoadFactor | TestResize: 13 размеров; TestZeroAndCapacity: 12 | Рост при load>0.75, сохранение всех пар, корректный zero-value, минимальная capacity |
| hashtable: Delete освобождает запись | TestDeleteUnlinksEntry: 15 ключей | Unlink головы/середины/хвоста, очистка key/value/next, повторное удаление |
| dataset: GenerateRandomSlice/Keys | TestDatasets: 12 размеров | Воспроизводимость seed, длина, независимость результатов, уникальность ключей |

Координирующие demo-функции проверяются отдельно по ожидаемому выводу. Integration запускает все 7 cmd, включая JSON визуализации. В сумме фактический `go test -count=1 -json ./...` выполнил **775 конечных именованных сценариев** (858 с родительскими Test/subtest). Покрытие внутренних пакетов — **99.5% statements**. Покрытие не является доказательством корректности и не заменяет инварианты.

Команды: `make test`, `make test-race`, `make coverage-check`, `make ci`. Они не требуют внешних сервисов. Стартовые красные тесты Homework описаны в его README (поставляется отдельным репозиторием).

После исправлений добавлены проверки benchreport (112 схем имён, разделение операций/размеров/CPU, медиана/среднее/min/max, единицы, неполные и ошибочные логи), CLI benchreport (запуск/ошибка процесса, файлы результата, pure raw stdout), фильтра размерности и дополнительного BufferTrace. Исходные тесты структур не удалялись и не ослаблялись.
