package check

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/XexusCode/hardenme/internal/runner"
)

// now is replaceable in tests.
var now = time.Now

const softwareUpdatePlist = "/Library/Preferences/com.apple.SoftwareUpdate"

func init() {
	Builtins["macos.screen_lock_delay"] = macosScreenLockDelay
	Builtins["macos.auto_lock_timeout"] = macosAutoLockTimeout
	Builtins["macos.find_my"] = macosFindMy
	Builtins["macos.auto_updates"] = macosAutoUpdates
	Builtins["macos.security_responses"] = macosSecurityResponses
	Builtins["macos.pending_updates"] = macosPendingUpdates
}

// run executes a command and appends its transcript to evidence.
func run(ctx context.Context, env Env, evidence *strings.Builder, name string, args ...string) (runner.Output, error) {
	out, err := env.Runner.Run(ctx, name, args...)
	evidence.WriteString(Evidence(name, args, out, err))
	return out, err
}

var delayRe = regexp.MustCompile(`screenLock delay is (\d+) seconds?`)

// macosScreenLockDelay: how long after sleep or the screen saver the Mac
// asks for the password. Adapted from fort's "screenlock" check, using
// sysadminctl because the old com.apple.screensaver keys stopped working
// after macOS 10.13.
func macosScreenLockDelay(ctx context.Context, env Env) Result {
	var ev strings.Builder
	out, err := run(ctx, env, &ev, "sysadminctl", "-screenLock", "status")
	if err != nil {
		return Result{Status: Unknown, Current: err.Error(), Evidence: ev.String()}
	}
	text := out.Combined()
	switch {
	case strings.Contains(text, "screenLock is immediate"), strings.Contains(text, "screenLock delay is immediate"):
		return Result{Status: Pass, Current: "immediately", Evidence: ev.String()}
	case strings.Contains(text, "screenLock is off"):
		return Result{Status: Fail, Current: "off", Evidence: ev.String()}
	}
	if m := delayRe.FindStringSubmatch(text); m != nil {
		secs, _ := strconv.Atoi(m[1])
		current := fmt.Sprintf("%d s", secs)
		switch {
		case secs <= 5:
			return Result{Status: Pass, Current: current, Evidence: ev.String()}
		case secs <= 60:
			return Result{Status: Warn, Current: current, Evidence: ev.String()}
		default:
			return Result{Status: Fail, Current: current, Evidence: ev.String()}
		}
	}
	return Result{Status: Unknown, Current: firstLine(text), Evidence: ev.String()}
}

var displaySleepRe = regexp.MustCompile(`(?m)^\s*displaysleep\s+(\d+)`)

// macosAutoLockTimeout: minutes of inactivity before the screen turns off or
// the screen saver starts, whichever comes first. The limit is stricter for
// devices that leave home or are shared.
func macosAutoLockTimeout(ctx context.Context, env Env) Result {
	var ev strings.Builder
	minutes := -1 // -1 = never

	if out, err := run(ctx, env, &ev, "pmset", "-g"); err == nil && out.ExitCode == 0 {
		if m := displaySleepRe.FindStringSubmatch(out.Stdout); m != nil {
			if v, _ := strconv.Atoi(m[1]); v > 0 {
				minutes = v
			}
		}
	} else {
		return Result{Status: Unknown, Current: "pmset unavailable", Evidence: ev.String()}
	}
	if out, err := run(ctx, env, &ev, "defaults", "-currentHost", "read", "com.apple.screensaver", "idleTime"); err == nil && out.ExitCode == 0 {
		if secs, convErr := strconv.Atoi(strings.TrimSpace(out.Stdout)); convErr == nil && secs > 0 {
			saver := (secs + 59) / 60
			if minutes < 0 || saver < minutes {
				minutes = saver
			}
		}
	}

	if minutes < 0 {
		return Result{Status: Fail, Current: "never", Evidence: ev.String()}
	}
	passMax, warnMax := 15, 30
	p := env.Profile
	if p.Portable || p.LeavesHome || p.Shared {
		passMax, warnMax = 5, 15
	}
	current := fmt.Sprintf("%d min", minutes)
	switch {
	case minutes <= passMax:
		return Result{Status: Pass, Current: current, Evidence: ev.String()}
	case minutes <= warnMax:
		return Result{Status: Warn, Current: current, Evidence: ev.String()}
	default:
		return Result{Status: Fail, Current: current, Evidence: ev.String()}
	}
}

// macosFindMy: the Find My Mac token in NVRAM proves it is on. Its absence
// does not prove it is off on every Mac, so we then ask the user to look.
func macosFindMy(ctx context.Context, env Env) Result {
	var ev strings.Builder
	out, err := run(ctx, env, &ev, "nvram", "fmm-mobileme-token-FMM")
	if err == nil && out.ExitCode == 0 && strings.Contains(out.Stdout, "fmm-mobileme-token-FMM") {
		// Never keep the token itself in the evidence.
		return Result{Status: Pass, Current: "on", Evidence: "$ nvram fmm-mobileme-token-FMM\nfmm-mobileme-token-FMM <present>\n"}
	}
	return Result{Status: Manual, Current: "could not confirm", Evidence: ev.String()}
}

