---
name: "hardenme mobile"
description: "Guided iOS and Android checklist in the CLI with a profile-weighted score, theft stories that join failures into concrete risks, and JSON/Markdown output to share the report."
created_at: "2026-11-08T10:00:00Z"

created_by:
  tool: "Claude Code"
  model:
    name: "Claude Opus"
    version: "5.5"
    reasoning_effort: "low"

has_completed_all_phases: "false"
---

# hardenme mobile

## 🎯 Goal

A non-technical person checks their phone's security with hardenme, without cables, adb or developer options. They get:
- a score weighted by how they use the phone;
- for each problem, why it matters and where the setting is, in their language;
- plain-language "theft stories" that say what a thief could do today;
- a report they can save or share.

This is v0.3 in the [roadmap](../2026_11_06-hardenme_security_auditor/roadmap.md). It depends on nothing from v0.2 except the shared check engine.

## 👀 Context

- Read [`AGENTS.md`](../../../AGENTS.md), the first plan and the [fix plan](../2026_11_07-hardenme_fix/2026_11_07-hardenme_fix-plan.md).
- The research already lists the iOS manual checklist with EN/ES paths and profile rules: `docs/research/prior-art.md` §3, iOS. Only "Stolen Device Protection" has a Spanish path checked against Apple ES. The Android list in §3 is for adb, but the settings map to questions.
- The audit engine already has what a checklist needs:
  - checks as YAML;
  - relevance by profile;
  - `manual`, `pass` and `fail` statuses;
  - the score;
  - the text report.
  The only new thing is a probe whose state comes from the user's answer.
- Menu paths on Android differ by manufacturer. Texts give the stock Android / Pixel path and say "on Samsung and others the name may change; search for it in Settings".
- References must be official: Apple Support (`support.apple.com`, ES pages at `support.apple.com/es-es/...`) and Google Help (`support.google.com/android`, `support.google.com/accounts`). Every URL is opened once before it goes into a YAML.

## 📜 Public contracts

### CLI
- `hardenme checklist ios` and `hardenme checklist android`:
  - first-run profile questions for the phone (it is always portable, and it is assumed to leave home);
  - one yes / no / "I don't know" question per check that applies to the profile;
  - then the same report as `scan`, with the platform name.
- Answers are saved, so the next run asks again only what was not good or not known. `--again` asks everything.
- `--only <ids>` as in `scan`. `--lang` as everywhere.
- The `scan --format` flag accepts `text`, `json` and `markdown`. `checklist` accepts the same flag.
- Exit codes are as in `scan`.

### Check data
- A new probe kind `ask: true`, exclusive with `command`, `builtin` and `manual`.
- `text.<lang>.question` is required when `ask` is set. It is a yes/no question where "yes" is the safe answer, for example "Is Stolen Device Protection on?".
- An `ask` check may add `text.<lang>.where`: the menu path, shown with the question so the person can look it up before answering.
- The schema and `catalog_test` enforce:
  - `question` with `ask`, and no `question` without it;
  - `platform` is one of `macos`, `ios` or `android`;
  - `ios` and `android` checks use only `ask` or `manual` (there is nothing to read on a phone from the CLI).

### Packages
- `check.Env` gains `Ask func(Check) Answer`. `Answer` is `Yes`, `No`, `DontKnow` or `Quit`. The status mapping:
  - `Yes` → pass;
  - `No` → fail;
  - `DontKnow` → manual, which does not count towards the score;
  - `Quit` → skipped, and the rest are skipped too.
  - A nil `Ask` → manual, so `scan` never blocks.
- `internal/answers`: the per-platform answer store at `~/.local/state/hardenme/answers-<platform>.json`, mode 0600. It holds `{check_id: "yes"|"no"|"dont_know", time}`. A saved `yes` is reused; anything else is asked again.
- `profile`: a phone profile stored in `profile-phone.json`, next to the Mac one. The three questions are shared, sensitive data and risk. `Portable` and `LeavesHome` are true.
- `internal/story` (phase 3): `Stories(report) []Story`. A story is data in `checks/stories/<platform>.yaml`: `id`, `when_failing: [check ids]` (all must fail or warn), and EN/ES `text`. They are rendered at the top of the report as "If someone steals your phone today…".
- `internal/render`: `json` (a stable schema with `schema_version: 1`) and `markdown` (the same sections as text, as headings and lists, with no ANSI codes).

