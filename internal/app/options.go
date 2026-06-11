package app

import (
	"flag"
	"fmt"
	"io"
	"strings"

	cfgpkg "github.com/oddship/wg-tui/internal/config"
)

type Options struct {
	ConfigPath string
	CachePath  string
	ShowHelp   bool
}

func ParseOptions(args []string) (Options, error) {
	opts := Options{ConfigPath: cfgpkg.ConfigPath()}
	fs := flag.NewFlagSet("wgt", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&opts.ConfigPath, "config", opts.ConfigPath, "path to config.huml")
	fs.StringVar(&opts.CachePath, "cache-path", "", "override cache/state dir for this run")
	fs.BoolVar(&opts.ShowHelp, "help", false, "show help")
	fs.BoolVar(&opts.ShowHelp, "h", false, "show help")
	if err := fs.Parse(args); err != nil {
		return Options{}, err
	}
	if opts.ShowHelp {
		return opts, nil
	}
	if fs.NArg() > 0 {
		return Options{}, fmt.Errorf("unexpected arguments: %s", strings.Join(fs.Args(), " "))
	}
	return opts, nil
}

func Usage() string {
	return fmt.Sprintf(`Usage: wgt [options]

Options:
  --config <path>      Path to config.huml (default: %s)
  --cache-path <dir>   Override cache/state dir for this run
  --help, -h           Show help
`, cfgpkg.ConfigPath())
}
