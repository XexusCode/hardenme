---
name: "hardenme security auditor"
description: "Open-source Go CLI and bilingual guide that audits a normal user's device security (macOS, Android, Windows, iOS) with a profile-weighted score and a strong focus on device theft."
created_at: "2026-11-06T14:30:00Z"

created_by:
  tool: "Claude Code"
  model:
    name: "Claude Opus"
    version: "5.5"
    reasoning_effort: "low"
---

# hardenme security auditor

## 🎯 Goal

Publish `hardenme` as an MIT open-source repository on GitHub. It ships two things:
- A single-binary Go CLI that audits a non-technical user's device and returns a 0-100 security score weighted by their profile. Every finding comes with a plain-language reason and the steps to fix it.
- A bilingual (EN/ES) guide covering general hardening and what to do before and after a device theft.

## 👀 Context

- New repository: [`/Users/jesus/Documents/GitHub/hardenme`](../../../). It is empty apart from this plan and the research.
- Prior-art research: [`docs/research/prior-art.md`](../../../docs/research/prior-art.md). It contains the design lessons, the check catalogue for each OS and the licensing rules. **Read it before implementing any phase.** Items marked _unverified_ there must be confirmed on a real device before they ship.
- Draft credits: [`docs/research/CREDITS.draft.md`](../../../docs/research/CREDITS.draft.md). This becomes `CREDITS.md` in Phase 1.
- There is no `AGENTS.md` yet. Phase 1 creates it with the conventions below.
- Toolchain: Go is not installed yet (`brew install go`). `adb` is needed from Phase 2 (`brew install android-platform-tools`). `libimobiledevice` is optional in Phase 3.
- Conventions:
  - Go stdlib only where possible. CLI framework: `cobra`.
  - Checks are declarative YAML data embedded with `go:embed` and validated against a JSON Schema.
  - Every system command runs through an injectable `Runner`, so tests use fixture outputs captured per OS version.
  - **The audit never needs sudo, never reads TCC-protected data and never sends data off the device.**
  - Read effective state, never policy keys alone.
  - Verification command (defined in `AGENTS.md` in Phase 1): `gofmt -l . && go vet ./... && go test ./...`.
- Licensing rules for adapted content are in section 5 of the prior-art research. Every adapted check sets `credits`, and MIT/Apache sources keep their license under `third_party/<project>/LICENSE`.

## 📜 Public contracts

### CLI
- `hardenme scan [--format text|json|markdown] [--lang en|es] [--only <ids>]`: audits the local machine.
- `hardenme scan --android [--serial <id>]`: audits an Android phone over adb (Phase 2).
- `hardenme profile [--portable] [--leaves-home] [--shared] [--sensitive-data] [--risk normal|elevated]`: creates or edits the profile at `~/.config/hardenme/profile.json`. Without flags it asks interactively.
- `hardenme checklist ios`: guided iOS checklist, plus USB reads when libimobiledevice or go-ios is available (Phase 3).
- `hardenme guide [theft-before|theft-after|accounts]`: prints a guide section in the chosen language.
- Exit codes: `0` all essential checks pass, `1` any essential or recommended check fails, `2` only warnings.

### Application services and types (package `internal/...`)
- `type Runner interface { Run(ctx context.Context, name string, args ...string) (stdout []byte, err error) }` with an `ExecRunner` and a `FakeRunner` for tests.
- `type Profile struct { Portable, LeavesHome, Shared, SensitiveData bool; Risk RiskLevel }`
- `ProfileStore.Load() (Profile, error)` / `ProfileStore.Save(Profile) error`
- `ProfileDetector.Detect(ctx, Runner) Profile` detects `Portable` by checking for a battery.
- `type Check struct { ID, Platform, Category, Severity string; Relevance []RelevanceRule; Probe Probe; MinOS, MaxOS string; Text map[Lang]CheckText; References []string; Credits []string }`
- `Check.RelevanceFor(Profile) Relevance` returns `essential | recommended | optional | not_applicable`.
- `type Result struct { CheckID string; Status Status; Current, Expected, Evidence string; Partial float64 }`, where `Status` is `pass | fail | warn | unknown | manual | skipped`.
- `Catalog.Load(platform) ([]Check, error)` loads and validates the embedded YAML.
- `Scanner.Scan(ctx, platform, Profile) Report`
- `Report.Score() Score{Value int; Grade string}`. The score is `Σ(earned×w)/Σ(max×w)×100` with w = 3/2/1/0. `unknown`, `manual` and `skipped` are excluded.
- `Report.TheftStories() []Story` correlates failing checks into plain-language theft scenarios (Phase 2).
- `Renderer.Render(w io.Writer, Report, Format, Lang) error`

