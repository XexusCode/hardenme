# Can they unlock my phone?

[Español](../es/phone-unlock.md)

Short answer: **an ordinary thief cannot guess the passcode of an up-to-date phone**. What really happens is something else: they watch you type it, or they take the phone while it is unlocked.

## Guessing the passcode: almost impossible

The phone's own security chip limits the attempts, and that cannot be bypassed from outside.

- **iPhone:** the first three wrong tries have no wait. Then you must wait 1 minute, 5 minutes, 15 minutes, 1 hour, 3 hours and 8 hours. At the tenth wrong try the iPhone locks itself (and erases itself, if you turned on Erase Data).
- **Android:** the waits also grow with each wrong try. On the latest versions they reach hours and days, and after 20 wrong tries no more are accepted.

A 6-digit passcode has a million combinations. With those waits, trying them is not practical. A 4-digit one has ten thousand, and it is also much easier to see at a glance.

## With forensic tools: sometimes, and not a street thief

Some companies sell tools to extract data from phones. What specialist press has reported:

- They work worse the newer and more up to date the phone is. Documents leaked in 2024 showed one of them could not handle a large share of modern iPhones.
- A phone that has just been switched on, and has not been unlocked even once, has all its data encrypted and is nearly impossible to read without the passcode.
- That is why an iPhone that has not been unlocked for 72 hours restarts by itself, as researchers have confirmed.

In practice: **keep your phone up to date**. It is your best defence against this.

## What really happens

### They watch you type the passcode, then steal the phone

In 2023, thefts like this were documented in bars in the United States: someone watches you type your passcode and then takes the phone. With the passcode, the thief changed the Apple Account password, turned off Find My, paid with Apple Pay and got into banking apps with the saved passwords.

**What to do:**

- **iPhone: turn on Stolen Device Protection** (Settings > Face ID & Passcode). Away from familiar places:
  - viewing saved passwords, erasing the iPhone or turning off Lost Mode needs your face or fingerprint, and the passcode no longer works;
  - changing the Apple Account password or the iPhone passcode also waits one hour.
- **Android: turn on Identity Check** if your phone has it (Settings > Google > All services > Theft protection). Outside your trusted places, changing the lock, the Google Account or viewing saved passwords needs your fingerprint or face.
- **Type the passcode in public as little as possible.** Use your fingerprint or face, and if you must type, cover it like a PIN at a cash machine.
- **Use 6 digits or more.**

### They snatch it from your hand while unlocked

The thief has an open phone and only needs it not to lock.

- **Android: turn on Theft Detection Lock.** It detects the snatch and the getaway and locks the screen.
- **A short auto-lock** (30 seconds or 1 minute) on any phone.
- If it happens, lock it remotely right away: icloud.com/find, android.com/find or android.com/lock.

## Summary

| Situation | Can they get in? |
|---|---|
| Locked, 6-digit passcode, up to date | An ordinary thief, no |
| Locked, old phone without updates | More at risk from specialist tools |
| They saw your passcode | Yes, except what Stolen Device Protection or Identity Check protect |
| They took it unlocked | Yes, until it locks |

Next: [Fingerprint and face unlock: how much they protect](biometrics.md).

## Sources

- Apple: [passcode security](https://support.apple.com/guide/security/passcodes-sec20230a10d/web), [Stolen Device Protection](https://support.apple.com/en-us/120340).
- Android: [attempt limits](https://source.android.com/docs/security/features/authentication/rate-limiting), [theft protection](https://support.google.com/android/answer/15146908?hl=en).
- Press: [thefts after watching the passcode (MacRumors' summary of the Wall Street Journal report)](https://www.macrumors.com/2023/02/24/iphone-stolen-passcodes-report/), [iPhone restart after 3 days (TechCrunch)](https://techcrunch.com/2024/11/14/new-apple-security-feature-reboots-iphones-after-3-days-researchers-confirm), [which phones resist (404 Media)](https://www.404media.co/leaked-docs-show-what-phones-cellebrite-can-and-cant-unlock/).
