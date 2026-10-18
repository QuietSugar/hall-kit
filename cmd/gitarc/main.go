// Command gitarc is a compatibility shim. The implementation lives in
// internal/gitarc and is also reachable as "hall gitarc".
package main

import (
	"os"

	"github.com/QuietSugar/hall-kit/internal/gitarc"
)

func main() {
	gitarc.Main(os.Args[1:])
}
