package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/francogalfre/coop/packages/cli/internal/config"
	"github.com/francogalfre/coop/packages/cli/internal/harness"
	"github.com/francogalfre/coop/packages/cli/internal/hooks"
	"github.com/francogalfre/coop/packages/cli/internal/ptywrap"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "coop: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) < 1 {
		return usageError()
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	switch args[0] {
	case "attach":
		return runAttach(ctx, cfg, args[1:])
	case "run":
		return runWrapped(ctx, cfg, args[1:])
	case "detach":
		return runDetach(args[1:])
	case "login":
		return runLogin(ctx, cfg)
	default:
		return usageError()
	}
}

func usageError() error {
	return fmt.Errorf("usage: coop attach [--harness=<name>] | coop run [--harness=<name>] -- <cmd> [args...] | coop detach [dir] | coop login")
}

func parseHarnessFlag(fsName string, args []string) (string, []string, error) {
	fs := flag.NewFlagSet(fsName, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	name := fs.String("harness", "", "install hooks for only this harness (claude-code, opencode, pi)")

	if err := fs.Parse(args); err != nil {
		return "", nil, err
	}

	return *name, fs.Args(), nil
}

func runAttach(ctx context.Context, cfg config.Config, args []string) error {
	harnessFlag, _, err := parseHarnessFlag("attach", args)
	if err != nil {
		return err
	}

	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getwd: %w", err)
	}

	adapters, err := selectHarnesses(dir, harnessFlag)
	if err != nil {
		return err
	}

	installed, err := hooks.Install(cfg, true)
	if err != nil {
		return err
	}

	installations, err := installHarnesses(dir, installed.BaseURL, adapters)
	if err != nil {
		_ = installed.Stop()
		return err
	}

	fmt.Printf("coop attach: session %s\n", cfg.SessionID)
	fmt.Printf("coop attach: relay   %s\n", cfg.RelayURL)
	fmt.Printf("coop attach: listening on %s\n", installed.BaseURL)
	fmt.Printf("coop attach: share    http://localhost:3000/sessions/%s?token=%s\n", cfg.SessionID, cfg.SessionToken)
	fmt.Println("coop attach: start your agent session now. Press Ctrl-C to stop.")

	<-ctx.Done()

	removeInstallations(installations)

	return installed.Stop()
}

func runWrapped(ctx context.Context, cfg config.Config, args []string) error {
	harnessFlag, remaining, err := parseHarnessFlag("run", args)
	if err != nil {
		return err
	}

	name, cmdArgs := parseRunArgs(remaining)

	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getwd: %w", err)
	}

	adapters, err := selectRunHarness(harnessFlag, name)
	if err != nil {
		return err
	}

	installed, err := hooks.Install(cfg, false)
	if err != nil {
		return err
	}
	defer installed.Stop()

	installations, err := installHarnesses(dir, installed.BaseURL, adapters)
	if err != nil {
		return err
	}
	defer removeInstallations(installations)

	return ptywrap.Run(ctx, cfg, name, cmdArgs)
}

func runDetach(args []string) error {
	dir := ""
	if len(args) > 0 {
		dir = args[0]
	}

	if dir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getwd: %w", err)
		}
		dir = wd
	}

	if err := harness.RemoveAllTraces(dir, allAdapters); err != nil {
		return err
	}

	fmt.Printf("coop detach: removed coop harness entries from %s\n", dir)

	return nil
}

func parseRunArgs(args []string) (string, []string) {
	if len(args) > 0 && args[0] == "--" {
		args = args[1:]
	}

	if len(args) == 0 {
		return "claude", nil
	}

	return args[0], args[1:]
}
