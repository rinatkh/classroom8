# Структуры: контракт, время и память

n — число элементов, c — число слотов выделенного массива (у hash — buckets m), k — длина строкового ключа. У смещённого slice-view cap(view) может быть меньше удерживаемого backing array; O(c) относится к исходному выделению. В кратких O(1) строках hash/map длина ключей считается ограниченной и hash хорошо распределяет ключи. Полный худший случай сравнения n длинных строк может быть O(nk); хеширование одной строки O(k). Спецификация Go не закрепляет внутреннее устройство или худшую сложность map.

## Сравнение операций

| Структура | Поиск | Доступ | Вставка | Удаление | Суммарная память |
|---|---|---|---|---|---|
| Array / Slice | значение: best O(1), average/worst O(n) | известный индекс O(1) | середина O(n); append amortized O(1), worst O(n) | ordered O(n), unordered O(1) с известным индексом | Array O(n), Slice O(c) |
| Singly List с tail | best O(1), average/worst O(n) | At(index) O(n) | Append/Prepend/InsertAfter(handle) O(1); поиск handle отдельно | PopFront/DeleteAfter(handle) O(1); RemoveFirst(value) best O(1), worst O(n) | O(n), один Next и owner на узел |
| Doubly List | best O(1), average/worst O(n) | по индексу O(n) | у концов O(1) | Remove(живой handle) O(1), поиск по value O(n) | O(n), Prev/Next/owner на узел |
| Slice Stack | перебор O(n), в API не нужен | Peek O(1), произвольный доступ не часть LIFO | Push amortized O(1), worst O(n) | Pop O(1) | O(c), cap может остаться после drain |
| Linked Stack | O(n), в API не нужен | Peek O(1) | Push O(1), новый узел | Pop O(1) | O(n) |
| Slice Queue (shift) | O(n), в API не нужен | Front O(1) | Enqueue amortized O(1), worst O(n) | Dequeue O(n), best O(1) при длине ≤1 | O(c) |
| Linked Queue | O(n), в API не нужен | Front O(1) | Enqueue O(1), tail | Dequeue O(1), head | O(n) |
| Ring Queue | O(n), в API не нужен | Front O(1), произвольного индекса в API нет | Enqueue O(1), full отклоняется | Dequeue O(1) | O(c) фиксировано |
| Linear / bytes.Buffer | поиск байта O(n), не главный API | Read k байт O(k) | Write k: обычно amortized O(k), worst O(n+k) при copy/grow | прочитать k: O(k), Reset в нашей версии O(used) | O(c), растущая |
| Custom Hash Table | Get expected O(1), worst O(n) при ограниченном k | ключ, а не порядковый индекс | Set expected amortized O(1), worst O(n+m) при росте | Delete expected O(1), worst O(n) | O(n+m), плюс память самих ключей |
| Go map | ожидаемая модель lookup O(1); spec не обещает worst | по ключу; чтение map[k] без ok не отличает отсутствующий ноль | ожидаемо amortized O(1) | ожидаемо O(1) | зависит от размера/hint/runtime; измеряем allocation |

## Память отдельной операции

- List Append/Prepend/InsertAfter выделяют один узел O(1); Find/At/Reverse/удаление требуют O(1) дополнительных ссылок. Values/Snapshot выделяет O(n). Отсоединяемый узел очищается, включая owner. Иначе сохранённый handle мог бы удерживать всю цепочку.
- Slice Push/Enqueue без роста O(1) дополнительной памяти, рост backing array O(n). Pop/Dequeue очищает последний использовавшийся слот, но не уменьшает capacity. Уменьшение len не равно освобождению массива.
- Ring New выделяет O(c), Enqueue/Dequeue/Front не выделяют память и используют O(1) дополнительных переменных. Snapshot требует O(c), в benchmark операций не входит.
- Linear Write k копирует вход: возможно O(n+k) нового массива, при compaction переносится unread. Read очищает только прочитанные k байт; при исчерпании сбрасывает len/offset без повторной очистки всей истории. Reset вручную очищает used и сохраняет capacity. Bytes нашей версии — копия O(n); bytes.Buffer.Bytes — alias, дополнительная память O(1).
- Table Set без роста выделяет O(1) узел только для нового ключа. Update переиспользует запись. Resize создаёт новый массив O(m), узлы переиспользует и переназначает цепочки. Delete массив не уменьшает. Строковые ключи удерживают исходные данные; сумма их байтов отдельно от O(n+m) заголовков.
- Снимки ReverseTrace хранят O(n²) данных за n шагов; обычный Reverse сохраняет O(1) дополнительной памяти. Не измеряем алгоритм по стоимости trace.

