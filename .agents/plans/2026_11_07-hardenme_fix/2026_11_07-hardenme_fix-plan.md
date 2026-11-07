---
name: "hardenme fix"
description: "Let the user pick the improvements from a scan and have hardenme apply the safe ones on macOS, with one native password prompt, verification, and undo."
created_at: "2026-11-07T15:30:00Z"

created_by:
  tool: "Claude Code"
  model:
    name: "Claude Opus"
    version: "5.5"
    reasoning_effort: "low"

has_completed_all_phases: "false"
---

# hardenme fix

## 🎯 Goal

After a scan, a non-technical user picks the improvements they want from a numbered list, and hardenme applies them.
- The safe changes are applied after one native macOS password prompt.
- For the rest, hardenme opens the exact System Settings pane and waits for the user.
- Every change is checked again afterwards.
- Every automatic change can be undone.

This is v0.2 in the [roadmap](../2026_11_06-hardenme_security_auditor/roadmap.md). It replaces "`--fix` later": the maintainer decided on 2026-11-07 that applying fixes comes before the mobile checklist.

## 👀 Context

- Read [`AGENTS.md`](../../../AGENTS.md) and the first plan, [`2026_11_06-hardenme_security_auditor-plan.md`](../2026_11_06-hardenme_security_auditor/2026_11_06-hardenme_security_auditor-plan.md). This plan builds on its packages:
  - `check`: YAML, schema and probes;
  - `scan`: report, score and grade;
  - `runner`: `Exec` and `Fake`;
  - `cli`.
- 21 macOS checks live in `checks/macos/*.yaml`. Only `firewall`, `firewall_stealth`, `gatekeeper` and `sip` carry a `command` in their text today, and those strings are for display only.
- Prior art: fort (MIT) has a `--fix` mode with per-item confirmation. Credit it for the interaction pattern. Never copy Lynis or CIS remediation text.
- **Rule change, written down on purpose.** The audit still never needs sudo. `fix` may change system settings, but only:
  - through one `osascript -e 'do shell script "…" with administrator privileges'` per batch, which shows the native password dialog;
  - with commands that are static strings embedded in the binary. Nothing from probe output, files or the user is ever interpolated into a command;
  - after the user has seen the list of changes and confirmed it.
- Never ask for or read the user's password. The only exception is `sysadminctl -screenLock`, which reads it itself. If that check cannot run without hardenme seeing the password, it stays a guided fix.

### Fix classes (initial proposal, each one verified on a real Mac in Phase 1)

| Class | Checks | Candidate command (to verify) |
|---|---|---|
| auto, admin | firewall | `socketfilterfw --setglobalstate on` |
| auto, admin | firewall_stealth | `socketfilterfw --setstealthmode on` |
| auto, admin | guest_account | `sysadminctl -guestAccount off` |
| auto, admin | auto_login | `defaults delete /Library/Preferences/com.apple.loginwindow autoLoginUser` and `rm -f /etc/kcpassword` |
| auto, admin | auto_updates, security_responses | `defaults write /Library/Preferences/com.apple.SoftwareUpdate <key> -bool true` (check that System Settings and softwareupdated respect it on 26) |
| auto, mixed | auto_lock | `pmset -a displaysleep N` (admin) and `defaults -currentHost write com.apple.screensaver idleTime` (user) |
| auto, admin | gatekeeper | `spctl --global-enable` (may need GUI confirmation on recent macOS, so it may become guided) |
| auto?, admin | remote_login, screen_sharing, file_sharing, remote_management | `systemsetup` needs Full Disk Access on recent macOS; try `launchctl disable/bootout` and the ARD `kickstart -deactivate`. Each one carries a warning: "if you use this to connect from another computer, keep it". |
| guided | filevault | Opens the pane. Never by command: the recovery key must be saved by the person |
| guided | find_my, icloud_adp, lockdown_mode, airdrop, admin_account, pending_updates | Opens the pane or Software Update. Restarting is the user's choice |
| guided? | screen_lock_delay | `sysadminctl -screenLock immediate -password -` if it reads the password from the TTY by itself; otherwise guided |
| info only | sip | Recovery-mode steps, never automated |

A check whose fix cannot be verified on a real Mac ships as `guided`.

## 📜 Public contracts

