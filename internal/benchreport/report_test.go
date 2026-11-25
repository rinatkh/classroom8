package benchreport

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"go/types"
	"io"
	"math"
	"strings"
	"testing"
)

const header = "goos: linux\ngoarch: amd64\npkg: example/benchmark\ncpu: Test CPU\n"
const footer = "\nPASS\nok example/benchmark 0.1s\n"
const line = "BenchmarkList/search-missing/slice/n=1000 100000 1181 ns/op 0 B/op 0 allocs/op"

func names() []string {
	var result []string
	for key, spec := range specs {
		root, op, _ := strings.Cut(key, "/")
		sizes := []int{100, 1000}
		switch root {
		case "BenchmarkList":
			if op == "walk" || op == "search-missing" {
				sizes = append(sizes, 10000)
			}
		case "BenchmarkStack":
			sizes = []int{32, 256}
		case "BenchmarkHash", "BenchmarkMemory":
			sizes = append(sizes, 10000)
		case "BenchmarkBuffer":
			if op == "ring-cyclic-64-pairs" {
				sizes = []int{32, 1024}
			}
		}
		for _, size := range sizes {
			for _, impl := range spec.Implementations {
				name := root + "/" + op + "/" + impl + fmt.Sprintf("/%s=%d", spec.Dimension, size)
				if root == "BenchmarkMemory" {
					name = root + "/" + impl + fmt.Sprintf("/n=%d", size)
				}
				if op == "ring-cyclic-64-pairs" {
					name = root + "/" + op + fmt.Sprintf("/cap=%d", size)
				}
				result = append(result, name)
			}
		}
	}
	return result
}
func sample(t *testing.T, name string, ns float64) Sample {
	t.Helper()
	n, err := ParseName(name)
	if err != nil {
		t.Fatal(err)
	}
	return Sample{Environment: Environment{"linux", "amd64", "example/benchmark", "Test CPU"}, Name: n, Iterations: 10, NS: ns, Bytes: 8, Allocs: 1}
}
func TestAll112Names(t *testing.T) {
	all := names()
	if len(all) != 112 {
		t.Fatal(len(all))
	}
	for _, raw := range all {
		t.Run(raw, func(t *testing.T) {
			n, err := ParseName(raw + "-8")
			if err != nil || n.Procs != 8 || n.Size <= 0 || n.Dimension == "" || n.Operation == "" {
				t.Fatalf("%+v %v", n, err)
			}
		})
	}
	for _, test := range []struct {
		raw                 string
		root, op, impl, dim string
		size                int
	}{
		{"BenchmarkList/search-missing/linked/n=1000", "BenchmarkList", "search-missing", "linked", "n", 1000},
		{"BenchmarkStack/push-pop/slice/n=256", "BenchmarkStack", "push-pop", "slice", "n", 256},
		{"BenchmarkBuffer/accumulate/linear/chunks=100", "BenchmarkBuffer", "accumulate", "linear", "chunks", 100},
		{"BenchmarkBuffer/ring-cyclic-64-pairs/cap=1024", "BenchmarkBuffer", "ring-cyclic-64-pairs", "ring", "cap", 1024},
		{"BenchmarkMemory/std-doubly-build/n=1000", "BenchmarkMemory", "build", "std-doubly-build", "n", 1000},
	} {
		n, err := ParseName(test.raw)
		if err != nil || n.Root != test.root || n.Operation != test.op || n.Implementation != test.impl || n.Dimension != test.dim || n.Size != test.size {
			t.Fatalf("%+v %v", n, err)
		}
	}
}
func TestMalformedNames(t *testing.T) {
	for _, raw := range []string{"", "BenchmarkList", "BenchmarkList/op", "BenchmarkOther/op/slice/n=100", "BenchmarkList/no/slice/n=100", "BenchmarkList/walk/no/n=100", "BenchmarkList/walk/slice/chunks=100", "BenchmarkList/walk/slice/n=0", "BenchmarkList/walk/slice/n=-1", "BenchmarkList/walk/slice/n=x", "BenchmarkList/walk/slice/n=100/extra", "BenchmarkList/walk/slice/n=100-0", "BenchmarkList/walk/slice/n=100-99999999999999999999999", "BenchmarkMemory/build/slice-build/n=100", "BenchmarkMemory/no/n=100", "BenchmarkBuffer/ring-cyclic-64-pairs/cap=0", "BenchmarkBuffer/accumulate/linear/n=100"} {
		t.Run(raw, func(t *testing.T) {
			if _, err := ParseName(raw); err == nil {
				t.Fatal("accepted", raw)
			}
		})
	}
}
func TestParserMetrics(t *testing.T) {
	a := strings.ReplaceAll(line, "1181 ns/op 0 B/op 0 allocs/op", "1.136 ns/op 1024 B/op 2 allocs/op")
	b := strings.ReplaceAll(line, "1181 ns/op 0 B/op 0 allocs/op", "3 allocs/op 2.5e3 ns/op 2048 B/op")
	samples, err := Parse(strings.NewReader(header + a + "\n" + b + footer))
	if err != nil || len(samples) != 2 {
		t.Fatalf("%v %v", samples, err)
	}
	if samples[0].NS != 1.136 || samples[0].Bytes != 1024 || samples[0].Allocs != 2 || samples[0].Iterations != 100000 || samples[1].NS != 2500 {
		t.Fatal(samples)
	}
	if _, err := Parse(strings.NewReader(strings.ReplaceAll(strings.ReplaceAll(header+line+footer, "cpu: Test CPU\n", ""), "\n", "\r\n\r\n"))); err != nil {
		t.Fatal(err)
	}
}