### Test suites
- `check`: the `ask` probe, covering all four answers, a nil `Ask` and the schema rules.
- `answers`: round trip, mode 0600, and reuse only of `yes`.
- `cli`: checklist end to end with scripted stdin, covering the first run, the second run asking only the failed checks, `--again` and quitting halfway.
- `story`: a story fires only when all its checks fail or warn; not applicable or unknown checks never fire one.
- `render`: golden files for JSON and Markdown in EN/ES, and JSON that round-trips through `encoding/json`.
- `catalog_test`: every iOS/Android check has EN/ES `question`, `title`, `why`, `fix_steps` and an official reference.

### Text shown to end users (ES and EN, plain language)
- "Vamos a revisar tu iPhone. Te haré unas preguntas; si no sabes una, contesta «no sé» y te digo dónde mirarlo."
- "¿Está activada la Protección en caso de robo del dispositivo? (s/n/no sé)"
- "Dónde: Ajustes > Face ID y código > Protección en caso de robo del dispositivo."
- "Si te roban el iPhone hoy, quien lo tenga podría apagar Buscar desde el centro de control y quedarse con él."

## 🪜 Phases

### Phase 1: Checklist engine and the iOS checklist

**Description:** Add the `ask` probe, the answer store, the phone profile and `hardenme checklist ios` with the iOS checks from the research.

- [ ] `ask` probe, `question` and `where` text, schema rules and `check.Env.Ask`, with tests.
- [ ] `internal/answers` with tests.
- [ ] Phone profile (questions and store) with tests.
- [ ] `hardenme checklist <platform>` (`--again`, `--only`, `--lang`) with EN/ES prompts and CLI tests.
- [ ] iOS checks as YAML (`checks/ios/`) from the research table, with official Apple references (EN and ES pages) opened once each:
  - Stolen Device Protection;
  - strong passcode;
  - Face ID / Touch ID;
  - USB accessories when locked;
  - Control Center on the lock screen;
  - notification previews;
  - SIM PIN;
  - Find My and Send Last Location;
  - erase after 10 attempts;
  - Advanced Data Protection;
  - Lockdown Mode;
  - Safety Check;
  - automatic updates;
  - two-factor authentication and recovery.
- [ ] Mark every Spanish path not checked against Apple ES as unverified in the plan notes, never in user text.
- [ ] Verify with `gofmt -l . && go vet ./... && go test ./...`.
- [ ] STOP. Show the questions in EN/ES to the user before Phase 2.

### Phase 2: The Android checklist

**Description:** Write the Android checks as `ask` YAML and check the paths against Google's help pages.

- [ ] Android checks (`checks/android/`):
  - screen lock;
  - auto-lock time;
  - lock-screen notifications;
  - Find My Device;
  - Theft Detection Lock and Offline Device Lock;
  - Identity Check;
  - security updates (patch level in the last 3 months);
  - Play Protect;
  - unknown-source installs;
  - Smart Lock / Extend Unlock;
  - SIM PIN;
  - Google account 2-Step Verification.
- [ ] Stock / Pixel paths, plus the manufacturer note. References from `support.google.com`.
- [ ] Tests and verification command.
- [ ] STOP. Show the questions to the user.

### Phase 3: Theft stories

**Description:** Join failing checks into plain-language risks, for the Mac and the phones, at the top of the report.

- [ ] `checks/stories/<platform>.yaml` and its schema. `internal/story` with tests.
- [ ] Two to four stories per platform, for example:
  - Mac: FileVault off and Find My unknown → "could read your files and you could not erase it".
  - iPhone: Control Center on the lock screen and Stolen Device Protection off → "could turn off Find My and change your Apple ID password with your passcode".
- [ ] Render the stories in text (and later in JSON/Markdown). Golden files.
- [ ] STOP.

### Phase 4: JSON and Markdown output

**Description:** Let the person save or share the report, and let other tools read it.

- [ ] `render.JSON` with `schema_version: 1`: platform, OS version, profile, score, grade, stories and entries (id, title, status, relevance, current, fix kind). No paths from the home directory and no device name, so sharing the report is safe.
- [ ] `render.Markdown`.
- [ ] Golden files and a round-trip test. README usage.
- [ ] STOP.

### Phase 5: Check on real devices and release

**Description:** Only with the maintainer's go-ahead and devices.

- [ ] Walk the iOS checklist on an iPhone in ES and EN, and the Android one on an Android phone. Fix wrong paths.
- [ ] Review the guide's phone menus (`docs/guide`) at the same time.
- [ ] README (EN/ES): the checklist section. Roadmap: tick v0.3.
- [ ] Release v0.3.0 only when the user asks.

**Done when:** someone checks their phone's security with hardenme, understands what a thief could do today, and knows what to change.

## ⏭️ Next step

Phase 1: the `ask` probe and the iOS checklist.
