package broker

import (
	"fmt"
	"os"
)

// VERSION is the version string for built artifacts. It's set by the build system, and should
// not be changed in this codebase
var VERSION = "UNKNOWN"

// PrintAndExit will print the version and exit.
func PrintAndExit() {
	fmt.Println(VERSION)
	os.Exit(0)
}
