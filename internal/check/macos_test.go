package check_test

import (
	"strings"
	"testing"
	"time"

	"github.com/XexusCode/hardenme/internal/check"
	"github.com/XexusCode/hardenme/internal/profile"
	"github.com/XexusCode/hardenme/internal/runner"
)

// Outputs marked "26.2" were captured on a real Mac running macOS 26.2
// (25C56). Others reproduce formats documented for older macOS versions.

const (
	socketfilterfw = "/usr/libexec/ApplicationFirewall/socketfilterfw"
	loginwindow    = "/Library/Preferences/com.apple.loginwindow"
	softwareUpdate = "/Library/Preferences/com.apple.SoftwareUpdate"
	missingKey     = "2026-11-06 15:47:44.579 defaults[99260:21639184] \nThe domain/default pair of (x, y) does not exist"
)

var laptop = profile.Profile{Portable: true, LeavesHome: true, Risk: profile.RiskNormal}
var homeDesktop = profile.Profile{Risk: profile.RiskNormal}

type macCase struct {
	name    string
	outputs map[string]runner.Output
	want    check.Status
	current string
}

func runCases(t *testing.T, id string, p profile.Profile, cases []macCase) {
	t.Helper()
	c := macCheck(t, id)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := runWith(t, c, p, tc.outputs)
			if got.Status != tc.want {
				t.Fatalf("status = %s, want %s (current %q)\nevidence:\n%s", got.Status, tc.want, got.Current, got.Evidence)
			}
			if tc.current != "" && got.Current != tc.current {
				t.Errorf("current = %q, want %q", got.Current, tc.current)
			}
			if got.Evidence == "" {
				t.Error("evidence should never be empty")
			}
		})
	}
}

// notFound is an empty output map: every command behaves as missing.
var notFound = map[string]runner.Output{}

func TestMacOSFileVault(t *testing.T) {
	k := runner.Key("fdesetup", "status")
	runCases(t, "macos.filevault", laptop, []macCase{
		{"on", map[string]runner.Output{k: out("FileVault is On.\n", 0)}, check.Pass, "FileVault is On."},
		{"encrypting", map[string]runner.Output{k: out("FileVault is On.\nEncryption in progress: Percent completed = 12.3\n", 0)}, check.Pass, ""},
		{"off 26.2", map[string]runner.Output{k: out("FileVault is Off.\n", 0)}, check.Fail, "FileVault is Off."},
		{"decrypting", map[string]runner.Output{k: out("FileVault is Off.\nDecryption in progress: Percent completed = 40\n", 0)}, check.Fail, ""},
		{"unexpected", map[string]runner.Output{k: out("something else\n", 0)}, check.Unknown, ""},
		{"missing tool", notFound, check.Unknown, ""},
	})
}

func TestMacOSFirewall(t *testing.T) {
	k := runner.Key(socketfilterfw, "--getglobalstate")
	runCases(t, "macos.firewall", laptop, []macCase{
		{"enabled", map[string]runner.Output{k: out("Firewall is enabled. (State = 1)\n", 0)}, check.Pass, ""},
		{"block all", map[string]runner.Output{k: out("Firewall is enabled. (State = 2)\n", 0)}, check.Pass, ""},
		{"disabled 26.2", map[string]runner.Output{k: out("Firewall is disabled. (State = 0)\n", 0)}, check.Fail, "Firewall is disabled. (State = 0)"},
		{"missing tool", notFound, check.Unknown, ""},
	})
}

func TestMacOSFirewallStealth(t *testing.T) {
	k := runner.Key(socketfilterfw, "--getstealthmode")
	runCases(t, "macos.firewall_stealth", laptop, []macCase{
		{"on", map[string]runner.Output{k: out("Firewall stealth mode is on\n", 0)}, check.Pass, ""},
		{"off 26.2", map[string]runner.Output{k: out("Firewall stealth mode is off\n", 0)}, check.Fail, ""},
		{"older wording", map[string]runner.Output{k: out("Stealth mode enabled\n", 0)}, check.Pass, ""},
	})
}

