// Package benchmark измеряет одинаковые контракты и явно называет пары операций.
// Все данные воспроизводимы (seed 7). Go 1.22: b.N, без b.Loop.
package benchmark

import (
	"bytes"
	"container/list"
	"fmt"
	"os"
	"runtime"
	"slices"
	"strconv"
	"testing"

	"github.com/rinatkh/classroom8/internal/buffer"
	"github.com/rinatkh/classroom8/internal/dataset"
	"github.com/rinatkh/classroom8/internal/hashtable"
	"github.com/rinatkh/classroom8/internal/linkedlist"
	"github.com/rinatkh/classroom8/internal/queue"
	"github.com/rinatkh/classroom8/internal/ringbuffer"
	"github.com/rinatkh/classroom8/internal/stack"
)

var sinkInt int
var sinkBool bool
var sinkSlice []int
var sinkList *linkedlist.List[int]
var sinkPairs []pair
var sinkTable *hashtable.Table
var sinkMap map[string]int
var sinkBytes []byte
var sinkLinear *buffer.Linear
var sinkBuffer *bytes.Buffer
var sinkStandardList *list.List
var sinkAny any

// selectedSize применяется до генерации и ResetTimer: измеряемая работа
// не меняется. Пустой фильтр сохраняет исходную полную матрицу.
func selectedSize(size int) bool {
	filter := os.Getenv("LESSON8_BENCH_SIZE")
	return filter == "" || filter == strconv.Itoa(size)
}

type pair struct {
	key   string
	value int
}

func newList(values []int) *linkedlist.List[int] {
	l := new(linkedlist.List[int])
	for _, v := range values {
		l.Append(v)
	}
	return l
}

