package demo

import "github.com/rinatkh/classroom8/internal/buffer"

// BufferTrace снимает состояние реального Linear через публичный API.
// Values — логический непрочитанный хвост, а не физические адреса backing array.
// Capacity показана отдельно; копирование для снимка не является benchmark.
func BufferTrace() []Frame {
	var b buffer.Linear
	var frames []Frame
	capture := func(title, note string) {
		var values []int
		for _, value := range b.Bytes() {
			values = append(values, int(value))
		}
		frames = append(frames, Frame{Title: title, Note: note, Values: values, Size: b.Len(), Capacity: b.Cap()})
	}
	capture("Пустой Linear", "len=0, capacity=0; память под данные ещё не выделена")
	b.Write([]byte("Go "))
	capture("Write(Go )", "Байты скопированы в собственный буфер; capacity может быть больше len")
	b.Write([]byte("API"))
	capture("Write(API)", "Накоплен весь payload; это растущий буфер, а не bounded ring")
	b.Read(make([]byte, 3))
	capture("Read(3) → Go ", "Остался непрочитанный API; backing array и capacity продолжают существовать")
	b.Write([]byte("abcdefghijkl"))
	capture("Write(12 байт)", "При нехватке места возможны compaction и новый массив; capacity взята из реального Go-вызова")
	b.Reset()
	capture("Reset()", "Непрочитанных данных нет, но capacity сохранена: len=0 не означает освобождение массива")
	b.Write([]byte("OK"))
	capture("Write(OK)", "Повторно используется сохранённая ёмкость")
	b.Read(make([]byte, 2))
	capture("Read(2) → OK", "Все данные прочитаны; capacity остаётся для следующих записей")
	return frames
}