### CLI
- `hardenme fix`: runs a scan, lists the failing checks with a fix in score-impact order, asks which to apply, shows what will change, asks for confirmation, applies, re-checks, and prints the score before and after.
- `hardenme fix --only <ids>`: limits the list to those checks.
- `hardenme fix --dry-run`: prints the exact commands and the settings that would open, and changes nothing.
- `hardenme fix --yes`: skips the confirmation for scripted use. It still needs the native password dialog. Without a terminal and without `--yes`, `fix` refuses to run.
- `hardenme fix --undo`: reverts the last batch of automatic changes, using the change log.
- `hardenme scan`: after the report, when stdin is a terminal and there are fixable findings, it adds a line saying to run `hardenme fix`.
- `scan.command` (the double-click launcher) runs `hardenme scan` and then `hardenme fix`, so the user never types a command.
- Exit codes: 0 when everything selected was fixed, 1 when something was not fixed or the user cancelled before applying, 3 for usage errors.

### Check data (`checks/schema.json`)
A new optional `fix` block:
```yaml
fix:
  kind: auto | guided | info
  admin: true            # auto only: needs the password prompt
  apply: [[/usr/libexec/ApplicationFirewall/socketfilterfw, --setglobalstate, "on"]]
  undo:  [[/usr/libexec/ApplicationFirewall/socketfilterfw, --setglobalstate, "off"]]
  value_from_profile: auto_lock_minutes   # optional, for profile-dependent values
```
- `apply` and `undo` are lists of argv arrays, never shell strings.
- `auto` requires `apply` and `undo`.
- `guided` uses the existing `deep_link` and `fix_steps`.
- A new optional `fix_warning` per language in `text`.
- `catalog_test` enforces:
  - an `auto` fix has an `undo`;
  - every argv is an absolute path;
  - no argv contains `$`, a backtick or `;`.

### Packages
- `internal/fix`:
  - `type Action struct { CheckID string; Kind; Admin bool; Apply, Undo [][]string }`.
  - `Plan(report scan.Report, catalog []check.Check) []Action` lists the failing and warned checks with a fix, ordered by weight then id.
  - `Script(actions []Action) string` builds the POSIX script for one admin batch. Every argv element is single-quoted. Each action prints `HARDENME_RESULT <id> <exit>` and the script does not stop on a failure.
  - `Apply(ctx, runner, actions) []Outcome` runs the user-level actions directly and the admin batch through one `osascript` call. It then re-runs each check's probe. The outcome is `fixed`, `not_changed`, `failed` or `cancelled` (the password dialog was cancelled).
  - `Log`: appends a batch of `{time, check_id, undo}` to `~/.local/state/hardenme/changes.json` (it honours `XDG_STATE_HOME`; file mode 0600). `Undo` replays the last batch in reverse order.
- `internal/render`: the fix menu, the change summary and the before/after lines, in EN/ES.

### Test suites
- `fix/script_test`: quoting with spaces and single quotes, the result-line parsing, and a cancelled dialog (osascript error -128 maps to `cancelled`).
- `fix/plan_test`: ordering, `not_applicable` and `pass` excluded, and guided fixes kept with no commands.
- `fix/apply_test` with `runner.Fake`:
  - an admin batch runs in a single osascript call;
  - the re-check decides the outcome, not the exit code;
  - the log is written only for actions that changed something.
- `catalog_test`: the new schema rules.
- `cli` tests for the menu input: `1,3`, `todo`/`all`, blank to exit, and invalid input asked again.
- Golden files for the EN/ES menu and the summary.

### Text shown to end users (ES and EN, plain language)
- "Puedo arreglar esto por ti: 1) Activar el cortafuegos …"
- "Escribe los números separados por comas, `todo`, o pulsa Intro para salir."
- "macOS te pedirá tu contraseña una vez. hardenme no la ve."
- "Esto lo tienes que hacer tú. Abro el ajuste; cuando termines, pulsa Intro y lo compruebo."
- "Nota antes: 57 (F). Nota ahora: 81 (B)."
- "Para deshacer estos cambios: `hardenme fix --undo`."

## 🪜 Phases

### Phase 1: Verify every candidate fix on a real Mac

**Description:** Before writing the engine, confirm for each auto candidate that it works on macOS 26 from Terminal with administrator privileges, that the check passes afterwards, and that the undo restores the previous state.

- [ ] For each auto candidate in the table, on the maintainer's Mac and with their go-ahead per change:
  - record the probe output;
  - apply;
  - run the probe again;
  - undo;
  - run the probe once more.
  Capture the outputs as fixtures. Check that the setting in System Settings matches.
- [ ] Mark every candidate that needs Full Disk Access, a GUI confirmation or a reboot, or that macOS ignores, as `guided`.
- [ ] Decide how to handle `screen_lock_delay` and the `remote_*` checks, from what was observed.
- [ ] Write the results in this plan's implementation notes and update the fix-class table.
- [ ] STOP. Present the table to the user before Phase 2.