// Lookup ищет наличие одного отсутствующего значения; обход суммирует те же ints.
// Подготовка и проверки вне таймера. Повторяемые цели прогревают cache/predictor.
func BenchmarkList(b *testing.B) {
	for _, n := range []int{100, 1000, 10000} {
		if !selectedSize(n) {
			continue
		}
		values := dataset.GenerateRandomSlice(n)
		for _, op := range []string{"search-missing", "walk"} {
			for _, impl := range []string{"slice", "linked"} {
				b.Run(fmt.Sprintf("%s/%s/n=%d", op, impl, n), func(b *testing.B) {
					l := newList(values)
					want := 0
					for _, v := range values {
						want += v
					}
					got := 0
					b.ReportAllocs()
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						if op == "search-missing" {
							if impl == "slice" {
								sinkBool = slices.Contains(values, 1234567)
							} else {
								sinkBool = l.Find(1234567) != nil
							}
						} else {
							sum := 0
							if impl == "slice" {
								for _, v := range values {
									sum += v
								}
							} else {
								for p := l.Head(); p != nil; p = p.Next {
									sum += p.Value
								}
							}
							got = sum
							sinkInt = sum
						}
					}
					b.StopTimer()
					if op == "walk" && got != want {
						b.Fatal("sum")
					}
					if op == "search-missing" && sinkBool {
						b.Fatal("unexpected match")
					}
					runtime.KeepAlive(l)
					runtime.KeepAlive(values)
				})
			}
		}
	}
	// Здесь одна op — законченная конструкция из n вставок, включая allocation.
	for _, n := range []int{100, 1000} {
		if !selectedSize(n) {
			continue
		}
		values := dataset.GenerateRandomSlice(n)
		for _, op := range []string{"build-append", "build-prepend"} {
			for _, impl := range []string{"slice", "linked"} {
				b.Run(fmt.Sprintf("%s/%s/n=%d", op, impl, n), func(b *testing.B) {
					b.ReportAllocs()
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						if impl == "slice" {
							v := make([]int, 0, n)
							for _, x := range values {
								if op == "build-append" {
									v = append(v, x)
								} else {
									v = append(v, 0)
									copy(v[1:], v[:len(v)-1])
									v[0] = x
								}
							}
							sinkSlice = v
						} else {
							l := new(linkedlist.List[int])
							for _, x := range values {
								if op == "build-append" {
									l.Append(x)
								} else {
									l.Prepend(x)
								}
							}
							sinkList = l
						}
					}
					b.StopTimer()
					want := slices.Clone(values)
					if op == "build-prepend" {
						slices.Reverse(want)
					}
					if impl == "slice" {
						if !slices.Equal(sinkSlice, want) {
							b.Fatal("build")
						}
					} else if !slices.Equal(sinkList.Values(), want) {
						b.Fatal("build")
					}
				})
			}
		}
		// Одна op — insert+undo либо delete+restore, обе части внутри таймера.
		// Так размер не растёт с b.N, и нет дорогого StopTimer на каждом элементе.
		for _, op := range []string{"prepend-undo", "append-undo", "middle-known-undo", "middle-index-undo", "delete-value-last-restore"} {
			for _, impl := range []string{"slice", "linked"} {
				b.Run(fmt.Sprintf("%s/%s/n=%d", op, impl, n), func(b *testing.B) {
					// Уникальные значения нужны для однозначного удаления последнего по значению.
					source := make([]int, n)
					for i := range source {
						source[i] = i
					}
					work := make([]int, n, n+1)
					copy(work, source)
					l := newList(source)
					middle, _ := l.At(n / 2)
					tail := l.Tail()
					position := n/2 + 1
					b.ReportAllocs()
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						value := -(i + 1)
						if impl == "slice" {
							switch op {
							case "prepend-undo", "middle-known-undo", "middle-index-undo":
								pos := position
								if op == "prepend-undo" {
									pos = 0
								}
								work = append(work, 0)
								copy(work[pos+1:], work[pos:n])
								work[pos] = value
								sinkInt = work[pos]
								copy(work[pos:], work[pos+1:])
								work[n] = 0
								work = work[:n]
							case "append-undo":
								work = append(work, value)
								sinkInt = work[n]
								work[n] = 0
								work = work[:n]
							case "delete-value-last-restore":
								index := slices.Index(work, n-1)
								sinkInt = work[index]
								copy(work[index:], work[index+1:])
								work[n-1] = 0
								work = work[:n-1]
								work = append(work, n-1)
							}
						} else {
							switch op {
							case "prepend-undo":
								l.Prepend(value)
								sinkInt, _ = l.PopFront()
							case "append-undo":
								l.Append(value)
								sinkInt, _ = l.DeleteAfter(tail)
							case "middle-known-undo":
								l.InsertAfter(middle, value)
								sinkInt, _ = l.DeleteAfter(middle)
							case "middle-index-undo":
								at, _ := l.At(n / 2)
								l.InsertAfter(at, value)
								sinkInt, _ = l.DeleteAfter(at)
							case "delete-value-last-restore":
								sinkBool = l.RemoveFirst(n - 1)
								l.Append(n - 1)
							}
						}
					}
					b.StopTimer()
					if !slices.Equal(work, source) || !slices.Equal(l.Values(), source) {
						b.Fatal("restoration")
					}
					runtime.KeepAlive(work)
					runtime.KeepAlive(l)
				})
			}
		}
	}
}

type lifo interface {
	Push(int)
	Pop() (int, bool)
	Len() int
}

func BenchmarkStack(b *testing.B) {
	for _, n := range []int{32, 256} {
		if !selectedSize(n) {
			continue
		}
		values := dataset.GenerateRandomSlice(n)
		want := 0
		for _, value := range values {
			want += value
		}
		for _, impl := range []string{"slice", "linked"} {
			b.Run(fmt.Sprintf("push-pop/%s/n=%d", impl, n), func(b *testing.B) {
				var s lifo
				if impl == "slice" {
					s = new(stack.Slice[int])
				} else {
					s = new(stack.Linked[int])
				}
				for _, v := range values {
					s.Push(v)
				}
				for range values {
					s.Pop()
				}
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					sum := 0
					for _, v := range values {
						s.Push(v)
					}
					for range values {
						v, ok := s.Pop()
						if !ok {
							b.Fatal("empty")
						}
						sum += v
					}
					sinkInt = sum
				}
				b.StopTimer()
				if s.Len() != 0 || sinkInt != want {
					b.Fatal("len")
				}
				sinkAny = s
			})
		}
	}
}

