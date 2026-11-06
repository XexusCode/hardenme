# hardenme: prior-art research

Research date: 2026-11-06. Repositories were cloned read-only. None of their code was executed.

**Status of the facts here.** Items marked _unverified_ were taken from the repos or from general knowledge. They have not been tested on a real device yet. Each one must be confirmed before it ships as a check.

## 1. Summary

| Project | Platform | What it is | License | Stars | Last commit | Reuse in hardenme |
|---|---|---|---|---|---|---|
| [Lissy93/personal-security-checklist](https://github.com/Lissy93/personal-security-checklist) | Generic guide | 300+ tips in YAML, a web app ([digital-defense.io](https://digital-defense.io)) | **CC BY 4.0** in the LICENSE file. The README says MIT, so we treat it as CC BY 4.0 | ~22.7k | 2026-11-01 | Data model and ideas. Adapted text keeps the CC BY attribution |
| [HardeningKitty](https://github.com/scipag/HardeningKitty) / [windows_hardening](https://github.com/0x6d69636b/windows_hardening) | Windows | PowerShell auditor with CSV finding lists and a 1–6 score | MIT (© 2017 Michael Schneider; © 2020 scip ag) | 1.9k / 2.7k | 2026-07 / 2026-10 | Check definitions, with attribution. Do **not** take content derived from CIS |
| [HARDAX](https://github.com/V33RU/hardax) | Android (adb) | Python auditor, ~850 JSON checks, 0–100 risk score | MIT (© 2026 V33RU) | 14 | 2026-09-28 | adb check definitions, with attribution |
| [fort](https://github.com/djadmin/fort) | macOS | Go CLI, 16 checks, N/16 score, `--fix` | MIT (© 2026 Dheeraj Joshi) | 82 | 2026-07-09 | Check design, fix UX and GoReleaser setup |
| [Lynis](https://github.com/CISOfy/lynis) | Unix/macOS | Shell auditor with a weighted "hardening index" | **GPL-3.0** | 16.6k | 2026-09-16 | **Ideas only**, no code |
| [drduh/macOS-Security-and-Privacy-Guide](https://github.com/drduh/macOS-Security-and-Privacy-Guide) | macOS | Guide | MIT (© 2015) | 22.6k | 2026-11-01 | Source for the guide content (paraphrase it) |
| [NIST mSCP](https://github.com/usnistgov/macos_security) | macOS/iOS | Compliance rules | CC BY 4.0. Apple "Vendor Description" text is excluded | 2.5k | 2026-11-05 | What to check, plus its iOS manual rules |
| [osquery](https://github.com/osquery/osquery) | Desktop | Lets you query the OS with SQL | Apache-2.0 OR GPL-2.0 | — | — | Reference only, not a dependency |
| [Kolide launcher](https://github.com/kolide/launcher) | Desktop | Device-trust agent | MIT **except `ee/`, which is proprietary** | — | — | Ideas only. Do **not** port `ee/tables` |
| [Netflix Stethoscope](https://github.com/Netflix-Skunkworks/stethoscope-app) | Desktop | Device-health app (inactive since 2023) | Apache-2.0 | — | 2023 | Remediation YAML schema and settings deep links |
| [libimobiledevice](https://github.com/libimobiledevice/libimobiledevice) | iOS | USB protocol tools | LGPL-2.1+ | — | — | Optional external binary only. Never link it |
| [go-ios](https://github.com/danielpaulus/go-ios) | iOS | Pure-Go iOS USB library | MIT | — | — | Candidate native iOS transport |
| [MVT](https://github.com/mvt-project/mvt) | iOS/Android | Forensic spyware check (Amnesty) | MVT License 1.1, with a consent clause | — | — | Recommend it to elevated-risk users. Do not bundle it |
| iVerify, Vigil Prism | — | Commercial | Proprietary | — | — | UX ideas only |

**Gap we fill.** None of these tools combines: cross-platform coverage, an audience of non-technical users, a score weighted by the user's profile, a before/after-theft playbook, and Spanish.

## 2. Design lessons

### Check data model
We merge the best parts of Lissy93, HARDAX, HardeningKitty and Stethoscope. Every check is declarative data, embedded with `go:embed` and validated against a JSON Schema in CI:

- `id`: stable and explicit. Lissy93 derives IDs from titles, so renaming an item breaks saved progress.
- `platform`, `category` (theft, accounts, updates, network, privacy), `severity`.
- `relevance` rules per profile: essential, recommended, optional or n/a.
- `probe`: the command, the parse rule, the pass condition and a `default_if_missing` value. The default idea comes from HardeningKitty's DefaultValue, but it is only applied to keys we know have a default.
- `min_os` / `max_os` gating. HARDAX gates checks by SDK level.
- Text per locale (en/es), following the Stethoscope and Kolide pattern:
  - a title that changes on pass and fail ("Your firewall is on" / "Turn on your firewall");
  - **why** it matters, in plain language;
  - **how** to fix it: numbered steps with a deep link to the setting;
  - an optional copy-paste command.
- `references[]`: required. Lissy93 only asks for references in its PR rules; we make it a schema field.
- `credits`: the source project for each adapted check.

### Statuses
`pass | fail | warn | unknown | manual | skipped(needs-admin)`. A probe that cannot run never counts as a pass. HARDAX calls this VERIFY; HardeningKitty excludes such checks from the total. `unknown`, `manual` and `skipped` are left out of the score.

### Scoring
- Lynis-style weighted points, with our profile as the multiplier: `score = Σ(earned × w) / Σ(max × w) × 100`, where w is 3 for essential, 2 for recommended, 1 for optional and 0 for n/a.
- Partial credit is allowed. For example, an Android patch 4–6 months old scores 50%.
- Grades A–F.
- Avoid HARDAX's count-based hazard formula: it ignores the profile and punishes devices for having many minor checks.
- Avoid HardeningKitty's equal weighting across hundreds of checks: one BitLocker failure gets drowned out.
- Borrow HardeningKitty's idea of two label sets (casual and professional) for the rating text.
- Borrow HARDAX's attack-chain correlation and turn it into **theft stories**. Example: no screen lock + notifications shown + no Find My → "Someone who steals this phone could read your 2FA codes and you couldn't erase it."

### Runtime rules
- **The audit never uses sudo.** Fixes appear as a settings deep link (`x-apple.systempreferences:…`, `ms-settings:…`) or as a command the user copies and runs.
- Never read TCC-protected stores on macOS.
- Wrap command execution in an injectable `Runner`. fort has no such seam, so its tests only run on a real Mac. We use fixture outputs for each OS version.
- Read **effective state**, never policy keys alone. HardeningKitty reads `Policies\…` registry keys, which are absent on home PCs, so it gives false passes.
- An explicit `--fix` comes later. It lists the exact changes, needs confirmation, supports `--dry-run`, and re-checks after applying (fort's model). Never apply everything at once, as HardeningKitty's HailMary mode does.
- iOS: after reading, offer `idevicepair unpair`.
- Never send data anywhere. HARDAX has an opt-in cloud LLM; we have nothing like it.

### Report UX
- One consolidated report: a count of issues first, then the fixes ordered by weight.
- A "fixed it? run again" flow with resolved items struck through (Kolide).
- Progress per category (iVerify, Lissy93 radar chart).
- An evidence transcript of the commands that were run (fort), for the JSON/HTML output.

## 3. Check catalogue

### macOS (phase 1). Sources: fort, mSCP, drduh, Lynis
| Check | Probe (no sudo) | Pass | Notes |
|---|---|---|---|
| FileVault | `fdesetup status` | "FileVault is On." (also treat "Encryption in progress" as on) | |
| Firewall | `/usr/libexec/ApplicationFirewall/socketfilterfw --getglobalstate` | enabled | |
| Firewall stealth | `socketfilterfw --getstealthmode` | on | optional |
| Gatekeeper | `spctl --status` | assessments enabled | |
| SIP | `csrutil status` | enabled | |
| Security data (XProtect) | `defaults read /Library/Preferences/com.apple.SoftwareUpdate ConfigDataInstall` / `CriticalUpdateInstall` | 1 or missing | missing = default on (_unverified on 14+, where updates are managed by DDM_) |
| Automatic updates | same plist: `AutomaticCheckEnabled`, `AutomaticDownload`, `AutomaticallyInstallMacOSUpdates` | 1 | |
| Pending updates | `PendingUpdateCount` / `LastFullSuccessfulDate` | 0 / within 30 days | avoid `softwareupdate -l` (slow, uses the network) |
| Password right after sleep | `sysadminctl -screenLock status` (prints to stderr) | immediate | _unverified_. fort uses the deprecated `askForPassword` keys |
| Auto-lock timeout | `defaults -currentHost read com.apple.screensaver idleTime`, `pmset -g` | ≤ 300 s (scaled by profile) | |
| Find My Mac | `nvram fmm-mobileme-token-FMM` | key exists | _unverified_ |
| Remote Login (SSH) | `launchctl print-disabled system` + `launchctl print system/com.openssh.sshd` | disabled | fort's `launchctl list` checks the user domain, so it misses system daemons |
| Screen sharing / file sharing | `launchctl print-disabled system` (com.apple.screensharing, com.apple.smbd) | disabled | |
| Remote management | `pgrep -x ARDAgent` | not running | |
| Auto login | `defaults read /Library/Preferences/com.apple.loginwindow autoLoginUser` | missing | |
| Guest account | same plist, `GuestEnabled` | 0 or missing | |
| Daily account is admin | `id` | not admin | warn only |
| AirDrop | `defaults read com.apple.sharingd DiscoverableMode` | Off / Contacts Only | |
| Lockdown Mode | `sysctl -n security.mac.lockdown_mode_state` | 1 | elevated-risk profile only |
| Password manager | known `.app` bundles in `/Applications` | found | fort idea. Soft check (warn) |
| iCloud Advanced Data Protection | — | manual | can't be read locally |

### Android (phase 2, adb, no root). Sources: HARDAX (MIT)
| Check | Probe | Pass | Notes |
|---|---|---|---|
| Patch age | `getprop ro.build.version.security_patch`, compared with today | ≤ 3 months pass, 4–6 warn | adapt HARDAX's logic. A future date means a spoofed value |
| Android version | `getprop ro.build.version.release` / `.sdk` | supported | |
| Encryption | `getprop ro.crypto.state` / `ro.crypto.type` | encrypted / file | |
| Screen lock | `dumpsys lock_settings`, `locksettings get-disabled` | credential set, not disabled | **don't** use `lockscreen.password_type` (broken in HARDAX) |
| Auto-lock | `settings get system screen_off_timeout` + `secure lock_screen_lock_after_timeout` | ≤ 30 s–2 min by profile | |
| Lock-screen notifications | `settings get secure lock_screen_show_notifications` / `lock_screen_allow_private_notifications` | 0 or private | |
| Trust agents / Smart Lock | `settings get secure enabled_trust_agents`, `dumpsys trust` | none | |
| Unknown-source installs | `appops query-op REQUEST_INSTALL_PACKAGES allow` | no unexpected apps | |
| Play Protect | `settings get global package_verifier_enable` | 1 | null → unknown |
| Wireless debugging | `settings get global adb_wifi_enabled` | 0 | |
| USB debugging | — | reminder to turn it off afterwards | it is necessarily on during the audit, so it can't fail |
| Accessibility / notification listeners / device admins | `settings get secure enabled_accessibility_services`, `enabled_notification_listeners`, `dumpsys device_policy` | list them by name for review | stalkerware signal |
| Bootloader / verified boot | `getprop ro.boot.flash.locked`, `ro.boot.verifiedbootstate`, `sys.oem_unlock_allowed` | 1 / green / 0 | |
| Root / Magisk, user CA certs | per HARDAX | none | |
| Find My Device | `settings get secure find_my_device` | 1 | _unverified_; fall back to manual |
| Theft Detection / Identity Check / Offline lock | `settings get secure theft_detection_lock_enabled`, `identity_check_enabled` | 1 | _unverified_. Treat as manual with a best-effort probe |
| SIM PIN | — | manual | no non-root method |

### Windows (phase 3). Sources: HardeningKitty (MIT). Read effective state
| Check | Probe | Pass | Admin? |
|---|---|---|---|
| BitLocker | `Get-BitLockerVolume` (`VolumeStatus` = FullyEncrypted); fallback: the Shell.Application `System.Volume.BitLockerProtection` property | encrypted | Get-BitLockerVolume: yes. The fallback is _unverified_ |
| Defender real-time protection | `Get-MpComputerStatus` (`RealTimeProtectionEnabled`, `AntivirusEnabled`) | True | no |
| Tamper protection | `Get-MpComputerStatus` (`IsTamperProtected`) | True | no |
| PUA protection | `Get-MpPreference` (`PUAProtection`) | ≥ 1 | _unverified_ |
| Firewall per profile | `Get-NetFirewallProfile`, or `HKLM\SYSTEM\CurrentControlSet\Services\SharedAccess\Parameters\FirewallPolicy\{Standard,Public,Domain}Profile\EnableFirewall` | 1 | no |
| UAC | `HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System` `EnableLUA`=1, `ConsentPromptBehaviorAdmin` ≠ 0, `PromptOnSecureDesktop`=1 | | no |
| SmartScreen | effective setting rather than the policy key | on | _unverified_ |
| Updates | policy `NoAutoUpdate` ≠ 1, plus last installed update ≤ 30 days (`Get-HotFix` / WU history) | | no |
| Screen lock timeout | `InactivityTimeoutSecs` or the screensaver keys (`<=!0` 900) | ≤ 900 s, not 0 | no |
| Remote Desktop | `HKLM\SYSTEM\CurrentControlSet\Control\Terminal Server\fDenyTSConnections` | 1 | no |
| SMBv1 | `HKLM\SYSTEM\CurrentControlSet\Services\LanmanServer\Parameters\SMB1`=0 / `MrxSmb10\Start`=4 | disabled | no (the optional-feature probe needs admin) |
| AutoRun | `NoDriveTypeAutoRun`=255 | | no |
| VBS / memory integrity | `Win32_DeviceGuard` (`SecurityServicesRunning` contains 2) | | no |
| Secure Boot | `HKLM\SYSTEM\CurrentControlSet\Control\SecureBoot\State\UEFISecureBootEnabled` | 1 | no |
| Daily account is admin | `whoami /groups` | not in Administrators | no. **HardeningKitty has no such check** |
| Sign-in method | Windows Hello / PIN set | yes | _unverified_ |

### iOS (phase 3)
**Readable over USB after pairing** with libimobiledevice (an external binary) or go-ios:
- `ProductVersion` (is the phone up to date?)
- `PasswordProtected`
- `com.apple.mobile.backup WillEncrypt`
- `com.apple.mobile.chaperone` (supervised?)
- `DeviceName` (does it leak the owner's name?)
- `ideviceprovision list` (unexpected profiles)

**Manual checklist.** Spanish paths: only Stolen Device Protection was checked against Apple ES; check the rest on a device.

| Setting | EN / ES path | Profile |
|---|---|---|
| Stolen Device Protection (iOS 17.3+) | Settings > Face ID & Passcode / Ajustes > Face ID y código > Protección en caso de robo del dispositivo | everyone, critical if it leaves home |
| Strong passcode (6+ digits or alphanumeric) | Face ID & Passcode > Change Passcode / Cambiar código | everyone; alphanumeric for sensitive data or elevated risk |
| Face ID | Face ID & Passcode | everyone who goes out |
| USB Accessories off when locked | Allow Access When Locked / Permitir acceso al estar bloqueado | sensitive data, elevated risk |
| Control Center off on lock screen | same section | everyone who goes out (blocks Airplane Mode → Find My off) |
| Notification previews only when unlocked | Notifications > Show Previews / Notificaciones > Mostrar previsualizaciones | everyone |
| SIM PIN | Cellular > SIM PIN / Datos móviles > PIN de la SIM | everyone who goes out |
| Find My + Send Last Location | [Name] > Find My / [nombre] > Buscar | everyone |
| Erase data after 10 attempts | Face ID & Passcode / Borrar datos | sensitive or elevated; not on a shared device or one used by kids |
| Advanced Data Protection | iCloud / Protección de datos avanzada | sensitive data, elevated risk |
| Lockdown Mode | Privacy & Security / Modo de aislamiento | elevated risk only |
| Safety Check | Privacy & Security / Comprobación de seguridad | shared device, abusive-partner risk |
| Automatic updates | General > Software Update | everyone |
| Apple ID 2FA + recovery contact/key | Sign-In & Security | everyone |

## 4. Theft playbook gaps
None of the projects reviewed covers what to do **after** a theft. It is our own content, so we write it from scratch:
- Mark the device as lost or erase it (Find My / Find My Device).
- Call the carrier to block the SIM and get a new one.
- Report it to the police with the IMEI. We need a "write down your IMEI now" step *before* any theft happens.
- Sign out of all sessions and revoke tokens: Google, Apple ID, Microsoft, WhatsApp, banking.
- Change passwords, starting with email.
- Warn your contacts about phishing sent from your number.
- Watch your bank accounts.

## 5. Licensing rules for contributors
- MIT sources (fort, HARDAX, HardeningKitty, drduh, go-ios, Kolide outside `ee/`): may be adapted. Keep their full license text in `third_party/<project>/LICENSE` and record the source commit.
- CC BY 4.0 sources (Lissy93, mSCP): attribute, link the license and state the changes. We prefer rewording to copying.
- Apache-2.0 (Stethoscope, osquery): attribute and keep the NOTICE file if we adapt anything.
- **Never copy:** Lynis code (GPL-3.0), Kolide `ee/` (proprietary), CIS Benchmark text or IDs (_believed to be CC BY-NC-SA; unverified_), Apple "Vendor Description" text in mSCP, Kolide's "Honest Security" text, MVT code.
- libimobiledevice (LGPL): call it as an external binary only.
- Commands and registry paths are facts, so they are not copyrightable. We still credit whoever curated them.