**Implementation notes (Phase 1, maintainer's MacBook, macOS 26.2, 2026-11-07):**
- One `osascript … with administrator privileges` call ran a 13-step root script behind a single native password dialog, and every step returned 0. This confirms the one-dialog-per-batch design.
- The maintainer chose to leave the failing settings fixed after the undo test. Afterwards, `hardenme scan --only firewall,firewall_stealth,auto_lock,guest_account,auto_updates` gave 5/5 pass.

| Check | Verified | Apply | Undo | Result |
|---|---|---|---|---|
| firewall | apply, undo, reapply | `socketfilterfw --setglobalstate on` | `… off` | **auto, admin** |
| firewall_stealth | apply, undo, reapply | `socketfilterfw --setstealthmode on` | `… off` | **auto, admin**. Needs the firewall on, so order it after `firewall` |
| auto_lock | apply, undo, reapply | `pmset -c displaysleep 5` (admin). `defaults -currentHost write com.apple.screensaver idleTime -int 300` (user) | the previous value | **auto, mixed** |
| guest_account | weaken, apply | `sysadminctl -guestAccount off` | `… on` | **auto, admin**. Turning it on creates the account, which takes about 2 s |
| auto_updates | weaken, apply (`AutomaticDownload`) | `defaults write /Library/Preferences/com.apple.SoftwareUpdate <key> -bool true` | the previous value | **auto, admin**. The probe reflects the change. Not checked: whether System Settings shows it |
| security_responses | not tested (same mechanism) | same, for `ConfigDataInstall` and `CriticalUpdateInstall` | the previous value | auto, admin, tested only on fixtures until checked |
| auto_login | not testable: auto-login was not enabled | — | — | **guided** |
| gatekeeper | not tested (the maintainer excluded it) | — | — | **guided** |
| remote_login, screen_sharing, file_sharing, remote_management | not tested (the maintainer uses this Mac and asked not to touch SSH) | — | — | **guided** |

**Design consequences for Phase 2:**
- **Undo cannot always be a static argv.**
  - `pmset` keeps separate values for battery (`-b`) and AC (`-c`). On this Mac, battery was already 5 and AC was 120.
  - `idleTime` and the SoftwareUpdate keys may be absent, which means the macOS default.
  - So a `fix` needs a `read` step that captures the previous value at apply time. The change log stores the concrete undo argv: the previous value, or `defaults delete` when the key was absent.
  - The catalog rule therefore becomes "an `auto` fix has an `undo` or an `undo_from_read`", not "a static `undo`".
- **Order inside a batch matters**: `firewall` comes before `firewall_stealth`.
- **`pmset`:** only the source that fails is changed. The probe reads the active source, and the fix applies to the source that is over the limit (`-c`, `-b`, or both).

### Phase 2: Fix engine, menu and undo

**Description:** Implement `fix` data, the `internal/fix` package and `hardenme fix` with the interactive menu, dry run, re-check, change log and undo.

- [ ] Extend `checks/schema.json`, `check.Check` and `compile()` with `fix`, `fix_warning` and the safety rules. Add `fix` blocks to the checks verified in Phase 1, with EN/ES `fix_warning` where it applies.
- [ ] Implement `fix.Plan`, `fix.Script`, `fix.Apply`, `fix.Log` and `fix.Undo`, with their tests.
- [ ] Implement `hardenme fix` (`--only`, `--dry-run`, `--yes`, `--undo`), the EN/ES menu and the summary rendering, with golden tests.
- [ ] Guided fixes: open the `deep_link` with `open`, wait for Return, then re-check.
- [ ] Add the "run `hardenme fix`" hint to `scan`, and change `install.sh` so `scan.command` runs `hardenme fix` after the scan.
- [ ] Run `hardenme fix --dry-run` and a real `hardenme fix` on the maintainer's Mac, with their go-ahead. Then run `--undo` and check the score returns to the earlier value.
- [ ] Verify with `gofmt -l . && go vet ./... && go test ./...`.
- [ ] Update the README files (EN/ES) with the fix flow and the undo command.
- [ ] STOP. Present the changes and suggest commit messages. Release v0.2.0 only when the user asks.

**Done when:** a non-technical person double-clicks hardenme.app, picks the numbers of what to fix, types their password once in the macOS dialog, sees the new score, and can undo it with a single command.

## ⏭️ Next step

Phase 1: verify the candidate fixes on the maintainer's Mac, one change at a time with their go-ahead.
