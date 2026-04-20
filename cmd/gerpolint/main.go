// Binary gerpolint is a standalone static analyzer that checks type
// compatibility between gerpo field pointers and filter arguments. Typical
// CI wiring: `gerpolint ./...`.
package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/insei/gerpo/internal/gerpolint"
)

func main() {
	singlechecker.Main(gerpolint.Analyzer)
}