func TestMacOSGatekeeperAndSIP(t *testing.T) {
	runCases(t, "macos.gatekeeper", laptop, []macCase{
		{"enabled 26.2", map[string]runner.Output{runner.Key("spctl", "--status"): out("assessments enabled\n", 0)}, check.Pass, ""},
		{"disabled", map[string]runner.Output{runner.Key("spctl", "--status"): out("assessments disabled\n", 1)}, check.Fail, ""},
	})
	k := runner.Key("csrutil", "status")
	runCases(t, "macos.sip", laptop, []macCase{
		{"enabled 26.2", map[string]runner.Output{k: out("System Integrity Protection status: enabled.\n", 0)}, check.Pass, ""},
		{"disabled", map[string]runner.Output{k: out("System Integrity Protection status: disabled.\n", 0)}, check.Fail, ""},
		{"custom", map[string]runner.Output{k: out("System Integrity Protection status: unknown (Custom Configuration).\n", 0)}, check.Warn, ""},
	})
}

func TestMacOSLoginWindow(t *testing.T) {
	auto := runner.Key("defaults", "read", loginwindow, "autoLoginUser")
	runCases(t, "macos.auto_login", laptop, []macCase{
		{"off 26.2", map[string]runner.Output{auto: errOut(missingKey, 1)}, check.Pass, ""},
		{"on", map[string]runner.Output{auto: out("alex\n", 0)}, check.Fail, ""},
		{"other error", map[string]runner.Output{auto: errOut("permission denied", 1)}, check.Unknown, ""},
	})
	guest := runner.Key("defaults", "read", loginwindow, "GuestEnabled")
	runCases(t, "macos.guest_account", laptop, []macCase{
		{"off 26.2", map[string]runner.Output{guest: out("0\n", 0)}, check.Pass, ""},
		{"never set", map[string]runner.Output{guest: errOut(missingKey, 1)}, check.Pass, ""},
		{"on", map[string]runner.Output{guest: out("1\n", 0)}, check.Fail, ""},
	})
}

func TestMacOSRemoteLogin(t *testing.T) {
	k := runner.Key("launchctl", "print-disabled", "system")
	runCases(t, "macos.remote_login", laptop, []macCase{
		{"enabled 26.2", map[string]runner.Output{k: out("\n\tdisabled services = {\n\t\t\"com.apple.ftpd\" => disabled\n\t\t\"com.openssh.sshd\" => enabled\n\t}\n", 0)}, check.Fail, `"com.openssh.sshd" => enabled`},
		{"disabled", map[string]runner.Output{k: out("\t\t\"com.openssh.sshd\" => disabled\n", 0)}, check.Pass, ""},
		{"older format enabled", map[string]runner.Output{k: out("\t\t\"com.openssh.sshd\" => false\n", 0)}, check.Fail, ""},
		{"older format disabled", map[string]runner.Output{k: out("\t\t\"com.openssh.sshd\" => true\n", 0)}, check.Pass, ""},
		{"never toggled", map[string]runner.Output{k: out("\tdisabled services = {\n\t\t\"com.apple.ftpd\" => disabled\n\t}\n", 0)}, check.Pass, ""},
	})
}

func TestMacOSSharingServices(t *testing.T) {
	ss := runner.Key("launchctl", "print", "system/com.apple.screensharing")
	runCases(t, "macos.screen_sharing", laptop, []macCase{
		{"off 26.2", map[string]runner.Output{ss: errOut("Could not find service \"com.apple.screensharing\" in domain for system", 113)}, check.Pass, ""},
		{"on", map[string]runner.Output{ss: out("system/com.apple.screensharing = {\n\tactive count = 0\n", 0)}, check.Fail, ""},
		{"other error", map[string]runner.Output{ss: errOut("boom", 5)}, check.Unknown, ""},
	})
	smb := runner.Key("launchctl", "print", "system/com.apple.smbd")
	runCases(t, "macos.file_sharing", laptop, []macCase{
		{"off 26.2", map[string]runner.Output{smb: errOut("Could not find service", 113)}, check.Pass, ""},
		{"on", map[string]runner.Output{smb: out("system/com.apple.smbd = {", 0)}, check.Fail, ""},
	})
	ard := runner.Key("pgrep", "-x", "ARDAgent")
	runCases(t, "macos.remote_management", laptop, []macCase{
		{"off 26.2", map[string]runner.Output{ard: out("", 1)}, check.Pass, ""},
		{"on", map[string]runner.Output{ard: out("512\n", 0)}, check.Fail, ""},
	})
}

