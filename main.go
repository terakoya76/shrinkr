package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/terakoya76/shrinkr/cmd"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	kctx, cli, err := cmd.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	_ = cli
	if err := kctx.Run(cmd.Context{Ctx: ctx}); err != nil {
		fmt.Fprintln(os.Stderr, "shrinkr:", err)
		os.Exit(1)
	}
}
