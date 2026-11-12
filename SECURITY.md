# Security Policy

## Reporting a vulnerability

Please **do not open a public issue** for security problems.

Report them privately through GitHub's **private vulnerability reporting**: go to the repository's **Security** tab and click **Report a vulnerability**. Only the maintainers can see the report.

Please include:

- what you found and why it is a problem;
- steps to reproduce it (OS and version, hardenme version or commit);
- the impact you expect.

We will acknowledge your report as soon as we can, keep you informed while we work on a fix, and credit you in the release notes if you wish. Reports in Spanish are welcome too.

## Supported versions

hardenme is in early development. Only the latest code on the default branch (and, once releases exist, the latest release) receives security fixes.

## Scope

In scope:

- Anything that makes `hardenme scan` **change** settings, files or state on the device, or that makes `hardenme fix` change something the user did not pick and confirm.
- Anything that makes it **send data** over the network or write data outside its own files (the profile in `~/.config/hardenme/profile.json` and the change log in `~/.local/state/hardenme/changes.json`).
- Anything that lets a **web page, another program or another user** talk to the local page server of `hardenme ui`, or make it run anything other than the embedded fixes.
- Anything that makes it **request or use elevated privileges** (sudo, admin) outside a `hardenme fix` the user confirmed, or that lets anything other than the embedded fix commands run as administrator (for example through the change log).
- Command injection or unsafe execution through check data, profile files or command-line arguments.
- Vulnerabilities in the build or release pipeline, or in dependencies as used by hardenme.
- A check that reports a **false pass** on an important protection (for example says disk encryption is on when it is not). Wrong security advice can hurt users, so we treat this seriously. Minor false positives can go to a normal issue.

Out of scope:

- Vulnerabilities in the operating systems or third-party tools that hardenme only reads from or recommends.
- Weaknesses that require an attacker to already control the user's account or machine.
- Disagreements about whether a recommendation should be essential or optional: please open a normal issue for those.

## Our safety principles

These are promises the code must keep. A change that breaks any of them is a security bug.

1. **The audit is read-only.** `hardenme scan` never changes settings. If a setting cannot be read without elevated privileges, the check is marked as skipped or manual.
2. **Fixes only with consent.** `hardenme fix` changes only what the user picks, after showing what will change. It runs only commands embedded in the binary, as argument lists with absolute paths. Admin changes go through one native macOS password dialog; hardenme never asks for, sees or stores the password. Every automatic change is logged with its undo, and `--undo` only runs commands that match the embedded catalog.
3. **No network.** hardenme makes no network connections during an audit. No telemetry, no update checks, no cloud features.
4. **The browser page is local and locked down.** `hardenme ui` listens only on `127.0.0.1`, on a random port, for the time the page is open. Every request needs the secret of that run, and requests with another `Host` or `Origin` are refused. The page loads nothing from the internet, and what it sends is only ever a check ID looked up in the embedded catalog, never a command.
5. **No data collection.** Nothing about you or your device leaves your machine. The only files it writes are your profile and the log of changes made by `hardenme fix`. It never reads privacy-protected data such as macOS TCC-protected stores.
6. **Never a false pass by design.** A check that cannot run is reported as unknown and is never counted as passing.
