// Command rmempty is a compatibility shim. The implementation lives in
// internal/rmempty and is also reachable as "hall rmempty".
package main

import (
	"os"

	"github.com/QuietSugar/hall-kit/internal/rmempty"
)

func main() {
	rmempty.Main(os.Args[1:])
}
