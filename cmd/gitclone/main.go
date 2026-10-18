// Command gitclone is a compatibility shim. The implementation lives in
// internal/gitclone and is also reachable as "hall gitclone".
package main

import (
	"os"

	"github.com/QuietSugar/hall-kit/internal/gitclone"
)

func main() {
	gitclone.Main(os.Args[1:])
}
