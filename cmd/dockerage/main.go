// Command dockerage is a compatibility shim. The implementation lives in
// internal/dockerage and is also reachable as "hall dockerage".
package main

import (
	"os"

	"github.com/QuietSugar/hall-kit/internal/dockerage"
)

func main() {
	dockerage.Main(os.Args[1:])
}