type brokenReader struct{}

func (brokenReader) Read([]byte) (int, error) { return 0, errors.New("read error") }
func TestParserRejectsIncomplete(t *testing.T) {
	bad := map[string]string{
		"empty": "", "header": header, "no data": header + "PASS\nok example/benchmark 1s\n",
		"no pass": header + line + "\nok example/benchmark 1s\n", "no ok": header + line + "\nPASS\n",
		"missing env": line + footer, "repeat env": header + "goos: darwin\n" + line + footer,
		"env after data": header + line + "\ncpu: Other\n" + line + footer, "repeat pass": header + line + "\nPASS" + footer,
		"after pass": header + line + "\nPASS\n" + line + footer, "after ok": header + line + footer + "FAIL\n",
		"interrupted": header + line + "\n^Csignal: interrupt\nFAIL\n", "failure": header + line + "\nFAIL\n",
		"other pkg": header + line + "\nPASS\nok other/pkg 1s\n", "bad duration": header + line + "\nPASS\nok example/benchmark NaNs\n",
		"cached": header + line + "\nPASS\nok example/benchmark (cached)\n", "blank env": "goos:\n" + line + footer,
	}
	for _, value := range []string{"0", "-1", "1.5", "oops"} {
		bad["b.N "+value] = header + strings.ReplaceAll(line, "100000", value) + footer
	}
	for _, value := range []string{"NaN", "+Inf", "-1", "oops"} {
		bad["metric "+value] = header + strings.ReplaceAll(line, "1181", value) + footer
	}
	bad["unit"] = header + strings.ReplaceAll(line, "ns/op", "ms/op") + footer
	bad["duplicate unit"] = header + strings.ReplaceAll(line, "B/op", "ns/op") + footer
	bad["no benchmem"] = header + strings.ReplaceAll(line, " 0 B/op 0 allocs/op", "") + footer
	bad["bad name"] = header + strings.ReplaceAll(line, "BenchmarkList", "BenchmarkNo") + footer
	for name, log := range bad {
		t.Run(name, func(t *testing.T) {
			if got, err := Parse(strings.NewReader(log)); err == nil || got != nil {
				t.Fatalf("partial success: %+v %v", got, err)
			}
		})
	}
	for _, reader := range []io.Reader{brokenReader{}, strings.NewReader(strings.Repeat("x", 1024*1024+1))} {
		if _, err := Parse(reader); err == nil {
			t.Fatal("lost read error")
		}
	}
}
func TestAggregation(t *testing.T) {
	var samples []Sample
	for i, value := range []float64{1, 2, 9} {
		s := sample(t, "BenchmarkStack/push-pop/slice/n=256", value)
		s.Iterations = uint64(i*i*1000 + 1)
		s.Bytes = value * 100
		s.Allocs = value
		samples = append(samples, s)
	}
	rows, err := Aggregate(samples, 3)
	if err != nil || len(rows) != 1 {
		t.Fatalf("%v %v", rows, err)
	}
	r := rows[0]
	if r.MedianNS != 2 || r.MeanNS != 4 || r.MinNS != 1 || r.MaxNS != 9 || r.MedianBytes != 200 || r.MeanBytes != 400 || r.MedianAllocs != 2 || r.MeanAllocs != 4 || r.Repeats != 3 {
		t.Fatal(r)
	}
	even, _ := Aggregate(samples[:2], 0)
	if even[0].MedianNS != 1.5 {
		t.Fatal(even)
	}
	for _, test := range []struct {
		samples []Sample
		n       int
	}{{nil, 3}, {samples, -1}, {samples[:2], 3}} {
		if _, err := Aggregate(test.samples, test.n); err == nil {
			t.Fatal("accepted", test)
		}
	}
	base := samples[0]
	mixed := []Sample{base}
	for _, change := range []func(*Sample){func(s *Sample) { s.Name.Size = 32 }, func(s *Sample) { s.Name.Procs = 2 }, func(s *Sample) { s.Name.Implementation = "linked" }, func(s *Sample) { s.Name.Operation = "other" }, func(s *Sample) { s.Name.Root = "other" }, func(s *Sample) { s.Name.Dimension = "cap" }, func(s *Sample) { s.Environment.CPU = "other" }, func(s *Sample) { s.Environment.GOOS = "other" }, func(s *Sample) { s.Environment.Package = "other" }, func(s *Sample) { s.Environment.GOARCH = "other" }} {
		s := base
		change(&s)
		mixed = append(mixed, s)
	}
	separate, err := Aggregate(mixed, 1)
	if err != nil || len(separate) != len(mixed) {
		t.Fatal("mixed dimensions", separate, err)
	}
}

