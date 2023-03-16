// ml_i - an ML/I macro processor ported to Go
// Copyright (c) 2023 Michael D Henderson.
// All rights reserved.

package lowl

import (
	"io"
)

type OpAction func(w io.Writer, cmd *command) error

func initOpTable() map[string]OpAction {
	tbl := make(map[string]OpAction)
	tbl["NOOP"] = func(w io.Writer, cmd *command) error {
		return printf(w, "// %8d: NOOP\n", cmd.line)
	}
	tbl["NB"] = func(w io.Writer, cmd *command) error {
		_ = printf(w, "// %8d:", cmd.line)
		for _, arg := range cmd.args {
			_ = printf(w, " %s", arg.String())
		}
		return printf(w, "\n")
	}

	return tbl
}
