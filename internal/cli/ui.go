package cli

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"runtime"
	"time"

	"github.com/spf13/cobra"

	"github.com/XexusCode/hardenme/checks"
	"github.com/XexusCode/hardenme/internal/check"
	"github.com/XexusCode/hardenme/internal/fix"
	"github.com/XexusCode/hardenme/internal/profile"
	"github.com/XexusCode/hardenme/internal/runner"
	"github.com/XexusCode/hardenme/internal/story"
	"github.com/XexusCode/hardenme/internal/ui"
)

// uiIdle is how long the server waits without hearing from the page (which
// pings every 30 seconds) before it closes itself. Browsers slow hidden
// tabs down to about one timer a minute, so this leaves room.
const uiIdle = 5 * time.Minute

func newUI(opts *options) *cobra.Command {
	var noOpen bool
	cmd := &cobra.Command{
		Use:   "ui",
		Short: "Open hardenme in your browser: see your score and fix things in a couple of clicks",
		RunE: func(cmd *cobra.Command, _ []string) error {
			lang, err := check.ParseLang(opts.lang)
			if err != nil {
				return err
			}
			platform, err := currentPlatform()
			if err != nil {
				return err
			}
			all, err := check.Load(checks.FS, platform)
			if err != nil {
				return err
			}
			stories, err := story.Load(checks.FS, platform)
			if err != nil {
				return err
			}
			store, err := profile.DefaultStore()
			if err != nil {
				return err
			}
			ctx := cmd.Context()
			probe := runner.Exec{}
			portable, _ := profile.DetectPortable(ctx, probe, runtime.GOOS)

			// Loopback only, on a port the system picks.
			ln, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				return err
			}
			secret := make([]byte, 24)
			if _, err := rand.Read(secret); err != nil {
				return err
			}
			token := hex.EncodeToString(secret)
			host := ln.Addr().String()
			srv := ui.New(ui.Deps{
				Platform: platform, OSVersion: osVersion(ctx, probe), Lang: lang,
				Catalog: all, Stories: stories,
				Probe: probe, Apply: runner.Exec{Timeout: 5 * time.Minute},
				Log: fix.DefaultLog, Profiles: store, Fallback: profile.Default(portable),
				Now: time.Now,
			}, host, token)
			httpSrv := &http.Server{Handler: srv.Handler(), ReadHeaderTimeout: 10 * time.Second}
			errc := make(chan error, 1)
			go func() { errc <- httpSrv.Serve(ln) }()

			url := "http://" + host + "/?t=" + token
			fmt.Fprintln(cmd.OutOrStdout(), t(lang, "ui_running"))
			if noOpen {
				fmt.Fprintln(cmd.OutOrStdout(), url)
			} else if _, err := probe.Run(ctx, "/usr/bin/open", url); err != nil {
				fmt.Fprintln(cmd.OutOrStdout(), url)
			}

			tick := time.NewTicker(15 * time.Second)
			defer tick.Stop()
			for {
				select {
				case err := <-errc:
					if errors.Is(err, http.ErrServerClosed) {
						return nil
					}
					return err
				case <-srv.Done():
				case <-ctx.Done():
				case <-tick.C:
					if srv.Idle() < uiIdle {
						continue
					}
				}
				shutdown, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				return httpSrv.Shutdown(shutdown)
			}
		},
	}
	cmd.Flags().BoolVar(&noOpen, "no-open", false, "print the address instead of opening the browser")
	return cmd
}