### Check data schema
- `checks/schema.json` (JSON Schema).
- `checks/<platform>/*.yaml`, one file per check.
- Required fields: `id`, `platform`, `category`, `severity`, `relevance`, `probe`, `text.en`, `text.es`, `references`.
- Each `text.<lang>` holds `title_pass`, `title_fail`, `why`, `fix_steps[]` and optional `deep_link` and `command`.

### Test suites
- `catalog_test`: every embedded check validates against the schema; IDs are unique and stable; both languages are present; references are non-empty; adapted checks carry credits.
- `profile_test`: save/load round trip; defaults when no file exists; portable detection from fixture output (battery / no battery).
- `scoring_test`:
  - weights per relevance;
  - `not_applicable` is excluded;
  - `unknown`/`manual`/`skipped` are excluded;
  - partial credit;
  - grade boundaries;
  - BitLocker/FileVault is essential on a portable profile and optional on a desktop that never leaves home and is not shared;
  - the score is 100 when nothing applicable fails.
- `renderer_test`: golden files for text, json and markdown in EN and ES.
- `checks/macos_test`: one table-driven test per check covering pass, fail and unknown (command error or unexpected output), with fixtures for macOS 14, 15 and 26.
- `checks/android_test`: the same, with fixtures per SDK level, including a null settings value → unknown and a future patch date → fail.
- `checks/windows_test`: the same, with fixtures for PowerShell and registry output, plus a non-admin case → skipped.
- `checks/ios_test`: parsing `ideviceinfo` output and building the manual checklist.
- `stories_test`: theft stories fire only when every contributing check failed.

### Text copies shown to end users
- The report header: the score, the grade with its casual label, and the issue count.
- For each check: the pass and fail titles, the why text, the fix steps and the deep link. All in EN and ES.
- The profile questions.
- The theft stories.
- The guide pages: before a theft, after a theft (including "write down your IMEI now"), and accounts (2FA with an app, password manager, SIM PIN).

## 🪜 Phases

### Phase 1: Repository foundation and macOS audit with profile-weighted score

**Description:** Deliver a runnable `hardenme scan` on macOS that asks for the user's profile, runs the macOS checks and prints a weighted score with fixes in EN/ES. Also publish the open-source scaffolding and the first guide pages.

- [ ] Install Go (`brew install go`) and initialise the module `github.com/<owner>/hardenme` (confirm the GitHub owner with the user).
- [ ] Add the open-source scaffolding:
  - `LICENSE` (MIT) and `README.md` (EN) with `README.es.md` (ES).
  - `CONTRIBUTING.md`, which requires references for every check and states the licensing rules from the research.
  - `CODE_OF_CONDUCT.md` and `SECURITY.md`.
  - `CREDITS.md`, promoted from `docs/research/CREDITS.draft.md`.
  - `third_party/fort/LICENSE`.
- [ ] Add `AGENTS.md` with the architecture, the conventions and the verification command.
- [ ] Implement `Runner` (`ExecRunner` and `FakeRunner`).
- [ ] Implement `Profile`, `ProfileStore`, `ProfileDetector` (macOS battery detection via `pmset -g batt`) and `hardenme profile` (interactive and flag-based).
- [ ] Implement the check schema, `Catalog.Load` with `go:embed`, and `catalog_test`.
- [ ] Implement `Scanner`, `Report.Score` and `scoring_test`.
- [ ] Add the macOS checks from the catalogue in `docs/research/prior-art.md` §3. Each one gets a YAML definition with EN/ES text and a `x-apple.systempreferences:` deep link where one exists. Checks adapted from fort credit it.
- [ ] Add the macOS check tests with fixtures.
- [ ] Verify the _unverified_ macOS probes on this Mac: `sysadminctl -screenLock status`, `nvram fmm-mobileme-token-FMM`, and the update keys on macOS 14+. Fix the probes or mark those checks `manual`.
- [ ] Implement the text `Renderer`, the `--lang` flag and the text golden tests.
- [ ] Wire up `hardenme scan` and the exit codes.
- [ ] Write the guide pages `docs/guide/{en,es}/theft-before.md` and `theft-after.md`, and the `hardenme guide` command that prints them.
- [ ] Run `go run . scan` on this Mac and attach the output to the review.
- [ ] Verify the changes in terms of typechecking, linting and tests using the project's verification command (`gofmt -l . && go vet ./... && go test ./...`). Fix issues if any.
- [ ] STOP. Present the changes to the user for review and suggest commit messages (or pull request titles, when the phases are implemented through pull requests). Do NOT proceed to the next phase until the user explicitly asks.

