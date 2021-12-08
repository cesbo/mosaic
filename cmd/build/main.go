package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

// Build
func build() {
	args := []string{
		"build",
		"-ldflags", "-s -w",
		"mosaic/cmd/mosaic",
	}

	cmd := exec.Command("go", args...)
	cmd.Env = make([]string, 0, 8)

	if v := os.Getenv("PATH"); len(v) != 0 {
		cmd.Env = append(cmd.Env, fmt.Sprintf("PATH=%s", v))
	}

	if v := os.Getenv("HOME"); len(v) != 0 {
		cmd.Env = append(cmd.Env, fmt.Sprintf("HOME=%s", v))
	}

	if v := os.Getenv("SHELL"); len(v) != 0 {
		cmd.Env = append(cmd.Env, fmt.Sprintf("SHELL=%s", v))
	}

	if v := os.Getenv("USER"); len(v) != 0 {
		cmd.Env = append(cmd.Env, fmt.Sprintf("USER=%s", v))
	}

	if v := os.Getenv("TARGET"); len(v) != 0 {
		target := strings.Split(v, "/")
		if len(target) == 2 {
			cmd.Env = append(cmd.Env, fmt.Sprintf("GOARCH=%s", target[0]))
			cmd.Env = append(cmd.Env, fmt.Sprintf("GOOS=%s", target[1]))
		}
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		log.Fatalf("build failed:\n\n%s\n\n", stderr.String())
	}
}

// Restore generated files
func restore() {
	args := []string{
		"restore",
		VERSION_GEN,
	}

	cmd := exec.Command("git", args...)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		log.Fatalf("restore generated files failed:\n\n%s\n\n", stderr.String())
	}
}

func main() {
	updateVersion()
	build()
	restore()
}
