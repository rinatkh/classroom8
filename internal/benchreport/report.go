package benchreport

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

func FormatTime(ns float64) string {
	units := []string{"ns", "µs", "ms", "s"}
	i := 0
	for ns >= 1000 && i < len(units)-1 {
		ns /= 1000
		i++
	}
	return number(ns) + " " + units[i]
}

// FormatBytes использует двоичные единицы. Это allocation на op, если значение
// взято из B/op; тот же formatter не превращает его в retained heap.
func FormatBytes(bytes float64) string {
	units := []string{"B", "KiB", "MiB", "GiB"}
	i := 0
	for bytes >= 1024 && i < len(units)-1 {
		bytes /= 1024
		i++
	}
	return number(bytes) + " " + units[i]
}

func number(value float64) string { return strconv.FormatFloat(value, 'g', 4, 64) }

// IntBytes выводится из архитектуры самого лога, а не машины, читающей файл.
// Неизвестная архитектура не получает выдуманной оценки размера данных.
func IntBytes(arch string) int {
	switch arch {
	case "386", "arm", "mips", "mipsle", "amd64p32":
		return 4
	case "amd64", "arm64", "ppc64", "ppc64le", "mips64", "mips64le", "s390x", "riscv64", "loong64", "sparc64", "wasm":
		return 8
	default:
		return 0
	}
}

func implementationLabel(s string) string {
	labels := map[string]string{"slice": "Slice", "linked": "Linked List", "slice-shift": "Slice Shift", "ring": "Circular Buffer", "custom": "Custom Hash Table", "map": "Go map", "append": "append([]byte)", "bytes-buffer": "bytes.Buffer", "linear": "Linear Buffer", "slice-build": "Slice", "singly-build": "Generic Singly", "std-doubly-build": "container/list (doubly)"}
	return labels[s]
}

type reportGroup struct {
	Environment
	Root, Operation, Dimension string
	Size, Procs                int
}

func groupOf(row Row) reportGroup {
	n := row.Name
	return reportGroup{row.Environment, n.Root, n.Operation, n.Dimension, n.Size, n.Procs}
}

func payload(n Name, arch string) string {
	width := IntBytes(arch)
	if n.Dimension == "cap" {
		return fmt.Sprintf("Готовые byte-слоты: %d B; len=%d, cap=%d. Это не накопленный payload.", n.Size, n.Size-1, n.Size)
	}
	if n.Dimension == "chunks" {
		return fmt.Sprintf("Полезные байты входа: %d × 16 B ≈ %s; заголовки фрагментов и запас capacity отдельно.", n.Size, FormatBytes(float64(n.Size)*16))
	}
	if width == 0 {
		return "Размер int архитектуры лога неизвестен; оценка памяти элементов не рассчитана."
	}
	if n.Root == "BenchmarkHash" {
		return fmt.Sprintf("Логические данные: %d пар, по 12 B ключа + %d B значения; без string headers, buckets, узлов и allocator overhead.", n.Size, width)
	}
	text := fmt.Sprintf("Значения int: ≈ %s (%d × %d B), без узлов, указателей, header, запаса cap и allocator overhead.", FormatBytes(float64(n.Size)*float64(width)), n.Size, width)
	if n.Root == "BenchmarkQueue" {
		text += fmt.Sprintf(" Ring выделил %d int-слотов до таймера.", n.Size+1)
	}
	return text
}

