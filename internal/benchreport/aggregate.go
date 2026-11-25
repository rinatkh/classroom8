package benchreport

import (
	"encoding/csv"
	"fmt"
	"io"
	"sort"
	"strconv"
)

// Row хранит разные статистики отдельно. Основная таблица использует медианы,
// а CSV также сохраняет арифметические средние и диапазон времени.
type Row struct {
	Environment                                       Environment
	Name                                              Name
	Repeats                                           int
	MeanNS, MeanBytes, MeanAllocs                     float64
	MedianNS, MedianBytes, MedianAllocs, MinNS, MaxNS float64
}
type key struct {
	Environment
	Name
}

func median(values []float64) float64 {
	sort.Float64s(values)
	middle := len(values) / 2
	if len(values)%2 == 1 {
		return values[middle]
	}
	return values[middle-1]/2 + values[middle]/2
}

// Aggregate не смешивает размер, тип параметра, операцию, реализацию и CPU.
// b.N не является ни размером входа, ни весом отдельных прогонов.
func Aggregate(samples []Sample, expected int) ([]Row, error) {
	if len(samples) == 0 || expected < 0 {
		return nil, fmt.Errorf("нет измерений или неверное число повторов")
	}
	groups := map[key][]Sample{}
	for _, s := range samples {
		k := key{s.Environment, s.Name}
		groups[k] = append(groups[k], s)
	}
	rows := make([]Row, 0, len(groups))
	for k, group := range groups {
		if expected > 0 && len(group) != expected {
			return nil, fmt.Errorf("%s/%s/%s %s=%d: %d прогонов вместо %d", k.Root, k.Operation, k.Implementation, k.Dimension, k.Size, len(group), expected)
		}
		row := Row{Environment: k.Environment, Name: k.Name, Repeats: len(group)}
		times, bytes, allocs := make([]float64, len(group)), make([]float64, len(group)), make([]float64, len(group))
		for i, s := range group {
			count := float64(i + 1)
			row.MeanNS += (s.NS - row.MeanNS) / count
			row.MeanBytes += (s.Bytes - row.MeanBytes) / count
			row.MeanAllocs += (s.Allocs - row.MeanAllocs) / count
			times[i], bytes[i], allocs[i] = s.NS, s.Bytes, s.Allocs
		}
		row.MedianNS, row.MedianBytes, row.MedianAllocs = median(times), median(bytes), median(allocs)
		row.MinNS, row.MaxNS = times[0], times[len(times)-1]
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool { return rowLess(rows[i], rows[j]) })
	return rows, nil
}

func environmentKey(e Environment) string {
	return e.GOOS + "\x00" + e.GOARCH + "\x00" + e.Package + "\x00" + e.CPU
}
func rowLess(a, b Row) bool {
	if a.Environment != b.Environment {
		return environmentKey(a.Environment) < environmentKey(b.Environment)
	}
	x, y := a.Name, b.Name
	if x.Procs != y.Procs {
		return x.Procs < y.Procs
	}
	if x.Root != y.Root {
		return x.Root < y.Root
	}
	if x.Dimension != y.Dimension {
		return x.Dimension < y.Dimension
	}
	if x.Size != y.Size {
		return x.Size < y.Size
	}
	if x.Operation != y.Operation {
		return x.Operation < y.Operation
	}
	return x.Implementation < y.Implementation
}

func WriteCSV(w io.Writer, rows []Row) error {
	output := csv.NewWriter(w)
	if err := output.Write([]string{"goos", "goarch", "pkg", "cpu", "gomaxprocs", "benchmark", "operation", "implementation", "dimension", "size", "repeats", "median_ns_per_op", "mean_ns_per_op", "min_ns_per_op", "max_ns_per_op", "median_B_per_op", "mean_B_per_op", "median_allocs_per_op", "mean_allocs_per_op"}); err != nil {
		return err
	}
	for _, row := range rows {
		n, e := row.Name, row.Environment
		record := []string{e.GOOS, e.GOARCH, e.Package, e.CPU, strconv.Itoa(n.Procs), n.Root, n.Operation, n.Implementation, n.Dimension, strconv.Itoa(n.Size), strconv.Itoa(row.Repeats)}
		for _, v := range []float64{row.MedianNS, row.MeanNS, row.MinNS, row.MaxNS, row.MedianBytes, row.MeanBytes, row.MedianAllocs, row.MeanAllocs} {
			record = append(record, strconv.FormatFloat(v, 'g', -1, 64))
		}
		if err := output.Write(record); err != nil {
			return err
		}
	}
	output.Flush()
	return output.Error()
}
