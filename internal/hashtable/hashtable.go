// Package hashtable — учебная string->int таблица с separate chaining.
// Это не описание внутреннего устройства Go map и не замена map для production.
package hashtable

// entry хранит hash, чтобы rehash не пересчитывал все байты строк заново.
type entry struct {
	key   string
	value int
	hash  uint64
	next  *entry
}
type Pair struct {
	Key   string
	Value int
}

// Table использует массив голов цепочек. n записей, m buckets: память O(n+m).
// Ожидаемая цена — при хорошем распределении и ограниченной длине ключа.
// Для строк длины k hash и сравнение до O(k): worst цепочки до O(n*k).
// Нулевая Table допустима. После первого Set Table нельзя копировать.
type Table struct {
	buckets []*entry
	size    int
	hasher  func(string) uint64
}

func New(capacity int) *Table { return NewWithHasher(capacity, nil) }

// NewWithHasher позволяет принудительно столкнуть ключи в лаборатории/тестах.
// Детерминированная FNV-1a удобна для воспроизводимости, не защищает от атак на hash.
func NewWithHasher(capacity int, hash func(string) uint64) *Table {
	if capacity < 4 {
		capacity = 4
	}
	return &Table{buckets: make([]*entry, capacity), hasher: hash}
}
func hashString(key string) uint64 {
	var h uint64 = 14695981039346656037
	for i := 0; i < len(key); i++ {
		h ^= uint64(key[i])
		h *= 1099511628211
	}
	return h
}
func (t *Table) hash(key string) uint64 {
	if t.hasher != nil {
		return t.hasher(key)
	}
	return hashString(key)
}
func (t *Table) Len() int      { return t.size }
func (t *Table) Capacity() int { return len(t.buckets) }
func (t *Table) LoadFactor() float64 {
	if t.Capacity() == 0 {
		return 0
	}
	return float64(t.size) / float64(t.Capacity())
}

// Set обновляет существующую пару, не меняя Len. Рост — до вставки нового ключа,
// если load factor превысил бы 0.75. Один рост O(n+m), вставка амортизированно
// ожидаемо O(1) только при ограниченной длине ключа и хорошем распределении.
func (t *Table) Set(key string, value int) {
	if len(t.buckets) == 0 {
		t.buckets = make([]*entry, 4)
	}
	hash := t.hash(key)
	index := int(hash % uint64(len(t.buckets)))
	for e := t.buckets[index]; e != nil; e = e.next {
		if e.hash == hash && e.key == key {
			e.value = value
			return
		}
	}
	if t.size+1 > len(t.buckets)*3/4 {
		t.resize(2 * len(t.buckets))
		index = int(hash % uint64(len(t.buckets)))
	}
	t.buckets[index] = &entry{key, value, hash, t.buckets[index]}
	t.size++
}
func (t *Table) Get(key string) (int, bool) {
	if len(t.buckets) == 0 {
		return 0, false
	}
	hash := t.hash(key)
	for e := t.buckets[int(hash%uint64(len(t.buckets)))]; e != nil; e = e.next {
		if e.hash == hash && e.key == key {
			return e.value, true
		}
	}
	return 0, false
}

// Delete не уменьшает массив buckets. Отсутствие — false; 0 как значение допустим.
func (t *Table) Delete(key string) bool {
	if len(t.buckets) == 0 {
		return false
	}
	hash := t.hash(key)
	index := int(hash % uint64(len(t.buckets)))
	previous := &t.buckets[index] // Указатель на связь убирает особый случай головы.
	for *previous != nil {
		e := *previous
		if e.hash == hash && e.key == key {
			*previous = e.next
			e.key = ""
			e.value = 0
			e.hash = 0
			e.next = nil
			t.size--
			return true
		}
		previous = &e.next
	}
	return false
}

// resize переиспользует узлы, но меняет цепочки согласно НОВОМУ числу buckets.
// Просто скопировать массив голов нельзя: hash % m меняется. Временная память O(m).
func (t *Table) resize(capacity int) {
	buckets := make([]*entry, capacity)
	for _, head := range t.buckets {
		for e := head; e != nil; {
			next := e.next
			index := int(e.hash % uint64(capacity))
			e.next = buckets[index]
			buckets[index] = e
			e = next
		}
	}
	t.buckets = buckets
}

// Snapshot — учебный независимый снимок цепочек, не измеряемая операция lookup.
func (t *Table) Snapshot() [][]Pair {
	out := make([][]Pair, len(t.buckets))
	for i, head := range t.buckets {
		for e := head; e != nil; e = e.next {
			out[i] = append(out[i], Pair{e.key, e.value})
		}
	}
	return out
}
