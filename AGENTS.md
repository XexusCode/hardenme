# AGENTS.md

How to work on hardenme. This is for people and coding agents alike.

## What this is

hardenme is a single-binary Go CLI for **non-technical users**. It reads a device's security settings and gives a 0-100 score, weighted by how the person uses the device (their profile). For every problem it explains why it matters and how to fix it, in English and Spanish. It also includes a guide on what to do before and after a device theft.

The plan lives in `.agents/plans/`. Prior-art research, the check catalogue for each OS and the licensing rules are in `docs/research/prior-art.md`.

## Verification command

```bash
gofmt -l . && go vet ./... && go test ./...
```

`gofmt -l .` must print nothing.

## Layout

| Path | What it holds |
|---|---|
| `main.go` | Entry point |
| `internal/cli` | Cobra commands (`scan`, `fix`, `profile`, `guide`) and their EN/ES prompts |
| `internal/runner` | `Runner` interface: `Exec` runs real commands, `Fake` replays captured outputs in tests |
| `internal/profile` | `Profile`, its JSON `Store`, and laptop detection |
| `internal/check` | Check model, catalogue loading and validation, probe execution, Go builtins (`builtins_<platform>.go`) |
| `internal/scan` | Runs the checks for a profile, then computes the weighted score and the exit code |
| `internal/fix` | Fix candidates, previous-value capture, the one-dialog admin batch (`osascript`), re-check, change log and catalog-checked undo |
| `internal/render` | Report output (text for now) with EN/ES labels; golden files in `testdata/` |
| `checks/` | **Check definitions as data**: `checks/<platform>/<name>.yaml`, plus `schema.json`. Embedded with `go:embed` |
| `docs/guide/<lang>/` | Guide pages, embedded and printed by `hardenme guide` |
| `third_party/<project>/LICENSE` | License texts of projects we adapted code or data from |

## Rules

1. **Checks are data.** Add a check by writing `checks/<platform>/<name>.yaml`. Its `id` must be `<platform>.<name>`.
   - Use a `command` probe with regex/exit-code `rules` whenever possible.
   - Write a Go `builtin` only when the logic needs thresholds that depend on the profile, or needs several reads.
   - Use `manual: true` when the setting cannot be read.
2. **Never require sudo to read, never read TCC-protected data, never use the network.** `hardenme scan` never changes settings. Only `hardenme fix` changes them, after the user picks and confirms, through the embedded `fix` blocks (see CONTRIBUTING, "Rules for automatic fixes"), with one native password dialog and an undo log.
3. **Read effective state, not policy keys alone.** A missing key must map to the OS default, never silently to "pass".
4. **A probe that cannot run is `unknown`, never `pass`.** `unknown`, `manual` and `skipped` do not count towards the score.
5. **Every check needs:**
   - EN and ES text: `title` (neutral), `title_pass`, `title_fail`, `why`, `fix_steps`;
   - at least one `references` URL (official docs);
   - `credits` when adapted from another project.
6. **Every check needs tests:** pass, fail and unknown cases, using `runner.Fake`. Label outputs captured on a real machine with the OS version (e.g. `"off 26.2"`). Never invent an output and present it as captured.
7. **User-facing text** is plain language for non-technical people: say why it matters before how to fix it. In ES use natural Spain Spanish. Do not use the "—" or "–" characters; a test enforces this for check texts.
8. **Licensing.** You may adapt MIT, Apache-2.0 and CC BY 4.0 sources with attribution. Never copy:
   - Lynis code (GPL);
   - anything under Kolide's `ee/` directory;
   - CIS Benchmark text or IDs;
   - Apple "Vendor Description" text from mSCP;
   - MVT code.

   See `docs/research/prior-art.md` §5 and `CREDITS.md`.

## Scoring

`score = Σ(credit × w) / Σ(w) × 100` over scored checks. The weight w comes from the check's relevance for the profile: essential 3, recommended 2, optional 1, not applicable 0.

A check's relevance is the highest level among its `relevance.default` and every `relevance.rules` entry whose `when` matches the profile.

Credit is 1 for pass, 0 for fail and 0.5 for warn (a rule can override this with `credit`).

Grades: A ≥ 90, B ≥ 80, C ≥ 70, D ≥ 60, F < 60.

Exit codes:
- 1: an essential or recommended check failed;
- 2: only warnings or optional failures;
- 0: otherwise.

## Commits

Conventional Commits, e.g. `feat(macos): add firewall stealth mode check`, `docs(guide): ...`, `test(render): ...`.
