package wikilang

import (
	"io"
	"unicode"
)

type CodeWriter interface {
	io.Writer

	FreshLine()
	NewLine()
	Indent()
	ChangeIndentation(i int)
	LiteralText(literal bool)
}

type StandardCodeWriter struct {
	writer io.Writer

	onNewLine   bool
	indentation int
	position    int

	maxPosition int

	useLiteralText bool
}

func NewCodeWriter(w io.Writer) *StandardCodeWriter {
	return &StandardCodeWriter{w, true, 0, 0, 80, false}
}

func (w *StandardCodeWriter) FreshLine() {
	if !w.onNewLine {
		w.NewLine()
	}
}

func (w *StandardCodeWriter) NewLine() {
	w.basicWrite([]byte("\n"))
}

func (w *StandardCodeWriter) Indent() {
	if w.onNewLine {
		spaces := make([]byte, w.indentation)
		for i, _ := range spaces {
			spaces[i] = ' '
		}
		w.basicWrite(spaces)
	}
}

func (w *StandardCodeWriter) ChangeIndentation(i int) {
	w.indentation += i
	if w.indentation < 0 {
		panic("CodeWriter indentation < 0")
	}
}

func (w *StandardCodeWriter) LiteralText(literal bool) {
	w.useLiteralText = literal
}

func (w *StandardCodeWriter) Write(b []byte) (n int, err error) {
	n = 0
	err = nil

	if w.useLiteralText {
		n, err = w.writer.Write(b)
	} else {
		lastWrite := 0
		for n < len(b) {
			nextWrite := n + w.maxPosition - w.position
			for ; nextWrite < len(b) && !unicode.IsSpace(rune(b[nextWrite])); nextWrite++ {
			}

			if nextWrite > len(b) {
				nextWrite = len(b)
			}
			written, innerErr := w.basicWrite(b[lastWrite:nextWrite])
			n += written
			if innerErr != nil {
				err = innerErr
				return
			}

			lastWrite = nextWrite

			for ; lastWrite < len(b) && unicode.IsSpace(rune(b[lastWrite])); lastWrite++ {
				n++
			}
		}
	}

	n = len(b)
	return
}

func (w *StandardCodeWriter) basicWrite(b []byte) (n int, err error) {
	n, err = w.writer.Write(b)
	if err != nil {
		return
	}

	if len(b) > 0 {
		w.onNewLine = b[len(b)-1] == '\n'
		if w.onNewLine {
			w.position = 0
		}

		if n+w.position >= w.maxPosition {
			w.NewLine()
		} else if b[len(b)-1] == '\n' {
			w.Indent()
			w.position = 0
		} else {
			w.position += n
		}
	}

	return
}