## Инварианты и завершение

List: head==nil ↔ tail==nil ↔ len==0, tail.Next==nil, число достижимых узлов равно len. Доступные Next читаем, но меняем только методами List. Для отдельной сырой цепочки разрешены эксперименты с cycle; не превращаем обычный контейнер в циклический.

Doubly: у каждой пары a.Next==b одновременно b.Prev==a, у head.Prev и tail.Next nil. Живой handle проверяется через owner за O(1). После удаления тот же handle повторно не принимается.

Reverse: обработанный префикс развёрнут, current указывает на оставшуюся цепочку. Сначала сохранить next, затем изменить Next. На каждой итерации на один необработанный узел меньше. Reverse требует конечную ациклическую цепочку.

Cycle: fast делает два шага, slow один. В ациклическом случае fast достигает nil; в цикле относительное расстояние меняется и указатели встречаются. Сравниваем адреса узлов, не Value: повторение значений не означает cycle.

Ring: 0≤len≤c, head/tail в [0,c), tail==(head+len)%c. Head — следующий читаемый, tail — следующий записываемый слот. Full и empty могут иметь head==tail; различаем по len. После c−1 индекс становится 0. Политика reject не затирает старое значение. Last-k с overwrite — другой контракт в Homework.

Hash: каждая запись ровно в одном bucket hash%m, ключ уникален, size равен числу узлов. Set сначала ищет существующий ключ, потом решает о росте. При m→2m меняется остаток, поэтому переносим все записи. Cached hash делает resize O(n+m), без повторного чтения всех строк. Случайная/равномерная hash — предпосылка ожидаемого времени; фиксированная учебная FNV этого не гарантирует для произвольного противника.

## Применение и ошибки выбора

- Slice удобен для плотных индексов и обхода, cache locality часто лучше узлов.
- List полезен для частых изменений по уже сохранённому handle; поиск каждый раз может убрать это преимущество. Для удаления по handle в обе стороны — Doubly.
- Stack хранит вложенность: скобки, undo, стек обхода и вызовы рекурсии. Сам по себе стек данных не заменяет системный call stack.
- FIFO сохраняет порядок ожидания событий worker. Ring подходит при фиксированном лимите, но переполнение нужно обработать; этот код не даёт blocking или goroutine safety как channel.
- Buffer накапливает JSON/байты перед записью в sink. Кольцо — bounded поток с повторным использованием слотов, не место хранения всего payload.
- Map подходит для lookup по ID; собственная таблица здесь нужна для понимания hash, коллизий и amortized роста.

## Источники

[Go slices](https://go.dev/blog/slices-intro), [bytes Go 1.22](https://pkg.go.dev/bytes@go1.22.12), [container/list](https://pkg.go.dev/container/list), [Open Data Structures: Hash Tables](https://opendatastructures.org/newhtml/ods/latex/hashing.html), [Cornell: hashing](https://www.cs.cornell.edu/courses/cs2110/2025fa/lectures/lec20/). Реализации сверены с исходниками установленного Go 1.22.12; подробности runtime map не преподаются как стабильный контракт.

Снимки Values/Snapshot копируют элементы и массивы, но не выполняют deep copy объектов по указателям. Очистка ссылки делает объект доступным для GC, не обещает немедленного возврата памяти ОС.
