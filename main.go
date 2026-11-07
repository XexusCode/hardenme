// Command hardenme audits a device's security for a normal user and explains
// how to improve it.
package main

import (
	"os"

	"github.com/XexusCode/hardenme/internal/cli"
)

// version is set at build time by GoReleaser (-X main.version=...).
var version = "dev"

func main() {
	os.Exit(cli.Execute(version))
}
