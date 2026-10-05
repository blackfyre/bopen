// Command argv is a stub browser for tests: it writes its arguments as JSON
// to the file named by BOPEN_ARGV_OUT.
package main

import (
	"encoding/json"
	"os"
)

func main() {
	data, _ := json.Marshal(os.Args[1:])
	out := os.Getenv("BOPEN_ARGV_OUT")
	_ = os.WriteFile(out+".tmp", data, 0o644)
	_ = os.Rename(out+".tmp", out)
}
