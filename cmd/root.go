// Package cmd wires the CLI. The kong tags on the CLI struct describe every
// subcommand and flag; kong parses argv into it and invokes Run() on the
// selected subcommand.
package cmd

import (
	"context"

	"github.com/alecthomas/kong"
)

type CLI struct {
	Run     RunCmd     `cmd:"" help:"Compress an album of photos and videos."`
	Doctor  DoctorCmd  `cmd:"" help:"Check external dependencies."`
	Presets PresetsCmd `cmd:"" help:"List available compression presets."`
}

type Context struct {
	Ctx context.Context
}

func Parse(argv []string) (*kong.Context, *CLI, error) {
	var cli CLI
	parser, err := kong.New(&cli,
		kong.Name("shrinkr"),
		kong.Description("Compress Google Photos Takeout albums so they can be re-uploaded to reclaim quota."),
		kong.UsageOnError(),
	)
	if err != nil {
		return nil, nil, err
	}
	kctx, err := parser.Parse(argv)
	if err != nil {
		return nil, nil, err
	}
	return kctx, &cli, nil
}
