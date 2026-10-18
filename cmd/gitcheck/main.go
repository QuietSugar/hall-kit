// Command gitcheck is a compatibility shim. The implementation lives in
// internal/gitcheck and is also reachable as "hall gitcheck".
package main

import (
	"os"

	"github.com/QuietSugar/hall-kit/internal/gitcheck"
)

func main() {
	gitcheck.Main(os.Args[1:])
}
