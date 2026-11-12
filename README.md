# hardenme

**English** | [Español](README.es.md)

> Check how safe your device is, in plain language, and get ready for the day it gets stolen.

`hardenme` is a free, open-source command-line tool for **non-technical people**. It looks at your device's security settings, gives you a **score from 0 to 100** (grade A to F), and tells you, for each problem, **why it matters and how to fix it**, in English or Spanish.

The score depends on **your situation**. Disk encryption is essential on a laptop you carry around, but optional on a desktop that never leaves home. `hardenme` asks a few simple questions and weighs every check accordingly.

> **Project status: early development, macOS only.** Android, Windows and iOS are on the [roadmap](#roadmap).

## Why

There are great security tools out there, but none of them combines all of this:

- **Cross-platform**: one tool for your Mac, your Windows laptop and your phone.
- **Made for non-technical people**: no jargon, clear steps, links straight to the right setting.
- **A score weighted by your profile**: what matters for a commuter's laptop is not what matters for a family desktop.
- **Focus on theft**: what to set up before your device is stolen, and what to do in the first minutes, hours and days after.
- **In English and Spanish.**

## Privacy promises

- **The check only reads.** `hardenme scan` never changes anything and never asks for a password.
- **Fixes only with your OK** (from v0.2). `hardenme fix` changes only what you pick, after showing you what will change. When a change needs your password, macOS asks in its own window: hardenme never sees it. Every automatic change can be undone with `hardenme fix --undo`.
- **No network.** Nothing is sent anywhere. No telemetry, no accounts, no cloud.
- **Open source** (MIT), so anyone can check these promises.

## Install

You do not need to know how to program. You will use Terminal **only once**, to paste one line. After that, hardenme is an app you open with a double-click.

1. **Open Terminal.** Press `Command (⌘) + Space`, type `Terminal` and press `Return`. A window with text appears.
2. **Copy this line, paste it into that window and press `Return`:**

   ```sh
   curl -fsSL https://github.com/XexusCode/hardenme/releases/latest/download/install.sh | sh
   ```

3. **Wait a few seconds.** You will see four steps (download, check, install, create the app) and then "Done!". It does not ask for your password and does not install anything outside your user folder or Applications.
4. **A Finder window opens with hardenme selected. Double-click it.** A page opens in your browser with your score, what could happen today and everything that can be improved, each with its reason. One button fixes what hardenme can change for you (macOS asks for your password once). For the rest, it opens the exact setting and ticks it off by itself when you change it.

From then on, open **hardenme** from Applications or Spotlight whenever you want to check again.

> **Is it safe to paste a command?** Only if you trust where it comes from. You can [read the installer](install.sh) first: it is short and commented. It downloads the release from GitHub, verifies its SHA-256 checksum and stops if anything does not match.

**Uninstall:**

```sh
curl -fsSL https://github.com/XexusCode/hardenme/releases/latest/download/install.sh | sh -s -- --uninstall
```

**For developers:** `go install github.com/XexusCode/hardenme@latest`, or clone the repository and run `go run . scan`.

## Usage

```sh
# Tell hardenme about your situation (asks interactively without flags)
hardenme profile
hardenme profile --portable --leaves-home --sensitive-data --risk normal

# Audit this device
hardenme scan
hardenme scan --lang es
hardenme scan --only filevault,firewall

# See your score in the browser and fix things in a couple of clicks (from v0.2)
hardenme ui

# Or do it in Terminal: pick the improvements and let hardenme apply them
hardenme fix
hardenme fix --dry-run      # show the exact commands, change nothing
hardenme fix --undo         # undo the last batch of changes

# Check your phone by answering a few questions (from v0.3)
hardenme checklist ios
hardenme checklist android
hardenme checklist ios --again    # ask everything again, not only what was wrong

# Save or share the report (from v0.3)
hardenme scan --format markdown > my-mac.md
hardenme scan --format json

# Read the theft guides in your terminal
hardenme guide theft-before
hardenme guide theft-after --lang es
```

Your profile is saved at `~/.config/hardenme/profile.json`. The questions are:

| Flag | Question |
|---|---|
| `--portable` | Is it a laptop? (detected automatically when possible) |
| `--leaves-home` | Do you take it outside? |
| `--shared` | Do other people use it? |
| `--sensitive-data` | Does it hold work or banking data? |
| `--risk normal\|elevated` | Is someone likely to target you personally? |

### Sample output

An excerpt from a real run on a MacBook (macOS 26.2):

```text
hardenme · macOS 26.2
Profile: laptop, goes outside, normal risk

Security score: 57/100 (F, At risk)
7 to fix, 1 suggestions, 2 to check yourself

FIX FIRST (essential for you) (5)

✗ Turn on disk encryption (FileVault)
  Now: FileVault is Off.
  Why it matters:
    Without FileVault, anyone who steals your Mac can take the disk out or
    boot it from another drive and read all your files, passwords and photos
    without knowing your password. [...]
  How to fix it:
    1. Open System Settings > Privacy & Security.
    2. Scroll to FileVault and click "Turn On". [...]
  Open: open "x-apple.systempreferences:com.apple.settings.PrivacySecurity.extension"

✗ Install the pending macOS update
  Now: 26.2, pending: macOS Tahoe 26.7.2
  [...]

CHECK IT YOURSELF (2)
? Make sure Find My Mac is on
  [...]

ALREADY GOOD (10)
✓ Gatekeeper checks the apps you open
✓ macOS updates itself automatically
[...]
```

### Exit codes

| Code | Meaning |
|---|---|
| `0` | Everything relevant passes |
| `1` | At least one essential or recommended check failed |
| `2` | Only warnings or optional items |

## How the score works

Each check gets a **weight** from your profile:

| Relevance for you | Weight |
|---|---|
| Essential | 3 |
| Recommended | 2 |
| Optional | 1 |
| Not applicable | 0 |

The score is the points you earned divided by the points you could earn, times 100:

```
score = Σ(earned × weight) / Σ(max × weight) × 100
```

Grades: **A** 90 or more, **B** 80 to 89, **C** 70 to 79, **D** 60 to 69, **F** below 60.

Some checks give partial credit. Checks that could not be read (**unknown**) or that you need to verify yourself (**manual**) do **not** count, neither for nor against you. A check that cannot run is never treated as a pass.

## Guides

- [Before your device is stolen](docs/guide/en/theft-before.md)
- [My device was stolen: what to do now](docs/guide/en/theft-after.md)

## Roadmap

- **v0.1 (now):** macOS audit, profile-weighted score, theft guides in EN/ES, one-line installer with a double-click app.
- **v0.2:** pick the improvements you want and hardenme applies them. You type your password once in the macOS dialog, see your new score, and can undo everything. For the settings only you should change (like FileVault), it opens the right page and checks again when you are done.
- **v0.3:** guided checklist for iPhone and Android, "theft stories" that explain what a thief could do with your current settings, JSON and Markdown output.
- **v0.4:** automatic Android audit over USB (`adb`, no root) for advanced users.
- **Later:** Windows audit, reading iPhone settings over USB, Homebrew tap.

## Contributing

Contributions are very welcome, especially new checks, corrections and translations. Please read [CONTRIBUTING.md](CONTRIBUTING.md) and our [Code of Conduct](CODE_OF_CONDUCT.md). To report a security issue, see [SECURITY.md](SECURITY.md).

## Credits

`hardenme` builds on the work of many open-source projects. See [CREDITS.md](CREDITS.md).

## License

[MIT](LICENSE)
