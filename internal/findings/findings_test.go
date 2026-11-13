package findings_test

import (
	"context"
	"strings"
	"testing"

	"github.com/XexusCode/hardenme/internal/check"
	"github.com/XexusCode/hardenme/internal/findings"
	"github.com/XexusCode/hardenme/internal/runner"
	"github.com/XexusCode/hardenme/internal/scan"
)

const home = "/Users/ana"

func key(name string, args ...string) string { return runner.Key(name, args...) }

func out(s string) runner.Output { return runner.Output{Stdout: s} }

// Signature lines as printed by `codesign -dv --verbose=2` on macOS 26.2.
const (
	signedDev   = "Executable=/x\nAuthority=Developer ID Application: Google LLC (EQHXZ8M8AV)\nAuthority=Developer ID Certification Authority\n"
	signedAdhoc = "CodeDirectory v=20400 size=87422 flags=0x20002(adhoc,linker-signed)\nSignature=adhoc\nTeamIdentifier=not set\n"
	notSigned   = "/x: code object is not signed at all\n"
	denied      = "/Library/PrivilegedHelperTools/com.docker.socket: Permission denied\n"
)

// cleanMac answers like a normal Mac: captured shapes from macOS 26.2 with
// the names changed.
func cleanMac() map[string]runner.Output {
	agents := home + "/Library/LaunchAgents"
	return map[string]runner.Output{
		key("/bin/ls", "-1", "/Applications"):          out("Safari.app\nGoogle Chrome.app\nUtilities\n"),
		key("/bin/ls", "-1", agents):                   out("com.google.GoogleUpdater.wake.plist\ncom.google.keystone.agent.plist\n"),
		key("/bin/ls", "-1", "/Library/LaunchAgents"):  out(""),
		key("/bin/ls", "-1", "/Library/LaunchDaemons"): out("com.docker.socket.plist\n"),
		key("/usr/bin/plutil", "-extract", "Program", "raw", agents+"/com.google.GoogleUpdater.wake.plist"):            {ExitCode: 1},
		key("/usr/bin/plutil", "-extract", "ProgramArguments.0", "raw", agents+"/com.google.GoogleUpdater.wake.plist"): out(home + "/Library/Application Support/Google/GoogleUpdater\n"),
		key("/usr/bin/codesign", "-dv", "--verbose=2", home+"/Library/Application Support/Google/GoogleUpdater"):       {Stderr: signedDev},
		key("/usr/bin/plutil", "-extract", "Program", "raw", agents+"/com.google.keystone.agent.plist"):                {ExitCode: 1},
		key("/usr/bin/plutil", "-extract", "ProgramArguments.0", "raw", agents+"/com.google.keystone.agent.plist"):     {ExitCode: 1},
		key("/usr/bin/plutil", "-extract", "Program", "raw", "/Library/LaunchDaemons/com.docker.socket.plist"):         out("/Library/PrivilegedHelperTools/com.docker.socket\n"),
		key("/usr/bin/codesign", "-dv", "--verbose=2", "/Library/PrivilegedHelperTools/com.docker.socket"):             {Stderr: denied, ExitCode: 1},
		key("/usr/bin/profiles", "list"):                                       out("There are no configuration profiles installed for user 'ana'\n"),
		key("/usr/bin/profiles", "status", "-type", "enrollment"):              out("Enrolled via DEP: No\nMDM enrollment: No\n"),
		key("/usr/bin/dscl", ".", "-read", "/Groups/admin", "GroupMembership"): out("GroupMembership: root ana _mbsetupuser\n"),
		key("/usr/sbin/scutil", "--proxy"):                                     out("<dictionary> {\n  ExceptionsList : <array> {\n    0 : *.local\n  }\n  FTPPassive : 1\n}\n"),
		key("/bin/cat", "/etc/hosts"):                                          out("##\n# Host Database\n##\n127.0.0.1\tlocalhost\n255.255.255.255\tbroadcasthost\n::1             localhost\n"),
	}
}

func scanWith(outputs map[string]runner.Output) []scan.Finding {
	return findings.Scan(context.Background(), findings.Env{Runner: &runner.Fake{Outputs: outputs}, Home: home, User: "ana"})
}

func find(fs []scan.Finding, kind string) []string {
	for _, f := range fs {
		if f.Kind == kind {
			return f.Items
		}
	}
	return nil
}

func TestANormalMacHasNoFindings(t *testing.T) {
	if got := scanWith(cleanMac()); len(got) != 0 {
		t.Fatalf("a normal Mac produced findings: %+v", got)
	}
}

func TestNothingReadableMeansNothingReported(t *testing.T) {
	if got := scanWith(map[string]runner.Output{}); len(got) != 0 {
		t.Fatalf("unreadable state must not become a finding: %+v", got)
	}
}