func TestMacOSAirDropAdminLockdown(t *testing.T) {
	ad := runner.Key("defaults", "read", "com.apple.sharingd", "DiscoverableMode")
	runCases(t, "macos.airdrop", laptop, []macCase{
		{"default 26.2", map[string]runner.Output{ad: errOut(missingKey, 1)}, check.Pass, ""},
		{"contacts", map[string]runner.Output{ad: out("Contacts Only\n", 0)}, check.Pass, ""},
		{"everyone", map[string]runner.Output{ad: out("Everyone\n", 0)}, check.Fail, ""},
	})
	id := runner.Key("id", "-Gn")
	runCases(t, "macos.admin_account", laptop, []macCase{
		{"admin 26.2", map[string]runner.Output{id: out("staff everyone localaccounts _appserverusr admin _appserveradm com.apple.sharepoint.group.1 _appstore _lpadmin _lpoperator _developer _analyticsusers com.apple.access_ftp com.apple.access_screensharing com.apple.access_ssh\n", 0)}, check.Warn, "admin"},
		{"standard", map[string]runner.Output{id: out("staff everyone localaccounts _lpadmin\n", 0)}, check.Pass, ""},
	})
	ld := runner.Key("sysctl", "-n", "security.mac.lockdown_mode_state")
	elevated := profile.Profile{Risk: profile.RiskElevated}
	runCases(t, "macos.lockdown_mode", elevated, []macCase{
		{"off 26.2", map[string]runner.Output{ld: out("0\n", 0)}, check.Fail, ""},
		{"on", map[string]runner.Output{ld: out("1\n", 0)}, check.Pass, ""},
		{"older macOS", map[string]runner.Output{ld: errOut("sysctl: unknown oid 'security.mac.lockdown_mode_state'", 1)}, check.Unknown, ""},
	})
}

func TestMacOSScreenLockDelay(t *testing.T) {
	k := runner.Key("sysadminctl", "-screenLock", "status")
	line := func(s string) map[string]runner.Output {
		return map[string]runner.Output{k: errOut("2026-11-06 15:47:44.510 sysadminctl[99242:21639140] "+s+"\n", 0)}
	}
	runCases(t, "macos.screen_lock_delay", laptop, []macCase{
		{"off 26.2", line("screenLock is off"), check.Fail, "off"},
		{"immediate", line("screenLock delay is immediate"), check.Pass, "immediately"},
		{"5 seconds", line("screenLock delay is 5 seconds"), check.Pass, "5 s"},
		{"1 minute", line("screenLock delay is 60 seconds"), check.Warn, "60 s"},
		{"1 hour", line("screenLock delay is 3600 seconds"), check.Fail, "3600 s"},
		{"unexpected", line("something new"), check.Unknown, ""},
		{"missing tool", notFound, check.Unknown, ""},
	})
}

func TestMacOSAutoLockDependsOnProfile(t *testing.T) {
	pm := runner.Key("pmset", "-g")
	saver := runner.Key("defaults", "-currentHost", "read", "com.apple.screensaver", "idleTime")
	pmset := func(display string) runner.Output {
		return out("System-wide power settings:\nCurrently in use:\n standby              1\n disksleep            10\n displaysleep         "+display+"\n hibernatemode        3\n", 0)
	}
	ten := map[string]runner.Output{pm: pmset("10"), saver: errOut(missingKey, 1)}

	runCases(t, "macos.auto_lock", laptop, []macCase{
		{"120 min 26.2", map[string]runner.Output{pm: pmset("120"), saver: errOut(missingKey, 1)}, check.Fail, "120 min"},
		{"10 min on a laptop", ten, check.Warn, "10 min"},
		{"screen saver sooner", map[string]runner.Output{pm: pmset("120"), saver: out("300\n", 0)}, check.Pass, "5 min"},
		{"never", map[string]runner.Output{pm: pmset("0"), saver: out("0\n", 0)}, check.Fail, "never"},
		{"pmset missing", notFound, check.Unknown, ""},
	})
	runCases(t, "macos.auto_lock", homeDesktop, []macCase{
		{"10 min at home", ten, check.Pass, "10 min"},
	})
}

func TestMacOSFindMy(t *testing.T) {
	k := runner.Key("nvram", "fmm-mobileme-token-FMM")
	runCases(t, "macos.find_my", laptop, []macCase{
		{"token present", map[string]runner.Output{k: out("fmm-mobileme-token-FMM\tbplist00secret\n", 0)}, check.Pass, "on"},
		{"token absent 26.2", map[string]runner.Output{k: errOut("nvram: Error getting variable - 'fmm-mobileme-token-FMM': (iokit/common) data was not found", 1)}, check.Manual, ""},
	})
	got := runWith(t, macCheck(t, "macos.find_my"), laptop, map[string]runner.Output{k: out("fmm-mobileme-token-FMM\tbplist00secret\n", 0)})
	if want := "fmm-mobileme-token-FMM <present>"; !strings.Contains(got.Evidence, want) || strings.Contains(got.Evidence, "secret") {
		t.Errorf("evidence must not leak the token: %q", got.Evidence)
	}
}

