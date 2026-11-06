# Credits

hardenme stands on the work of these projects. Thank you to their authors and contributors.

When a check or text is adapted from one of them, the check file records it in a `credits` field. Full license texts for adapted MIT/Apache sources live in `third_party/<project>/LICENSE`.

## Adapted (code, check definitions or data)

- **fort** by Dheeraj Joshi: https://github.com/djadmin/fort. MIT License, Copyright (c) 2026 Dheeraj Joshi. We adapted the macOS check design, the evidence transcript and the fix-confirmation UX, and followed its approach to GoReleaser/Homebrew distribution.
- **HARDAX (Hardening Audit eXaminer)** by V33RU (Mr-IoT / IOTSRG): https://github.com/V33RU/hardax. MIT License, Copyright (c) 2026 V33RU. We adapted Android adb check definitions: security-patch age logic, encryption, verified boot and bootloader state, lock-screen and trust-agent settings, and the Android 15+ theft-protection probes. We were also inspired by its explicit "verify" state and its attack-chain correlation.
- **HardeningKitty** by Michael Schneider (0x6d69636b) and scip AG: https://github.com/scipag/HardeningKitty and https://github.com/0x6d69636b/windows_hardening. MIT License, Copyright (c) 2017 Michael Schneider and Copyright (c) 2020 scip ag. We adapted Windows check definitions and were inspired by its severity-weighted score with casual and professional rating labels. No CIS-derived content is included.
- **Personal Security Checklist (Digital Defense)** by Alicia Sykes (Lissy93) and contributors: https://github.com/Lissy93/personal-security-checklist. Licensed under CC BY 4.0 (https://creativecommons.org/licenses/by/4.0/). Some of our checklist structure and recommendations are inspired by or adapted from it. Changes were made: the content was rewritten, translated to Spanish, re-prioritized per user profile and mapped to automated checks.
- **macOS Security and Privacy Guide** by drduh and contributors: https://github.com/drduh/macOS-Security-and-Privacy-Guide. MIT License, Copyright (c) 2015. It informed our macOS guide content.
- **NIST macOS Security Compliance Project (mSCP)**: https://github.com/usnistgov/macos_security. CC BY 4.0. It informed our macOS check logic and our iOS manual checklist, which were adapted and modified. No Apple "Vendor Description" text is included.
- **Netflix Stethoscope** by Netflix: https://github.com/Netflix-Skunkworks/stethoscope-app. Apache-2.0. We followed its schema for remediation instructions and its system-settings deep links.
- **go-ios** by danielpaulus: https://github.com/danielpaulus/go-ios. MIT License. It is the candidate native transport for iOS device reads.

## Inspiration only (no code or text used)

- **Lynis** by CISOfy: https://github.com/CISOfy/lynis. GPL-3.0. Our weighted scoring was inspired by its hardening index. No Lynis code is included.
- **osquery**: https://github.com/osquery/osquery. Apache-2.0 OR GPL-2.0-only. We consulted it as a reference for where device facts live.
- **Kolide / 1Password Device Trust**: https://github.com/kolide/launcher (MIT outside `ee/`). Its check catalogue, its "why + how to fix" remediation pattern and its re-check flow inspired ours. No `ee/` code and no "Honest Security" text is used.
- **iVerify** (Trail of Bits spin-out) and **Vigil Prism**: commercial products, reviewed only for UX and scope ideas.

## External tools (not bundled)

- **libimobiledevice**: https://libimobiledevice.org. LGPL-2.1+. Optional; it is called as an external binary for iOS reads.
- **MVT (Mobile Verification Toolkit)** by Amnesty International Security Lab: https://github.com/mvt-project/mvt. MVT License 1.1. We recommend it to elevated-risk users who need a forensic check.