func TestEachSignIsFound(t *testing.T) {
	agents := home + "/Library/LaunchAgents"
	m := cleanMac()
	m[key("/bin/ls", "-1", "/Applications")] = out("Safari.app\nAnyDesk.app\nTeamViewer.app\nTeamViewer notes.txt\n")
	m[key("/bin/ls", "-1", agents)] = out("com.x.helper.plist\ncom.y.sync.plist\ncom.z.adhoc.plist\n")
	m[key("/usr/bin/plutil", "-extract", "Program", "raw", agents+"/com.x.helper.plist")] = out(home + "/.hidden/helper\n")
	m[key("/usr/bin/codesign", "-dv", "--verbose=2", home+"/.hidden/helper")] = runner.Output{Stderr: notSigned, ExitCode: 1}
	m[key("/usr/bin/plutil", "-extract", "Program", "raw", agents+"/com.y.sync.plist")] = runner.Output{ExitCode: 1}
	m[key("/usr/bin/plutil", "-extract", "ProgramArguments.0", "raw", agents+"/com.y.sync.plist")] = out("/bin/sh\n")
	m[key("/usr/bin/plutil", "-extract", "Program", "raw", agents+"/com.z.adhoc.plist")] = out("/usr/local/bin/tool\n")
	m[key("/usr/bin/codesign", "-dv", "--verbose=2", "/usr/local/bin/tool")] = runner.Output{Stderr: signedAdhoc}
	// The format of `profiles list` with a profile installed was not
	// captured on a real Mac: this output is constructed.
	m[key("/usr/bin/profiles", "list")] = out("ana[1] attribute: profileIdentifier: com.evil.proxy\nThere are 1 configuration profiles installed for user 'ana'\n")
	m[key("/usr/bin/profiles", "status", "-type", "enrollment")] = out("Enrolled via DEP: No\nMDM enrollment: Yes (User Approved)\n")
	m[key("/usr/bin/dscl", ".", "-read", "/Groups/admin", "GroupMembership")] = out("GroupMembership: root ana _mbsetupuser support\n")
	m[key("/bin/cat", home+"/.ssh/authorized_keys")] = out("# mine\nssh-ed25519 AAAAC3Nza laptop@work\nssh-rsa AAAAB3Nza\n")
	// Constructed from the documented keys of `scutil --proxy`.
	m[key("/usr/sbin/scutil", "--proxy")] = out("<dictionary> {\n  HTTPEnable : 1\n  HTTPPort : 8080\n  HTTPProxy : 10.0.0.9\n  HTTPSEnable : 0\n}\n")
	m[key("/bin/cat", "/etc/hosts")] = out("127.0.0.1 localhost\n10.0.0.9 mibanco.es www.mibanco.es # x\n127.0.0.1 app.localhost\n")

	got := scanWith(m)
	want := map[string][]string{
		findings.RemoteTool:      {"AnyDesk", "TeamViewer"},
		findings.StartupUnsigned: {"com.x.helper (~/.hidden/helper)", "com.z.adhoc (/usr/local/bin/tool)"},
		findings.StartupScript:   {"com.y.sync"},
		findings.Profile:         {"com.evil.proxy"},
		findings.Managed:         {"MDM"},
		findings.ExtraAdmin:      {"support"},
		findings.SSHKeys:         {"(ssh-rsa)", "laptop@work"},
		findings.Proxy:           {"HTTP → 10.0.0.9"},
		findings.Hosts:           {"mibanco.es → 10.0.0.9", "www.mibanco.es → 10.0.0.9"},
	}
	for kind, items := range want {
		if g := find(got, kind); strings.Join(g, "|") != strings.Join(items, "|") {
			t.Errorf("%s: got %v, want %v", kind, g, items)
		}
	}
	if len(got) != len(want) {
		t.Errorf("got %d kinds, want %d: %+v", len(got), len(want), got)
	}
	for i, f := range got {
		if f.Kind != findings.Kinds[i] {
			t.Errorf("order: position %d is %s", i, f.Kind)
		}
	}
	for _, f := range got {
		for _, item := range f.Items {
			if strings.Contains(item, "AAAA") {
				t.Errorf("key material leaked into the report: %s", item)
			}
		}
	}
}

func TestLongListsAreCapped(t *testing.T) {
	m := cleanMac()
	var hosts strings.Builder
	for _, h := range []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"} {
		hosts.WriteString("10.0.0.1 " + h + ".example.com\n")
	}
	m[key("/bin/cat", "/etc/hosts")] = out(hosts.String())
	items := find(scanWith(m), findings.Hosts)
	if len(items) != 9 || items[8] != "+3" {
		t.Errorf("items %v", items)
	}
}

func TestEveryKindIsExplainedInBothLanguages(t *testing.T) {
	for _, kind := range findings.Kinds {
		for _, lang := range []check.Lang{check.EN, check.ES} {
			text := findings.TextFor(kind, lang)
			if text.Title == "" || text.Why == "" || text.Do == "" {
				t.Errorf("%s (%s): incomplete text", kind, lang)
			}
			if !strings.HasSuffix(text.Title, "?") {
				t.Errorf("%s (%s): the title must be a question, got %q", kind, lang, text.Title)
			}
			if strings.ContainsAny(text.Title+text.Why+text.Do, "—–") {
				t.Errorf("%s (%s): long dash", kind, lang)
			}
		}
		if findings.TextFor(kind, check.EN).Why == findings.TextFor(kind, check.ES).Why {
			t.Errorf("%s: Spanish text missing", kind)
		}
	}
}