func TestMacOSUpdateSettings(t *testing.T) {
	key := func(k string) string { return runner.Key("defaults", "read", softwareUpdate, k) }
	all := map[string]runner.Output{
		key("AutomaticCheckEnabled"):            errOut(missingKey, 1), // 26.2: missing means on
		key("AutomaticDownload"):                out("1\n", 0),
		key("AutomaticallyInstallMacOSUpdates"): out("1\n", 0),
	}
	noInstall := map[string]runner.Output{
		key("AutomaticCheckEnabled"):            out("1\n", 0),
		key("AutomaticDownload"):                out("1\n", 0),
		key("AutomaticallyInstallMacOSUpdates"): errOut(missingKey, 1), // missing means off
	}
	none := map[string]runner.Output{
		key("AutomaticCheckEnabled"):            out("0\n", 0),
		key("AutomaticDownload"):                out("0\n", 0),
		key("AutomaticallyInstallMacOSUpdates"): out("0\n", 0),
	}
	runCases(t, "macos.auto_updates", laptop, []macCase{
		{"all on 26.2", all, check.Pass, ""},
		{"install off", noInstall, check.Warn, "off: AutomaticallyInstallMacOSUpdates"},
		{"all off", none, check.Fail, ""},
	})
	runCases(t, "macos.security_responses", laptop, []macCase{
		{"on 26.2", map[string]runner.Output{key("ConfigDataInstall"): out("1\n", 0), key("CriticalUpdateInstall"): out("1\n", 0)}, check.Pass, ""},
		{"defaults", map[string]runner.Output{key("ConfigDataInstall"): errOut(missingKey, 1), key("CriticalUpdateInstall"): errOut(missingKey, 1)}, check.Pass, ""},
		{"off", map[string]runner.Output{key("ConfigDataInstall"): out("0\n", 0), key("CriticalUpdateInstall"): out("0\n", 0)}, check.Fail, ""},
	})
}

func TestMacOSPendingUpdates(t *testing.T) {
	defer check.SetNow(func() time.Time { return time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC) })()
	ver := runner.Key("sw_vers", "-productVersion")
	recs := runner.Key("defaults", "read", softwareUpdate, "RecommendedUpdates")
	last := runner.Key("defaults", "read", softwareUpdate, "LastFullSuccessfulDate")
	// Captured on 26.2: a minor update with a marketing name, and a new major.
	pending := `(
        {
        "Display Name" = "Command Line Tools for Xcode 26.5";
        "Product Key" = "047-91568";
    },
        {
        "Display Name" = "macOS 27.2";
        MobileSoftwareUpdate = 1;
    },
        {
        "Display Name" = "macOS Tahoe 26.7.2";
        MobileSoftwareUpdate = 1;
    }
)`
	onlyMajor := `(
        {
        "Display Name" = "macOS 27.2";
    }
)`
	runCases(t, "macos.pending_updates", laptop, []macCase{
		{"minor pending 26.2", map[string]runner.Output{ver: out("26.2\n", 0), recs: out(pending, 0), last: out("2026-11-06 05:35:35 +0000\n", 0)}, check.Fail, "26.2, pending: macOS Tahoe 26.7.2"},
		{"only a new major", map[string]runner.Output{ver: out("26.7.2\n", 0), recs: out(onlyMajor, 0), last: out("2026-11-06 05:35:35 +0000\n", 0)}, check.Pass, "26.7.2, upgrade available: macOS 27.2"},
		{"nothing pending", map[string]runner.Output{ver: out("26.7.2\n", 0), recs: errOut(missingKey, 1), last: out("2026-11-01 05:35:35 +0000\n", 0)}, check.Pass, "26.7.2, up to date"},
		{"not checked for 40 days", map[string]runner.Output{ver: out("26.7.2\n", 0), recs: errOut(missingKey, 1), last: out("2026-09-27 05:35:35 +0000\n", 0)}, check.Warn, ""},
		{"no sw_vers", notFound, check.Unknown, ""},
	})
}