// readBoolKey reads a 0/1 key from a defaults domain. missing reports that
// the key does not exist (macOS then uses its default).
func readBoolKey(ctx context.Context, env Env, ev *strings.Builder, domain, key string) (value, missing bool, err error) {
	out, err := run(ctx, env, ev, "defaults", "read", domain, key)
	if err != nil {
		return false, false, err
	}
	if out.ExitCode != 0 {
		if strings.Contains(out.Combined(), "does not exist") {
			return false, true, nil
		}
		return false, false, fmt.Errorf("defaults read %s %s: exit %d", domain, key, out.ExitCode)
	}
	return strings.TrimSpace(out.Stdout) == "1", false, nil
}

// boolKey is a SoftwareUpdate key and the value macOS uses when it is missing.
type boolKey struct {
	name        string
	missingIsOn bool
}

// checkKeys passes when every key is on, warns when only some are.
func checkKeys(ctx context.Context, env Env, keys []boolKey) Result {
	var ev strings.Builder
	var off []string
	for _, key := range keys {
		on, missing, err := readBoolKey(ctx, env, &ev, softwareUpdatePlist, key.name)
		if err != nil {
			return Result{Status: Unknown, Current: err.Error(), Evidence: ev.String()}
		}
		if missing {
			on = key.missingIsOn
		}
		if !on {
			off = append(off, key.name)
		}
	}
	switch {
	case len(off) == 0:
		return Result{Status: Pass, Current: "all on", Evidence: ev.String()}
	case len(off) < len(keys):
		return Result{Status: Warn, Current: "off: " + strings.Join(off, ", "), Evidence: ev.String()}
	default:
		return Result{Status: Fail, Current: "off: " + strings.Join(off, ", "), Evidence: ev.String()}
	}
}

// macosAutoUpdates: check, download and install macOS updates automatically.
// Checking and downloading are on by default; installing is off by default.
func macosAutoUpdates(ctx context.Context, env Env) Result {
	return checkKeys(ctx, env, []boolKey{
		{"AutomaticCheckEnabled", true},
		{"AutomaticDownload", true},
		{"AutomaticallyInstallMacOSUpdates", false},
	})
}

// macosSecurityResponses: XProtect/security data files and Rapid Security
// Responses. Both are on by default.
func macosSecurityResponses(ctx context.Context, env Env) Result {
	return checkKeys(ctx, env, []boolKey{
		{"ConfigDataInstall", true},
		{"CriticalUpdateInstall", true},
	})
}

var (
	macosUpdateRe = regexp.MustCompile(`"Display Name" = "macOS (?:[A-Za-z]+ )?(\d+)[^"]*"`)
	plistDateRe   = regexp.MustCompile(`\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2} [+-]\d{4}`)
)

// macosPendingUpdates fails when an update for the installed macOS version is
// waiting, and warns when the Mac has not checked for updates in 30 days.
// Upgrades to a new major version are reported but do not fail the check.
func macosPendingUpdates(ctx context.Context, env Env) Result {
	var ev strings.Builder
	ver, err := run(ctx, env, &ev, "sw_vers", "-productVersion")
	if err != nil || ver.ExitCode != 0 {
		return Result{Status: Unknown, Current: "could not read the macOS version", Evidence: ev.String()}
	}
	version := strings.TrimSpace(ver.Stdout)
	major := strings.SplitN(version, ".", 2)[0]

	recs, err := run(ctx, env, &ev, "defaults", "read", softwareUpdatePlist, "RecommendedUpdates")
	if err != nil {
		return Result{Status: Unknown, Current: err.Error(), Evidence: ev.String()}
	}
	var sameMajor, newMajor []string
	if recs.ExitCode == 0 {
		for _, m := range macosUpdateRe.FindAllStringSubmatch(recs.Stdout, -1) {
			name := strings.TrimPrefix(strings.TrimSuffix(m[0], `"`), `"Display Name" = "`)
			if m[1] == major {
				sameMajor = append(sameMajor, name)
			} else {
				newMajor = append(newMajor, name)
			}
		}
	}
	if len(sameMajor) > 0 {
		return Result{Status: Fail, Current: version + ", pending: " + strings.Join(sameMajor, ", "), Evidence: ev.String()}
	}

	last, err := run(ctx, env, &ev, "defaults", "read", softwareUpdatePlist, "LastFullSuccessfulDate")
	current := version + ", up to date"
	if len(newMajor) > 0 {
		current = version + ", upgrade available: " + strings.Join(newMajor, ", ")
	}
	if err != nil || last.ExitCode != 0 {
		return Result{Status: Warn, Current: current + "; last check unknown", Evidence: ev.String()}
	}
	when, perr := time.Parse("2006-01-02 15:04:05 -0700", plistDateRe.FindString(last.Stdout))
	if perr != nil {
		return Result{Status: Warn, Current: current + "; last check unknown", Evidence: ev.String()}
	}
	if days := int(now().Sub(when).Hours() / 24); days > 30 {
		return Result{Status: Warn, Current: fmt.Sprintf("%s; last check %d days ago", current, days), Evidence: ev.String()}
	}
	return Result{Status: Pass, Current: current, Evidence: ev.String()}
}
