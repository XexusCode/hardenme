// Package findings looks for things on a Mac that a person should
// recognise: programs that start by themselves and are not signed, remote
// access tools, configuration profiles, extra administrators, SSH keys and
// network redirections.
//
// These are signs, not verdicts. The list is short and conservative on
// purpose: a normal Mac should produce none. Anything that cannot be read
// is skipped, never reported.
package findings

import (
	"context"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/XexusCode/hardenme/internal/runner"
	"github.com/XexusCode/hardenme/internal/scan"
)

// Kinds of finding, in the order they are reported.
const (
	RemoteTool      = "remote_tool"
	StartupUnsigned = "startup_unsigned"
	StartupScript   = "startup_script"
	Profile         = "profile"
	Managed         = "managed"
	ExtraAdmin      = "extra_admin"
	SSHKeys         = "ssh_keys"
	Proxy           = "proxy"
	Hosts           = "hosts"
)

// Kinds lists every kind in report order.
var Kinds = []string{RemoteTool, StartupUnsigned, StartupScript, Profile, Managed, ExtraAdmin, SSHKeys, Proxy, Hosts}

// maxItems caps how many items one finding lists.
const maxItems = 8

// Env is where to look.
type Env struct {
	Runner runner.Runner
	Home   string // the user's home folder
	User   string // the user's short name
}

// Scan looks for every kind of finding on macOS.
func Scan(ctx context.Context, env Env) []scan.Finding {
	found := map[string][]string{}
	add := func(kind, item string) {
		if item != "" {
			found[kind] = append(found[kind], item)
		}
	}
	run := func(name string, args ...string) (string, bool) {
		out, err := env.Runner.Run(ctx, name, args...)
		if err != nil {
			return "", false
		}
		return out.Combined(), out.ExitCode == 0
	}

	// Remote access tools in the Applications folders.
	for _, dir := range []string{"/Applications", path.Join(env.Home, "Applications")} {
		if out, ok := run("/bin/ls", "-1", dir); ok {
			for _, app := range strings.Split(out, "\n") {
				if isRemoteTool(app) {
					add(RemoteTool, strings.TrimSuffix(strings.TrimSpace(app), ".app"))
				}
			}
		}
	}

	// Programs that start by themselves.
	for _, dir := range []string{path.Join(env.Home, "Library/LaunchAgents"), "/Library/LaunchAgents", "/Library/LaunchDaemons"} {
		out, ok := run("/bin/ls", "-1", dir)
		if !ok {
			continue
		}
		for _, name := range strings.Split(out, "\n") {
			name = strings.TrimSpace(name)
			if !strings.HasSuffix(name, ".plist") {
				continue
			}
			file := path.Join(dir, name)
			program, ok := run("/usr/bin/plutil", "-extract", "Program", "raw", file)
			if !ok {
				program, ok = run("/usr/bin/plutil", "-extract", "ProgramArguments.0", "raw", file)
			}
			program = strings.TrimSpace(program)
			if !ok || !strings.HasPrefix(program, "/") {
				continue
			}
			label := strings.TrimSuffix(name, ".plist")
			if interpreters[program] {
				add(StartupScript, label)
				continue
			}
			sig, _ := run("/usr/bin/codesign", "-dv", "--verbose=2", program)
			if unsigned(sig) {
				add(StartupUnsigned, label+" ("+tilde(program, env.Home)+")")
			}
		}
	}

	// Configuration profiles and remote management.
	if out, ok := run("/usr/bin/profiles", "list"); ok {
		for _, m := range profileID.FindAllStringSubmatch(out, -1) {
			add(Profile, m[1])
		}
	}
	if out, ok := run("/usr/bin/profiles", "status", "-type", "enrollment"); ok && mdmYes.MatchString(out) {
		add(Managed, "MDM")
	}

	// Administrators other than this user.
	if out, ok := run("/usr/bin/dscl", ".", "-read", "/Groups/admin", "GroupMembership"); ok {
		for _, name := range strings.Fields(strings.TrimPrefix(strings.TrimSpace(out), "GroupMembership:")) {
			if name != "root" && name != env.User && !strings.HasPrefix(name, "_") {
				add(ExtraAdmin, name)
			}
		}
	}

	// Keys that let someone sign in over SSH without the password.
	if out, ok := run("/bin/cat", path.Join(env.Home, ".ssh/authorized_keys")); ok {
		for _, line := range content(out) {
			fields := strings.Fields(line)
			name := "(" + fields[0] + ")"
			if len(fields) >= 3 {
				name = strings.Join(fields[2:], " ")
			}
			add(SSHKeys, name)
		}
	}

	// Traffic sent through a proxy.
	if out, ok := run("/usr/sbin/scutil", "--proxy"); ok {
		for _, m := range proxyOn.FindAllStringSubmatch(out, -1) {
			kind := strings.TrimSuffix(m[1], "Enable")
			host := ""
			if h := regexp.MustCompile(kind + `(?:Proxy|URLString) : (\S+)`).FindStringSubmatch(out); h != nil {
				host = " → " + h[1]
			}
			add(Proxy, kind+host)
		}
	}

	// Names redirected in the hosts file.
	if out, ok := run("/bin/cat", "/etc/hosts"); ok {
		for _, line := range content(out) {
			fields := strings.Fields(strings.SplitN(line, "#", 2)[0])
			for _, host := range fields[min(1, len(fields)):] {
				if host != "localhost" && host != "broadcasthost" && !strings.HasSuffix(host, ".localhost") && !strings.HasSuffix(host, ".local") {
					add(Hosts, host+" → "+fields[0])
				}
			}
		}
	}

	var out []scan.Finding
	for _, kind := range Kinds {
		items := found[kind]
		if len(items) == 0 {
			continue
		}
		sort.Strings(items)
		items = dedupe(items)
		if len(items) > maxItems {
			extra := len(items) - maxItems
			items = append(items[:maxItems:maxItems], "+"+strconv.Itoa(extra))
		}
		out = append(out, scan.Finding{Kind: kind, Items: items})
	}
	return out
}

