// Copyright (c) 2014, Daniel Gallagher
// Use of this source code is covered by the MIT License, the full
// text of which can be found in the LICENSE file.

package wikilang

import (
	"io"
	"unicode"
)

// A CodeWriter lets a code generator create structured text with
// specific indentation levels for each line.
type CodeWriter interface {
	io.Writer

	FreshLine()               // Write out a newline if we are not on a new line
	NewLine()                 // Write out a newline unconditionally
	Indent()                  // Write out indentation spaces
	ChangeIndentation(i int)  // Update the indentatino level
	LiteralText(literal bool) // Write text without changing it if true
}

type standardCodeWriter struct {
	writer io.Writer

	onNewLine   bool
	indentation int
	position    int

	maxPosition int

	useLiteralText bool
}

// NewCodeWriter returns a new code writer that writes to the provided
// writer.
func NewCodeWriter(w io.Writer) CodeWriter {
	return &standardCodeWriter{w, true, 0, 0, 80, false}
}

func (w *standardCodeWriter) FreshLine() {
	if !w.onNewLine {
		w.NewLine()
	}
}

func (w *standardCodeWriter) NewLine() {
	w.basicWrite([]byte("\n"))
}

func (w *standardCodeWriter) Indent() {
	if w.onNewLine {
		spaces := make([]byte, w.indentation)
		for i, _ := range spaces {
			spaces[i] = ' '
		}
		w.basicWrite(spaces)
	}
}

func (w *standardCodeWriter) ChangeIndentation(i int) {
	w.indentation += i
	if w.indentation < 0 {
		panic("CodeWriter indentation < 0")
	}
}

func (w *standardCodeWriter) LiteralText(literal bool) {
	w.useLiteralText = literal
}

func (w *standardCodeWriter) Write(b []byte) (n int, err error) {
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

func (w *standardCodeWriter) basicWrite(b []byte) (n int, err error) {
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