// Render подписывает ВСЮ operation. Числа не пересчитываются на один метод
// или элемент: это исказило бы пары, тройки и серии исходной лаборатории.
func Render(rows []Row, view string) (string, error) {
	if len(rows) == 0 || !oneOf(view, "compact", "full") {
		return "", fmt.Errorf("нет строк или неизвестный вид %q", view)
	}
	var out strings.Builder
	fmt.Fprintln(&out, "# Структуры данных — реальные benchmark-измерения")
	fmt.Fprintln(&out, "Таблица: МЕДИАНЫ прогонов. k — число прогонов (-count), не b.N; min/max относятся ко времени, средние — в CSV.")
	fmt.Fprintln(&out, "Первая цифра raw после имени — b.N: вызовы внутри одного прогона. n/chunks/cap — разные параметры входа.")
	fmt.Fprintln(&out, "Время / op относится ко ВСЕМ действиям в заголовке: на n или число методов его НЕ умножают.")
	fmt.Fprintln(&out, "Heap / op — выделенные B/op; allocs/op — число выделений. 0 B/op не означает нулевую память структуры.")
	fmt.Fprintln(&out, "Go 1.22 округляет allocation/op вниз. Готовые backing arrays, узлы и стек не становятся нулевыми от этого.")
	preferred := map[string]bool{}
	for _, row := range rows {
		n := row.Name
		if n.Root == "BenchmarkList" && oneOf(n.Operation, "search-missing", "walk", "build-append") || n.Root == "BenchmarkHash" && oneOf(n.Operation, "build", "lookup-last") {
			preferred[n.Root] = true
		}
	}
	list := append([]Row(nil), rows...)
	sort.Slice(list, func(i, j int) bool { return rowLess(list[i], list[j]) })
	groups := []reportGroup{}
	byGroup := map[reportGroup][]Row{}
	hidden := 0
	for _, row := range list {
		n := row.Name
		if view == "compact" && preferred[n.Root] && (n.Root == "BenchmarkList" && !oneOf(n.Operation, "search-missing", "walk", "build-append") || n.Root == "BenchmarkHash" && !oneOf(n.Operation, "build", "lookup-last")) {
			hidden++
			continue
		}
		g := groupOf(row)
		if _, ok := byGroup[g]; !ok {
			groups = append(groups, g)
		}
		byGroup[g] = append(byGroup[g], row)
	}
	if hidden > 0 {
		fmt.Fprintf(&out, "Компактный вид: скрыто %d агрегированных строк других операций. Все есть в summary.csv; BENCHVIEW=full показывает их в таблицах.\n", hidden)
	}
	for _, g := range groups {
		first := byGroup[g][0]
		spec := specs[g.Root+"/"+g.Operation]
		fmt.Fprintf(&out, "\n## %s · %s=%d · GOMAXPROCS=%d\n", strings.TrimPrefix(g.Root, "Benchmark"), g.Dimension, g.Size, g.Procs)
		fmt.Fprintf(&out, "1 op: %s.\n", operationText(first.Name))
		cpu := g.CPU
		if cpu == "" {
			cpu = "модель CPU не указана"
		}
		fmt.Fprintf(&out, "%s/%s · %s.\n%s\n", g.GOOS, g.GOARCH, cpu, spec.Preparation)
		fmt.Fprintln(&out, payload(first.Name, g.GOARCH))
		cells := [][]string{{"Реализация", "Медиана / op", "Min / op", "Max / op", "Heap / op", "allocs/op", "k"}}
		for _, row := range byGroup[g] {
			cells = append(cells, []string{implementationLabel(row.Name.Implementation), FormatTime(row.MedianNS), FormatTime(row.MinNS), FormatTime(row.MaxNS), FormatBytes(row.MedianBytes), number(row.MedianAllocs), strconv.Itoa(row.Repeats)})
		}
		out.WriteString(table(cells))
		if spec.Comparable && len(byGroup[g]) > 1 {
			best, worst := first, first
			for _, r := range byGroup[g][1:] {
				if r.MedianNS < best.MedianNS {
					best = r
				}
				if r.MedianNS > worst.MedianNS {
					worst = r
				}
			}
			if best.MedianNS > 0 {
				fmt.Fprintf(&out, "В этой группе min медианы: %s; max/min = %s×. Статистическая значимость не оценивалась.\n", implementationLabel(best.Name.Implementation), number(worst.MedianNS/best.MedianNS))
			}
		}
	}
	fmt.Fprintln(&out, "\nB/op — allocation, не retained/peak heap. Полезные данные, дополнительные структуры и выделения/op не суммируются в псевдоточный total.")
	fmt.Fprintln(&out, "Lookup готовой таблицы не включает build. Ring bounded; accumulation хранит весь payload. Известный handle не равен поиску узла по индексу.")
	fmt.Fprintln(&out, "Медианы сопоставимы внутри одинакового контракта и окружения. Benchmark не доказывает Big O и не выбирает универсально лучшую структуру.")
	return out.String(), nil
}

func table(cells [][]string) string {
	widths := make([]int, len(cells[0]))
	for i := range widths {
		widths[i] = 3 // Не меньше трёх дефисов в разделителе Markdown-таблицы.
	}
	for _, row := range cells {
		for i, cell := range row {
			if n := utf8.RuneCountInString(cell); n > widths[i] {
				widths[i] = n
			}
		}
	}
	var result strings.Builder
	for j, row := range cells {
		result.WriteString("| ")
		for i, cell := range row {
			result.WriteString(cell)
			result.WriteString(strings.Repeat(" ", widths[i]-utf8.RuneCountInString(cell)))
			result.WriteString(" | ")
		}
		result.WriteByte('\n')
		if j == 0 {
			result.WriteString("| ")
			for _, width := range widths {
				result.WriteString(strings.Repeat("-", width))
				result.WriteString(" | ")
			}
			result.WriteByte('\n')
		}
	}
	return result.String()
}