### Phase 2: Android audit over adb and theft stories

**Description:** Extend `hardenme scan --android` to audit a USB-connected Android phone without root, add JSON and Markdown output, and correlate failures into theft stories.

- [ ] Install `adb` (`brew install android-platform-tools`). Implement device discovery: `adb devices`, `--serial`, and clear messages for unauthorised or missing devices.
- [ ] Add the Android checks from `docs/research/prior-art.md` §3:
  - patch age with partial credit;
  - encryption, screen lock (via `dumpsys lock_settings`) and auto-lock;
  - lock-screen notifications and trust agents;
  - unknown-source apps (via `appops`), Play Protect and wireless debugging;
  - accessibility services, notification listeners and device admins, listed by name;
  - bootloader and verified boot;
  - root and user CA certificates.

  Checks adapted from HARDAX credit it, and `third_party/hardax/LICENSE` is added.
- [ ] Show USB debugging as an end-of-report reminder to turn it off, never as a failure.
- [ ] Add Find My Device, Theft Detection Lock and Identity Check as best-effort probes that fall back to `manual`. Make SIM PIN a `manual` item.
- [ ] Add the Android check tests with fixtures for each SDK level.
- [ ] Verify the _unverified_ Android settings keys on at least one real phone (ask the user to connect one), and update the probes or mark them `manual`.
- [ ] Implement `Report.TheftStories()` and `stories_test`, and render the stories in the report.
- [ ] Add the JSON and Markdown renderers with golden tests, including an evidence transcript of the commands that ran.
- [ ] Write the guide page `docs/guide/{en,es}/accounts.md` (2FA with an app, password manager, SIM PIN, recovery codes).
- [ ] Run `go run . scan --android` against a real phone and attach the output to the review.
- [ ] Verify the changes in terms of typechecking, linting and tests using the project's verification command (`gofmt -l . && go vet ./... && go test ./...`). Fix issues if any.
- [ ] STOP. Present the changes to the user for review and suggest commit messages (or pull request titles, when the phases are implemented through pull requests). Do NOT proceed to the next phase until the user explicitly asks.

### Phase 3: Windows audit, iOS checklist and public release

**Description:** Add the Windows audit, which reads effective state through PowerShell and the registry without admin. Add the guided iOS checklist with optional USB reads. Then publish the project with CI and signed release binaries.

- [ ] Add the Windows checks from `docs/research/prior-art.md` §3:
  - BitLocker, with a non-admin fallback and `skipped` when it can't be read;
  - Defender real-time, tamper and PUA protection;
  - the firewall per profile;
  - UAC, SmartScreen, update recency and the screen-lock timeout;
  - Remote Desktop, SMBv1 and AutoRun;
  - VBS/memory integrity and Secure Boot;
  - whether the daily account is an admin, and the sign-in method.

  Use `ms-settings:` deep links. Checks adapted from HardeningKitty credit it, and `third_party/hardeningkitty/LICENSE` is added. No CIS-derived content.
- [ ] Add portable detection on Windows (`Win32_Battery`) and the Windows check tests with fixtures, including the non-admin cases.
- [ ] Implement `hardenme checklist ios`:
  - the interactive EN/ES manual checklist from the research, weighted by profile;
  - optional USB reads via an external `ideviceinfo` or go-ios: version freshness, passcode set, backup encryption, supervision, device name, provisioning profiles;
  - an `idevicepair unpair` offer at the end.

  Add `checks/ios_test`.
- [ ] Add GitHub Actions: the verification command on macOS, Windows and Linux runners, plus JSON Schema validation of the checks.
- [ ] Add GoReleaser: macOS universal and Windows binaries, checksums and a Homebrew tap. Write the install instructions in the READMEs.
- [ ] Add issue templates (new check, check correction, false positive) and a "good first issue" list.
- [ ] Run `hardenme scan` on a Windows machine or VM (ask the user) and attach the output to the review.
- [ ] Verify the changes in terms of typechecking, linting and tests using the project's verification command (`gofmt -l . && go vet ./... && go test ./...`). Fix issues if any.
- [ ] STOP. Present the changes to the user for review and suggest commit messages (or pull request titles, when the phases are implemented through pull requests). Do NOT proceed to the next phase until the user explicitly asks.

## ⏭️ Next step

Implement Phase 1: the repository foundation and the macOS audit with the profile-weighted score.

Hardened, scored and theft-ready, brought to you by 🐢 💨 (Turbotuga™, [Codely](https://codely.com)'s mascot).
