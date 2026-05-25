package main

import (
	"github.com/browser-cli/internal/commands"
)

var version = "dev"

func main() {
	commands.Execute(version)
}