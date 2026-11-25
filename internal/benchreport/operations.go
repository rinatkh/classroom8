package benchreport

import (
	"fmt"
	"strconv"
)

type operationSpec struct {
	Dimension          string
	Implementations    []string
	Title, Preparation string
	Comparable         bool
}

// Каталог соответствует существующим b.Run. Пара/тройка/серия — вся op;
// восстановление, если оно названо, также остаётся в измеренном времени.
var specs = map[string]operationSpec{
	"BenchmarkList/search-missing":            {"n", []string{"slice", "linked"}, "Один поиск отсутствующего значения", "Обе структуры готовы; одинаковый target вне диапазона данных. Время O(n), дополнительная память поиска O(1).", true},
	"BenchmarkList/walk":                      {"n", []string{"slice", "linked"}, "Один полный обход с суммированием", "Подготовка исключена; проходит весь вход, O(n) времени и O(1) дополнительной памяти.", true},
	"BenchmarkList/build-append":              {"n", []string{"slice", "linked"}, "Построение: %d Append", "Создание и allocation включены. Slice выделяет cap=n внутри op; Linked выделяет узлы. Это весь build, O(n).", true},
	"BenchmarkList/build-prepend":             {"n", []string{"slice", "linked"}, "Построение: %d Prepend", "Создание включено. Slice заранее выделяет cap=n, но каждый раз сдвигает хвост: O(n²) за build; Linked — O(n).", true},
	"BenchmarkList/prepend-undo":              {"n", []string{"slice", "linked"}, "Prepend + обратное удаление (2 действия)", "Оба действия внутри op; размер восстанавливается. Slice имеет cap=n+1. Новые Linked-узлы выделяются в таймере.", true},
	"BenchmarkList/append-undo":               {"n", []string{"slice", "linked"}, "Append + обратное удаление (2 действия)", "Готовый tail-handle Linked и свободный слот Slice; поиск позиции исключён. Обе части внутри op, размер постоянен.", true},
	"BenchmarkList/middle-known-undo":         {"n", []string{"slice", "linked"}, "Вставка в середину + отмена (позиция известна)", "Linked получает готовый handle, Slice — готовую позицию. Получение позиции исключено; сдвиги Slice и allocation узла Linked включены.", true},
	"BenchmarkList/middle-index-undo":         {"n", []string{"slice", "linked"}, "Поиск узла по индексу + вставка и отмена", "Индекс известен обоим. Linked.At(n/2) внутри таймера добавляет O(n) проход; это не O(1) для всего вызова.", true},
	"BenchmarkList/delete-value-last-restore": {"n", []string{"slice", "linked"}, "Найти и удалить последнее значение + восстановить", "Уникальные значения; поиск по значению внутри op, восстановление тоже. Размер не растёт с b.N.", true},
	"BenchmarkStack/push-pop":                 {"n", []string{"slice", "linked"}, "%d Push + %d Pop — вся серия", "Slice прогрет до нужной capacity до таймера; Linked выделяет узлы в каждой серии. Интерфейсные вызовы и суммирование входят в op.", true},
	"BenchmarkQueue/64-pairs":                 {"n", []string{"slice-shift", "linked", "ring"}, "64 Enqueue + 64 Dequeue — вся серия FIFO", "Длина n постоянна. Ring имеет готовую cap=n+1, все записи приняты; Slice Shift сдвигает элементы, Linked выделяет узлы. Это не универсальное сравнение всех queue API.", true},
	"BenchmarkHash/build":                     {"n", []string{"custom", "map", "slice"}, "Создать структуру из %d пар string→int", "Создание и рост включены. Custom и map строят индекс, Slice только хранит пары: результаты имеют разную готовность к lookup. Общий победитель не выбирается.", false},
	"BenchmarkHash/lookup-last":               {"n", []string{"custom", "map", "slice"}, "Один поиск существующего ключа", "Структуры готовы. last — последняя позиция Slice, не худший bucket. Ключи одинаковой ограниченной длины; custom/map ожидаемо O(1), Slice O(n).", true},
	"BenchmarkHash/lookup-missing":            {"n", []string{"custom", "map", "slice"}, "Один поиск отсутствующего ключа", "Построение исключено. Custom/map ожидаемо O(1) при хорошем распределении; худшая цепочка O(n) для ограниченной длины ключа.", true},
	"BenchmarkHash/insert-read-delete-triple": {"n", []string{"custom", "map", "slice"}, "Set + Get + Delete нового ключа (3 действия)", "Новый ключ отсутствует до op; чтение и удаление внутри. Размер восстанавливается, запас capacity подготовлен. Slice ищет ключ перед удалением.", true},
	"BenchmarkHash/delete-restore-pair":       {"n", []string{"custom", "map", "slice"}, "Delete существующего ключа + восстановление (2 действия)", "Обе части в таймере; custom повторно выделяет entry. В Slice поиск ключа включён. Размер остаётся n.", true},
	"BenchmarkBuffer/accumulate":              {"chunks", []string{"append", "bytes-buffer", "linear"}, "Накопить %d фрагментов — весь payload", "Каждый фрагмент 16 B, новый объект на op. Подготовка фрагментов и проверка результата исключены; рост включён. Linear.Bytes возвращает копию вне таймера, bytes.Buffer.Bytes — alias.", true},
	"BenchmarkBuffer/ring-cyclic-64-pairs":    {"cap", []string{"ring"}, "64 записи + 64 чтения — циклический FIFO", "Ring[byte] создан до таймера; занято cap−1 слотов. Все записи принимаются. Это bounded поток, он не накапливает весь payload как Buffer.", false},
	"BenchmarkMemory/build":                   {"n", []string{"slice-build", "singly-build", "std-doubly-build"}, "Построить %d элементов — allocation всей структуры", "Создание включено. container/list — двусвязный API с interface-значениями, не эквивалент generic singly. Измеряются allocation/op, не retained heap.", false},
}

var roots = map[string]string{"list": "BenchmarkList", "stack": "BenchmarkStack", "queue": "BenchmarkQueue", "hash": "BenchmarkHash", "buffer": "BenchmarkBuffer", "memory": "BenchmarkMemory"}

func Pattern(kind string, size int, operation string) (string, error) {
	root, ok := roots[kind]
	if !ok || size < 0 {
		return "", fmt.Errorf("неизвестная группа или отрицательный size: %s/%d", kind, size)
	}
	pattern := "^" + root + "$"
	if operation != "" {
		if _, ok := specs[root+"/"+operation]; !ok {
			return "", fmt.Errorf("операция %q отсутствует в %s", operation, kind)
		}
		if kind != "memory" {
			pattern += "/^" + operation + "$"
		}
	}
	return pattern, nil
}

// Фильтр применяется в benchmark до подготовки/таймера. Размер относится
// к параметру самого имени: n, chunks или cap; эти измерения не смешиваются.
func SizeFilter(size int) string {
	if size == 0 {
		return ""
	}
	return strconv.Itoa(size)
}

func operationText(n Name) string {
	title := specs[n.Root+"/"+n.Operation].Title
	switch n.Operation {
	case "push-pop":
		return fmt.Sprintf(title, n.Size, n.Size)
	case "build", "build-append", "build-prepend", "accumulate":
		return fmt.Sprintf(title, n.Size)
	default:
		return title
	}
}