type fifo interface {
	Enqueue(int)
	Dequeue() (int, bool)
	Front() (int, bool)
	Len() int
}

// Ring имеет bool-return Enqueue; адаптер одинаково проверяет неудачу записи.
type ringQueue struct{ r *ringbuffer.Ring[int] }

func (q ringQueue) Enqueue(v int) {
	if !q.r.Enqueue(v) {
		panic("benchmark capacity")
	}
}
func (q ringQueue) Dequeue() (int, bool) { return q.r.Dequeue() }
func (q ringQueue) Front() (int, bool)   { return q.r.Front() }
func (q ringQueue) Len() int             { return q.r.Len() }

// Одна op — 64 пары Enqueue/Dequeue при длине n. Все модели сохраняют FIFO;
// Ring bounded, но capacity n+1 гарантирует принятие всех записей этой нагрузки.
func BenchmarkQueue(b *testing.B) {
	for _, n := range []int{100, 1000} {
		if !selectedSize(n) {
			continue
		}
		for _, impl := range []string{"slice-shift", "linked", "ring"} {
			b.Run(fmt.Sprintf("64-pairs/%s/n=%d", impl, n), func(b *testing.B) {
				var q fifo
				switch impl {
				case "slice-shift":
					q = new(queue.Slice[int])
				case "linked":
					q = new(queue.Linked[int])
				case "ring":
					r, _ := ringbuffer.New[int](n + 1)
					q = ringQueue{r}
				}
				for i := 0; i <= n; i++ {
					q.Enqueue(i)
				}
				q.Dequeue()
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					sum := 0
					for j := 0; j < 64; j++ {
						q.Enqueue(i + j)
						v, ok := q.Dequeue()
						if !ok {
							b.Fatal("empty")
						}
						sum += v
					}
					sinkInt = sum
				}
				b.StopTimer()
				if q.Len() != n {
					b.Fatal("growing queue")
				}
				front, ok := q.Front()
				removed := int64(b.N) * 64
				wantFront := 1 + removed
				if removed >= int64(n) {
					offset := removed - int64(n)
					wantFront = offset/64 + offset%64
				}
				if !ok || int64(front) != wantFront {
					b.Fatal("front/FIFO")
				}
				sinkAny = q
			})
		}
	}
}

