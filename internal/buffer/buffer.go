// Package buffer показывает линейный растущий байтовый буфер и offset чтения.
package buffer

import "io"

// Linear накапливает байты. Это не bounded очередь и не текстовый rune-буфер.
// Память O(cap), Write k байт обычно O(k), иногда O(unread+k) из-за copy/grow.
type Linear struct {
	data   []byte
	offset int
}

func (b *Linear) Len() int { return len(b.data) - b.offset }
func (b *Linear) Cap() int { return cap(b.data) }

// Write копирует вход, поэтому последующая правка p не меняет буфер.
// Compaction переносит непрочитанный хвост, когда съедена хотя бы половина.
// Иначе append может увеличить массив. Это не гарантированное O(1) на вызов.
func (b *Linear) Write(p []byte) (int, error) {
	if len(p) > cap(b.data)-len(b.data) && b.offset > 0 && b.offset >= len(b.data)/2 {
		n := copy(b.data, b.data[b.offset:])
		clear(b.data[n:])
		b.data = b.data[:n]
		b.offset = 0
	}
	b.data = append(b.data, p...)
	return len(p), nil
}

// Read копирует байты в p. Очищает только прочитанные k байт; при исчерпании сбрасывает len/offset, сохраняя cap.
// Для p длины 0 возвращает (0,nil), даже у пустого буфера, как bytes.Buffer.
func (b *Linear) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if b.Len() == 0 {
		return 0, io.EOF
	}
	n := copy(p, b.data[b.offset:])
	clear(b.data[b.offset : b.offset+n]) // O(k), а не повторная очистка всей истории.
	b.offset += n
	if b.offset == len(b.data) {
		b.data = b.data[:0]
		b.offset = 0
	}
	return n, nil
}

// Bytes — независимая копия. ВАЖНО: стандартный bytes.Buffer.Bytes возвращает
// alias, действующий до следующего изменения. Контракты здесь различаются.
func (b *Linear) Bytes() []byte { return append([]byte(nil), b.data[b.offset:]...) }

// Reset очищает использованную область, но не возвращает capacity системе.
// Отбрасывание всего backing array потребовало бы b.data=nil.
func (b *Linear) Reset() { clear(b.data); b.data = b.data[:0]; b.offset = 0 }
