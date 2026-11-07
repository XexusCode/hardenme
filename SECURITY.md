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

- Anything that makes hardenme **change** settings, files or state on the device.
- Anything that makes it **send data** over the network or write data outside its own profile file (`~/.config/hardenme/profile.json`).
- Anything that makes it **request or use elevated privileges** (sudo, admin).
- Command injection or unsafe execution through check data, profile files or command-line arguments.
- Vulnerabilities in the build or release pipeline, or in dependencies as used by hardenme.
- A check that reports a **false pass** on an important protection (for example says disk encryption is on when it is not). Wrong security advice can hurt users, so we treat this seriously. Minor false positives can go to a normal issue.

Out of scope:

- Vulnerabilities in the operating systems or third-party tools that hardenme only reads from or recommends.
- Weaknesses that require an attacker to already control the user's account or machine.
- Disagreements about whether a recommendation should be essential or optional: please open a normal issue for those.

## Our safety principles

These are promises the code must keep. A change that breaks any of them is a security bug.

1. **Read-only.** hardenme audits; it never changes settings by itself. Fixes are shown as steps, deep links to the right settings page, or commands that you decide to run yourself.
2. **No sudo, no admin.** The audit never asks for, needs or uses elevated privileges. If a setting cannot be read without them, the check is marked as skipped or manual.
3. **No network.** hardenme makes no network connections during an audit. No telemetry, no update checks, no cloud features.
4. **No data collection.** Nothing about you or your device leaves your machine. The only file it writes is your profile. It never reads privacy-protected data such as macOS TCC-protected stores.
5. **Never a false pass by design.** A check that cannot run is reported as unknown and is never counted as passing.