// Hash использует одинаковые string->int ключи. Стоимость hash строки зависит
// от длины ключа, здесь ключи одной длины. Map уже построена в lookup/mutation.
func BenchmarkHash(b *testing.B) {
	for _, n := range []int{100, 1000, 10000} {
		if !selectedSize(n) {
			continue
		}
		keys := dataset.Keys(n)
		for _, impl := range []string{"custom", "map", "slice"} {
			b.Run(fmt.Sprintf("build/%s/n=%d", impl, n), func(b *testing.B) {
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					switch impl {
					case "custom":
						t := hashtable.New(4)
						for j, k := range keys {
							t.Set(k, j)
						}
						sinkTable = t
					case "map":
						m := make(map[string]int)
						for j, k := range keys {
							m[k] = j
						}
						sinkMap = m
					case "slice":
						p := make([]pair, 0, n)
						for j, k := range keys {
							p = append(p, pair{k, j})
						}
						sinkPairs = p
					}
				}
				b.StopTimer()
				switch impl {
				case "custom":
					if sinkTable.Len() != n {
						b.Fatal("len")
					}
				case "map":
					if len(sinkMap) != n {
						b.Fatal("len")
					}
				case "slice":
					if len(sinkPairs) != n {
						b.Fatal("len")
					}
				}
			})
		}
		for _, op := range []string{"lookup-last", "lookup-missing", "insert-read-delete-triple", "delete-restore-pair"} {
			for _, impl := range []string{"custom", "map", "slice"} {
				b.Run(fmt.Sprintf("%s/%s/n=%d", op, impl, n), func(b *testing.B) {
					table := hashtable.New(2 * n)
					m := make(map[string]int, 2*n)
					pairs := make([]pair, 0, n+1)
					for j, k := range keys {
						table.Set(k, j)
						m[k] = j
						pairs = append(pairs, pair{k, j})
					}
					key := keys[n-1]
					if op == "lookup-missing" {
						key = "absent-xxxxx"
					}
					newKey := "new-key-xxxx"
					b.ReportAllocs()
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						switch op {
						case "lookup-last", "lookup-missing":
							switch impl {
							case "custom":
								sinkInt, sinkBool = table.Get(key)
							case "map":
								sinkInt, sinkBool = m[key]
							case "slice":
								value, ok := 0, false
								for _, p := range pairs {
									if p.key == key {
										value, ok = p.value, true
										break
									}
								}
								sinkInt, sinkBool = value, ok
							}
						case "insert-read-delete-triple":
							switch impl {
							case "custom":
								table.Set(newKey, i)
								sinkInt, _ = table.Get(newKey)
								sinkBool = table.Delete(newKey)
							case "map":
								m[newKey] = i
								sinkInt = m[newKey]
								delete(m, newKey)
							case "slice":
								pairs = append(pairs, pair{newKey, i})
								index := 0
								for pairs[index].key != newKey {
									index++
								}
								sinkInt = pairs[index].value
								copy(pairs[index:], pairs[index+1:])
								pairs[n] = pair{}
								pairs = pairs[:n]
							}
						case "delete-restore-pair":
							switch impl {
							case "custom":
								sinkBool = table.Delete(key)
								table.Set(key, n-1)
							case "map":
								delete(m, key)
								m[key] = n - 1
							case "slice":
								index := 0
								for pairs[index].key != key {
									index++
								}
								sinkInt = pairs[index].value
								copy(pairs[index:], pairs[index+1:])
								pairs[n-1] = pair{}
								pairs = pairs[:n-1]
								pairs = append(pairs, pair{key, n - 1})
							}
						}
					}
					b.StopTimer()
					if table.Len() != n || len(m) != n || len(pairs) != n {
						b.Fatal("size")
					}
					if op == "lookup-last" && (!sinkBool || sinkInt != n-1) {
						b.Fatal("lookup")
					}
					if op == "lookup-missing" && sinkBool {
						b.Fatal("missing")
					}
					// Проверяем результат мутаций после таймера: Len сам по себе
					// не доказывает ни корректное чтение, ни восстановление пары.
					get := func(k string) (int, bool) {
						switch impl {
						case "custom":
							return table.Get(k)
						case "map":
							v, ok := m[k]
							return v, ok
						default:
							for _, p := range pairs {
								if p.key == k {
									return p.value, true
								}
							}
							return 0, false
						}
					}
					if op == "insert-read-delete-triple" {
						_, exists := get(newKey)
						if sinkInt != b.N-1 || exists || impl == "custom" && !sinkBool {
							b.Fatal("triple result")
						}
					}
					if op == "delete-restore-pair" {
						value, exists := get(key)
						if !exists || value != n-1 || impl == "custom" && !sinkBool {
							b.Fatal("restore result")
						}
					}
					runtime.KeepAlive(table)
					runtime.KeepAlive(m)
					runtime.KeepAlive(pairs)
				})
			}
		}
	}
}