type badWriter struct{}

func (badWriter) Write([]byte) (int, error) { return 0, errors.New("write error") }
func TestCSV(t *testing.T) {
	s := sample(t, "BenchmarkQueue/64-pairs/ring/n=1000-2", 1.136)
	s.Environment.CPU = "CPU, \"quoted\""
	rows, _ := Aggregate([]Sample{s}, 1)
	var out bytes.Buffer
	if err := WriteCSV(&out, rows); err != nil {
		t.Fatal(err)
	}
	values, err := csv.NewReader(&out).ReadAll()
	if err != nil || len(values) != 2 || len(values[0]) != 19 {
		t.Fatalf("%v %v", values, err)
	}
	if values[1][3] != s.Environment.CPU || values[1][4] != "2" || values[1][6] != "64-pairs" || values[1][8] != "n" || values[1][11] != "1.136" {
		t.Fatal(values)
	}
	if err := WriteCSV(badWriter{}, rows); err == nil {
		t.Fatal("lost flush error")
	}
	rows[0].Environment.CPU = strings.Repeat("x", 10000)
	if err := WriteCSV(badWriter{}, rows); err == nil {
		t.Fatal("lost record error")
	}
}
func TestUnitsAndArchitecture(t *testing.T) {
	for _, test := range []struct {
		value       float64
		time, bytes string
	}{{0, "0 ns", "0 B"}, {1.136, "1.136 ns", "1.136 B"}, {999, "999 ns", "999 B"}, {1000, "1 µs", "1000 B"}, {1024, "1.024 µs", "1 KiB"}, {1e6, "1 ms", "976.6 KiB"}, {1048576, "1.049 ms", "1 MiB"}, {1e9, "1 s", "953.7 MiB"}, {1073741824, "1.074 s", "1 GiB"}, {0.00001, "1e-05 ns", "1e-05 B"}} {
		if FormatTime(test.value) != test.time || FormatBytes(test.value) != test.bytes {
			t.Fatal(test)
		}
	}
	for _, arch := range []string{"386", "amd64", "amd64p32", "arm", "arm64", "loong64", "mips", "mipsle", "mips64", "mips64le", "ppc64", "ppc64le", "riscv64", "s390x", "sparc64", "wasm"} {
		if int64(IntBytes(arch)) != types.SizesFor("gc", arch).Sizeof(types.Typ[types.Int]) {
			t.Fatal(arch)
		}
	}
	if IntBytes("unknown") != 0 {
		t.Fatal("guessed int width")
	}
}
func TestPattern(t *testing.T) {
	for kind, root := range roots {
		got, err := Pattern(kind, 0, "")
		if err != nil || got != "^"+root+"$" {
			t.Fatal(got, err)
		}
	}
	if p, err := Pattern("list", 1000, "walk"); err != nil || p != "^BenchmarkList$/^walk$" {
		t.Fatal(p, err)
	}
	if p, err := Pattern("memory", 100, "build"); err != nil || p != "^BenchmarkMemory$" {
		t.Fatal(p, err)
	}
	for _, test := range []struct {
		k, op string
		n     int
	}{{"other", "", 0}, {"list", "", -1}, {"list", "no", 100}} {
		if _, err := Pattern(test.k, test.n, test.op); err == nil {
			t.Fatal(test)
		}
	}
	if SizeFilter(0) != "" || SizeFilter(32) != "32" {
		t.Fatal("filter")
	}
}
func TestReportOperationUnits(t *testing.T) {
	var samples []Sample
	for i, name := range names() {
		samples = append(samples, sample(t, name, float64(i+1)))
	}
	// Неизвестная архитектура не заимствует размер int у машины formatter.
	other := sample(t, "BenchmarkList/walk/slice/n=256-2", 0)
	other.Environment.GOARCH = "unknown"
	other.Environment.CPU = ""
	samples = append(samples, other)
	copy := other
	copy.Name.Implementation = "linked"
	samples = append(samples, copy)
	rows, err := Aggregate(samples, 1)
	if err != nil {
		t.Fatal(err)
	}
	full, err := Render(rows, "full")
	if err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{"256 Push + 256 Pop", "64 Enqueue + 64 Dequeue", "3 действия", "2 действия", "1000 фрагментов", "cap=1024", "chunks=1000", "МЕДИАНЫ", "не b.N", "НЕ умножают", "готовый handle", "container/list", "неизвестен"} {
		if !strings.Contains(full, text) {
			t.Errorf("missing %q", text)
		}
	}
	compact, err := Render(rows, "compact")
	if err != nil || !strings.Contains(compact, "скрыто 51") {
		t.Fatalf("%v %s", err, compact[:600])
	}
	if strings.Count(full, "| Circular Buffer") != 4 {
		t.Fatal("lost ring rows")
	}
	for _, mode := range []string{"compact", "full"} {
		if _, err := Render(nil, mode); err == nil {
			t.Fatal("empty accepted")
		}
	}
	if _, err := Render(rows, "bad"); err == nil {
		t.Fatal("view accepted")
	}
	// Выбранная операция вне compact-набора должна остаться видимой.
	selected, _ := Aggregate([]Sample{sample(t, "BenchmarkHash/delete-restore-pair/custom/n=100", 2)}, 1)
	report, err := Render(selected, "compact")
	if err != nil || !strings.Contains(report, "Delete существующего") {
		t.Fatal(err, report)
	}
	// Медиана 2, а не среднее 4, действительно попадает в основной столбец.
	med, _ := Aggregate([]Sample{sample(t, "BenchmarkList/walk/slice/n=100", 1), sample(t, "BenchmarkList/walk/slice/n=100", 2), sample(t, "BenchmarkList/walk/slice/n=100", 9)}, 3)
	report, _ = Render(med, "full")
	if !strings.Contains(report, "2 ns") || strings.Contains(report, "4 ns") || !strings.Contains(report, "Min / op") || !strings.Contains(report, "Max / op") || !strings.Contains(report, "1 ns") || !strings.Contains(report, "9 ns") {
		t.Fatal(report)
	}
	var previous int
	for _, r := range rows {
		if r.Name.Root == "BenchmarkList" && r.Environment.GOARCH == "amd64" {
			if r.Name.Size < previous {
				t.Fatal("lexicographic n")
			}
			previous = r.Name.Size
		}
	}
	for _, cell := range strings.Split(strings.Trim(strings.Split(table([][]string{{"k"}, {"3"}}), "\n")[1], " |"), "|") {
		if len(strings.TrimSpace(cell)) < 3 {
			t.Fatal("invalid Markdown separator")
		}
	}
	if math.IsNaN(rows[0].MedianNS) {
		t.Fatal("invalid median")
	}
}