// remoteTools are apps that let someone else see or control this Mac, or
// that are sold to watch what a person does. Matching is by app name.
var remoteTools = []string{
	"teamviewer", "anydesk", "rustdesk", "screenconnect", "connectwise", "logmein", "gotomypc",
	"splashtop", "vnc", "chrome remote desktop", "dwagent", "dwservice", "supremo", "ammyy",
	"remotepc", "zoho assist", "atera", "ultraviewer", "nomachine", "parsec",
	"mspy", "flexispy", "spyrix", "refog", "kidlogger", "hoverwatch", "spytech",
}

func isRemoteTool(app string) bool {
	name := strings.ToLower(strings.TrimSpace(app))
	if !strings.HasSuffix(name, ".app") {
		return false
	}
	for _, tool := range remoteTools {
		if strings.Contains(name, tool) {
			return true
		}
	}
	return false
}

// interpreters run a script, so their own signature says nothing about
// what starts.
var interpreters = map[string]bool{
	"/bin/sh": true, "/bin/bash": true, "/bin/zsh": true, "/bin/dash": true, "/bin/ksh": true,
	"/usr/bin/python3": true, "/usr/bin/perl": true, "/usr/bin/ruby": true,
	"/usr/bin/osascript": true, "/usr/bin/env": true, "/usr/bin/curl": true, "/usr/bin/nc": true,
}

var (
	// identified matches a signature by Apple or a registered developer.
	identified = regexp.MustCompile(`(?m)^Authority=(Developer ID Application|Apple |Software Signing|Apple Mac OS)`)
	profileID  = regexp.MustCompile(`(?m)profileIdentifier:\s*(\S+)`)
	mdmYes     = regexp.MustCompile(`(?m)^MDM enrollment:\s*Yes`)
	proxyOn    = regexp.MustCompile(`(?m)^\s*(HTTPEnable|HTTPSEnable|SOCKSEnable|ProxyAutoConfigEnable)\s*:\s*1\s*$`)
)

// unsigned reports whether codesign's output shows a program with no
// signature or a local (ad-hoc) one. Output it cannot classify (missing
// file, permission denied) is not treated as unsigned.
func unsigned(sig string) bool {
	if identified.MatchString(sig) {
		return false
	}
	return strings.Contains(sig, "not signed at all") || strings.Contains(sig, "Signature=adhoc")
}

// content returns the non-empty, non-comment lines.
func content(text string) []string {
	var out []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			out = append(out, line)
		}
	}
	return out
}

func tilde(p, home string) string {
	if home != "" && strings.HasPrefix(p, home+"/") {
		return "~" + strings.TrimPrefix(p, home)
	}
	return p
}

func dedupe(sorted []string) []string {
	out := sorted[:0]
	for i, s := range sorted {
		if i == 0 || s != sorted[i-1] {
			out = append(out, s)
		}
	}
	return out
}