// Accumulate строит новый результат из одинаковых байтовых фрагментов.
// Глобальный sink сохраняет последнюю структуру, Bytes не копируется в таймере.
func BenchmarkBuffer(b *testing.B) {
	for _, n := range []int{100, 1000} {
		if !selectedSize(n) {
			continue
		}
		fragments := make([][]byte, n)
		for i := range fragments {
			fragments[i] = []byte(fmt.Sprintf("chunk-%09d\n", i))
		}
		for _, impl := range []string{"append", "bytes-buffer", "linear"} {
			b.Run(fmt.Sprintf("accumulate/%s/chunks=%d", impl, n), func(b *testing.B) {
				want := bytes.Join(fragments, nil)
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					switch impl {
					case "append":
						var out []byte
						for _, p := range fragments {
							out = append(out, p...)
						}
						sinkBytes = out
					case "bytes-buffer":
						out := new(bytes.Buffer)
						for _, p := range fragments {
							out.Write(p)
						}
						sinkBuffer = out
					case "linear":
						out := new(buffer.Linear)
						for _, p := range fragments {
							out.Write(p)
						}
						sinkLinear = out
					}
				}
				b.StopTimer()
				var got []byte
				switch impl {
				case "append":
					got = sinkBytes
				case "bytes-buffer":
					got = sinkBuffer.Bytes()
				case "linear":
					got = sinkLinear.Bytes()
				}
				if !bytes.Equal(got, want) {
					b.Fatal("bytes")
				}
			})
		}
	}
	// Другой контракт: bounded циклическое движение, не накопление всего потока.
	for _, capacity := range []int{32, 1024} {
		if !selectedSize(capacity) {
			continue
		}
		b.Run(fmt.Sprintf("ring-cyclic-64-pairs/cap=%d", capacity), func(b *testing.B) {
			r, _ := ringbuffer.New[byte](capacity)
			for i := 0; i < capacity-1; i++ {
				r.Enqueue(byte(i))
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				sum := 0
				for j := 0; j < 64; j++ {
					if !r.Enqueue(byte(i + j)) {
						b.Fatal("full")
					}
					v, ok := r.Dequeue()
					if !ok {
						b.Fatal("empty")
					}
					sum += int(v)
				}
				sinkInt = sum
			}
			b.StopTimer()
			if r.Len() != capacity-1 {
				b.Fatal("len")
			}
			// FIFO-модель без повторения Ring: после R чтений первый
			// непрочитанный байт имеет индекс R в общем входном потоке.
			removed := int64(b.N) * 64
			want := byte(removed)
			if removed >= int64(capacity-1) {
				offset := removed - int64(capacity-1)
				want = byte(offset/64 + offset%64)
			}
			if got, ok := r.Front(); !ok || got != want {
				b.Fatal("ring FIFO result")
			}
			sinkAny = r
		})
	}
}

// B/op — allocation, НЕ весь retained heap. container/list имеет другой API,
// два указателя и interface values: это reference по allocation, не чистый singly.
func BenchmarkMemory(b *testing.B) {
	for _, n := range []int{100, 1000, 10000} {
		if !selectedSize(n) {
			continue
		}
		values := dataset.GenerateRandomSlice(n)
		for _, impl := range []string{"slice-build", "singly-build", "std-doubly-build"} {
			b.Run(fmt.Sprintf("%s/n=%d", impl, n), func(b *testing.B) {
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					switch impl {
					case "slice-build":
						sinkSlice = slices.Clone(values)
					case "singly-build":
						sinkList = newList(values)
					case "std-doubly-build":
						l := list.New()
						for _, v := range values {
							l.PushBack(v)
						}
						sinkStandardList = l
					}
				}
				b.StopTimer()
				switch impl {
				case "slice-build":
					if !slices.Equal(sinkSlice, values) {
						b.Fatal("slice contents")
					}
				case "singly-build":
					if !slices.Equal(sinkList.Values(), values) {
						b.Fatal("list contents")
					}
				case "std-doubly-build":
					if sinkStandardList.Len() != n || sinkStandardList.Front().Value != values[0] || sinkStandardList.Back().Value != values[n-1] {
						b.Fatal("std list contents")
					}
				}
			})
		}
	}
}
