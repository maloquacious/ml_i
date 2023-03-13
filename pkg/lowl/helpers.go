package lowl

import (
	"fmt"
	"io"
)

func fprintf(w io.Writer, format string, args ...any) {
	_, _ = w.Write([]byte(fmt.Sprintf(format, args...)))
}

func printf(w io.Writer, format string, args ...any) error {
	_, err := w.Write([]byte(fmt.Sprintf(format, args...)))
	return err
}
