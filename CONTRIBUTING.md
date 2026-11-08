# Contributing to hardenme

> **¿Hablas español?** Puedes abrir issues, pull requests y comentarios en español: son igual de bienvenidos. Todas las comprobaciones necesitan texto en inglés y en español, así que nos viene muy bien gente que escriba con naturalidad en español.

Thank you for helping make devices safer for people who are not security experts. This document explains how the project is organised and the rules every contribution must follow.

## Ways to help

- **Add or improve a check** (the most valuable contribution).
- **Fix wrong or outdated text**, for example a settings path that moved in a new OS version.
- **Report a false positive or false negative**, ideally with the command output from your machine.
- **Improve the guides** in `docs/guide/en` and `docs/guide/es`.
- **Translate or polish Spanish** so it reads naturally.

For anything bigger than a small fix, please open an issue first so we can agree on the approach.

## How checks work

Checks are **data, not code**. Each one is a YAML file in `checks/<platform>/`, one file per check, embedded in the binary and validated against `checks/schema.json`.

```text
checks/
  schema.json
  macos/
    filevault.yaml
    firewall.yaml
    ...
```

Required fields:

| Field | Purpose |
|---|---|
| `id` | Stable, explicit identifier. **Never rename it** once released: users filter and track by it. |
| `platform` | `macos`, `android`, `windows` or `ios`. |
| `category` | For example `theft`, `accounts`, `updates`, `network`, `privacy`. |
| `severity` | How serious a failure is. |
| `relevance` | Rules that turn the user's profile into `essential`, `recommended`, `optional` or `not_applicable`. |
| `probe` | The command to run, how to parse it, the pass condition, and a default only where the OS really has one. |
| `text.en`, `text.es` | `title_pass`, `title_fail`, `why`, `fix_steps[]`, and optional `deep_link` and `command`. |
| `references` | **Required.** At least one link to official documentation (Apple, Google, Microsoft) that backs the check. |

Optional fields include `min_os` / `max_os` and `credits` (see below).

### Writing the text

Remember who reads it: someone who has never opened Terminal.

- `title_pass` / `title_fail` change with the result: "Your firewall is on" / "Turn on your firewall".
- `why` explains the real-world risk in one or two sentences, without jargon. Theft scenarios work well.
- `fix_steps` are numbered, concrete steps that name the exact menu path. Add a `deep_link` (`x-apple.systempreferences:...`, `ms-settings:...`) when one exists.
- The Spanish text must be natural Spain-Spanish, not a word-for-word translation. Use the real Spanish names of menus as they appear on the device.

## Rules for probes

1. **Never require sudo or admin rights to read.** If a setting cannot be read without them, the check returns `skipped` or `manual`.
2. **Read the effective state, not policy keys.** A policy key that is absent on a home computer does not mean the feature is on. Read what the system actually does.
3. **Never read privacy-protected data** (for example TCC-protected stores on macOS) and **never use the network**.
4. **A probe that cannot run is never a pass.** Command errors and unexpected output must produce `unknown`.
5. **Run every command through the `Runner` interface**, so tests can replace it with fixtures.
6. Avoid slow commands. A full scan should finish in a few seconds.

## Rules for automatic fixes

A check may add a `fix` block so `hardenme fix` can apply it. This changes the user's device, so the bar is higher:

1. **Verify it on a real Mac first**: apply, re-run the probe, undo, re-run the probe. Write what you saw in the pull request. A fix that cannot be verified stays guided (only `deep_link` and `fix_steps`).
2. **Every step has an `undo`.** When the previous value varies, capture it with `read` (and `read_match`, `previous_map`) and use `{previous}` in `undo`; use `undo_if_missing` when the setting may not exist.
3. **Commands are argv arrays with absolute paths.** No shell, no `$`, `;`, `|`, `&`, `<`, `>` or backslashes. The tests enforce this.
4. **Mark `admin: true` only when needed.** User-level settings (for example `defaults -currentHost`) must run as the user, not as root.
5. **Add a `fix_warning`** in EN and ES when the change has a visible side effect (restarts, a service someone may use).
6. **Never automate a fix that can lock the user out or lose data**, such as turning on FileVault (the recovery key must be saved by the person).

## Tests and fixtures

Every check needs table-driven tests with captured command output covering at least:

- **pass**
- **fail**
- **unknown** (command error or unexpected output)

Put fixtures next to the platform tests, one per OS version you have verified (for example macOS 14, 15 and 26). If you captured output on your own machine, remove anything personal (user names, serial numbers, hostnames) before committing.

If a probe is marked _unverified_ in `docs/research/prior-art.md`, confirm it on a real device and say so in the pull request.

## Verification

Before opening a pull request, run:

```sh
gofmt -l . && go vet ./... && go test ./...
```

`gofmt -l .` must print nothing, and all tests must pass.

## Licensing and attribution

hardenme is MIT. We learn from many projects, but we only reuse content when its license allows it. These rules are not optional.

### What you may adapt

- **MIT sources** (fort, HARDAX, HardeningKitty, drduh's macOS guide, go-ios, Kolide launcher **outside `ee/`**): may be adapted. Keep their full license text in `third_party/<project>/LICENSE` and record the source commit in your pull request.
- **CC BY 4.0 sources** (Lissy93's Personal Security Checklist, NIST mSCP): attribute, link the license and state what you changed. Prefer rewording to copying.
- **Apache-2.0 sources** (Netflix Stethoscope, osquery): attribute, and keep their NOTICE file if you adapt anything.
- **Commands, settings keys and registry paths** are facts, not creative work. You may use them, but still credit whoever curated them.

When a check is adapted from another project, set the `credits` field:

```yaml
credits:
  - "fort (MIT): https://github.com/djadmin/fort"
```

and make sure the project is listed in [CREDITS.md](CREDITS.md).

### What you must never copy

- **Lynis code** (GPL-3.0). Ideas only.
- **Kolide `ee/` code** (proprietary) and Kolide's "Honest Security" text.
- **CIS Benchmark text or IDs.**
- **Apple "Vendor Description" text** found in NIST mSCP.
- **MVT code** (MVT License). We may recommend the tool, never bundle or copy it.
- **libimobiledevice** (LGPL) may only be called as an external binary, never linked.

If you are unsure whether something can be used, ask in the issue before writing the check.

## Commits and pull requests

We use [Conventional Commits](https://www.conventionalcommits.org/):

```text
feat(macos): add AirDrop discoverability check
fix(scoring): exclude manual results from the total
docs(guide): add Spain-specific steps to theft-after
test(macos): add macOS 26 fixtures for FileVault
```

Keep pull requests focused: one check or one fix per pull request is ideal. Describe how you verified the probe (OS version, device) and link the official references.

## Code of Conduct

By taking part you agree to follow our [Code of Conduct](CODE_OF_CONDUCT.md). To report a vulnerability, see [SECURITY.md](SECURITY.md) instead of opening a public issue.
