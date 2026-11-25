// Package benchreport разбирает результаты testing.B и сравнивает одинаковые опыты.
// Сами измерения остаются в internal/benchmark; этот пакет не запускает алгоритмы.
package benchreport

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// Environment входит в ключ группировки: данные разных машин не усредняются.
// Версию Go и параметры запуска сохраняет отдельный environment.json.
type Environment struct {
	GOOS, GOARCH, Package, CPU string
}

// Name описывает одну операцию. Procs — суффикс GOMAXPROCS, не размер массива.
type Name struct {
	Root, Operation, Implementation, Dimension string
	Size, Procs                                int
}

// Sample — один прогон. Iterations сохраняется для проверки сырого формата,
// но НЕ участвует в среднем: ns/op уже нормирован на b.N самим testing.B.
type Sample struct {
	Environment       Environment
	Name              Name
	Iterations        uint64
	NS, Bytes, Allocs float64
}

var procsSuffix = regexp.MustCompile(`-([0-9]+)$`)

func positive(text string) (int, error) {
	n, err := strconv.Atoi(text)
	if err != nil || n < 1 {
		return 0, fmt.Errorf("ожидалось положительное целое: %q", text)
	}
	return n, nil
}

// ParseName принимает только схемы имён существующей лаборатории. Неизвестный
// алгоритм или лишнее измерение нельзя незаметно потерять при форматировании.
func ParseName(raw string) (Name, error) {
	n := Name{Procs: 1}
	if match := procsSuffix.FindStringSubmatch(raw); match != nil {
		var err error
		n.Procs, err = positive(match[1])
		if err != nil {
			return Name{}, err
		}
		raw = strings.TrimSuffix(raw, match[0])
	}
	parts := strings.Split(raw, "/")
	if len(parts) < 3 {
		return Name{}, fmt.Errorf("неполное имя benchmark: %q", raw)
	}
	n.Root = parts[0]
	var dimension string
	switch {
	case n.Root == "BenchmarkMemory" && len(parts) == 3:
		n.Operation, n.Implementation, dimension = "build", parts[1], parts[2]
	case n.Root == "BenchmarkBuffer" && len(parts) == 3 && parts[1] == "ring-cyclic-64-pairs":
		n.Operation, n.Implementation, dimension = parts[1], "ring", parts[2]
	case len(parts) == 4 && n.Root != "BenchmarkMemory" && !(n.Root == "BenchmarkBuffer" && parts[1] == "ring-cyclic-64-pairs"):
		n.Operation, n.Implementation, dimension = parts[1], parts[2], parts[3]
	default:
		return Name{}, fmt.Errorf("неизвестная схема benchmark: %q", raw)
	}
	spec, ok := specs[n.Root+"/"+n.Operation]
	if !ok || !oneOf(n.Implementation, spec.Implementations...) {
		return Name{}, fmt.Errorf("неизвестная операция/реализация: %q", raw)
	}
	key, value, ok := strings.Cut(dimension, "=")
	if !ok || key != spec.Dimension {
		return Name{}, fmt.Errorf("неверный параметр размера: %q", raw)
	}
	var err error
	n.Size, err = positive(value)
	if err != nil {
		return Name{}, err
	}
	n.Dimension = key
	return n, nil
}

func oneOf(value string, choices ...string) bool {
	for _, choice := range choices {
		if value == choice {
			return true
		}
	}
	return false
}

func parseSample(line string, env Environment) (Sample, error) {
	fields := strings.Fields(line)
	if len(fields) != 8 {
		return Sample{}, fmt.Errorf("ожидались имя, b.N и три метрики: %q", line)
	}
	name, err := ParseName(fields[0])
	if err != nil {
		return Sample{}, err
	}
	iterations, err := strconv.ParseUint(fields[1], 10, 64)
	if err != nil || iterations == 0 {
		return Sample{}, fmt.Errorf("некорректный b.N: %q", fields[1])
	}
	metrics := map[string]float64{}
	for i := 2; i < len(fields); i += 2 {
		value, err := strconv.ParseFloat(fields[i], 64)
		unit := fields[i+1]
		_, duplicate := metrics[unit]
		if err != nil || math.IsNaN(value) || math.IsInf(value, 0) || value < 0 || duplicate ||
			!oneOf(unit, "ns/op", "B/op", "allocs/op") {
			return Sample{}, fmt.Errorf("некорректная метрика: %q", line)
		}
		metrics[unit] = value
	}
	return Sample{env, name, iterations, metrics["ns/op"], metrics["B/op"], metrics["allocs/op"]}, nil
}

// Parse требует завершённый вывод go test одного пакета с PASS и итоговой ok.
// Прерванный, смешанный или повреждённый лог не превращается в успешную сводку.
// Для другого окружения нужен отдельный файл; объединение CSV выполняют явно.
func Parse(reader io.Reader) ([]Sample, error) {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	var env Environment
	var samples []Sample
	pass, complete := false, false
	seenHeaders := map[string]bool{}
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if complete {
			return nil, fmt.Errorf("строка %d: данные после итоговой ok; используйте отдельные логи", lineNumber)
		}
		if strings.HasPrefix(line, "Benchmark") {
			if pass || env.GOOS == "" || env.GOARCH == "" || env.Package == "" {
				return nil, fmt.Errorf("строка %d: нет окружения или benchmark после PASS", lineNumber)
			}
			sample, err := parseSample(line, env)
			if err != nil {
				return nil, fmt.Errorf("строка %d: %w", lineNumber, err)
			}
			samples = append(samples, sample)
			continue
		}
		if line == "PASS" {
			if pass || len(samples) == 0 {
				return nil, fmt.Errorf("строка %d: неожиданный PASS", lineNumber)
			}
			pass = true
			continue
		}
		fields := strings.Fields(line)
		if fields[0] == "ok" && len(fields) == 3 && fields[1] == env.Package && strings.HasSuffix(fields[2], "s") && pass {
			duration, err := strconv.ParseFloat(strings.TrimSuffix(fields[2], "s"), 64)
			if err == nil && duration >= 0 && !math.IsNaN(duration) && !math.IsInf(duration, 0) {
				complete = true
				continue
			}
		}
		header, value, ok := strings.Cut(line, ":")
		if ok && oneOf(header, "goos", "goarch", "pkg", "cpu") && len(samples) == 0 && !seenHeaders[header] && strings.TrimSpace(value) != "" {
			seenHeaders[header] = true
			value = strings.TrimSpace(value)
			switch header {
			case "goos":
				env.GOOS = value
			case "goarch":
				env.GOARCH = value
			case "pkg":
				env.Package = value
			case "cpu":
				env.CPU = value
			}
			continue
		}
		return nil, fmt.Errorf("строка %d: ошибка или неподдерживаемый вывод go test: %q", lineNumber, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("чтение benchmark: %w", err)
	}
	if !complete || len(samples) == 0 {
		return nil, fmt.Errorf("незавершённый benchmark: нужны измерения, PASS и итоговая ok")
	}
	return samples, nil
}
