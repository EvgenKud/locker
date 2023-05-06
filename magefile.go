//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/magefile/mage/mg"
)

func execCommand(cmd *exec.Cmd) error {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

type Test mg.Namespace

func (Test) Tests() error {
	fmt.Println("start all tests...")
	mg.SerialDeps(mg.F(Test.Vet), mg.F(Test.GolangciLint), mg.F(Test.GoTests))
	return nil
}

func (Test) Vet() error {
	fmt.Println("Start vet...")
	return execCommand(
		exec.Command("go", "vet", "./..."),
	)
}

func (Test) GolangciLint() error {
	fmt.Println("Start golangci-lint...")
	return execCommand(
		exec.Command("golangci-lint", "run"),
	)
}

func (Test) GoTests() error {
	fmt.Println("Start tests...")
	return execCommand(
		exec.Command("go", "test"),
	)
}
