# Fingerprint and face unlock: how much they protect

[Español](../es/biometrics.md)

Short answer: **unlocking with your fingerprint or face is safer than typing your passcode in public**, because nobody can see it over your shoulder. Its weak point is the passcode behind it.

## What is worth knowing

### Your fingerprint and face do not leave the phone

- **iPhone:** Face ID and Touch ID data is kept in a separate chip in the phone itself. It is not backed up to iCloud and does not reach Apple's servers.
- **Android (Pixel):** your face model and fingerprints do not leave the phone either, and are not shared with Google or with apps.

### Someone else's face or finger is very unlikely to work

According to Apple, the chance that a random person unlocks your iPhone is:

- less than 1 in 1,000,000 with Face ID (higher for twins, siblings who look alike and children under 13);
- 1 in 50,000 with Touch ID.

### Careful with face unlock on some Android phones

Not all face unlock is equal. Some phones only use the normal camera, with no depth sensors, and **can be fooled by a photo**.

- In tests by the UK consumer organisation Which? (2026), a printed photo unlocked 64% of the phones tested since 2022. iPhones, the Pixel 8 and later and the Galaxy S26 passed.
- Android sorts biometrics into three levels. Only the highest ("strong") can be used to pay with Google Wallet. Most banking apps do not accept weak face unlock either.
- **If your Android phone can be fooled by a photo, unlock with your fingerprint**, not your face.

### The passcode is always behind it

Biometrics are a shortcut. If they fail, the phone asks for the passcode, and whoever knows the passcode can add their own face or fingerprint. So:

- your fingerprint or face is only as strong as your passcode: use 6 digits or more;
- on iPhone, [Stolen Device Protection](phone-unlock.md) means the passcode is not enough for the most sensitive actions.

The phone asks for the passcode, not biometrics, in these cases:

- **iPhone:** when it is switched on or restarted, if it has not been unlocked for more than 48 hours, after five failed face or fingerprint tries, and after opening the power-off or emergency screen.
- **Android (Pixel):** when it restarts, after several failed tries, and when more than 72 hours have passed since you last used the PIN, pattern or password.

### If someone forces you

The real risk with biometrics is not that they are faked, it is that someone forces you to look at the phone or put your finger on it. You can turn them off in a second:

- **iPhone:** hold the side button and a volume button for 2 seconds, until the power-off screen appears. From then on only the passcode works.
- **Android:** use "Lockdown" in the power menu, if your phone shows it (the name changes with the brand). It turns off fingerprint and face unlock and hides notifications until you enter the PIN. On Samsung, hold Volume down and Power (on older versions you must enable it in settings first).

Face ID also checks that your eyes are open and looking at the iPhone, so it does not unlock while you sleep. Do not turn off "Require Attention for Face ID".

## What to do

1. Use your fingerprint or face so you do not type the passcode in public.
2. Set a passcode of 6 digits or more.
3. On Android, check whether your face unlock is the secure kind. If you do not know, use your fingerprint.
4. Learn the gesture that turns biometrics off in an emergency.

Next: [Is it safe to have your bank's app on your phone?](bank-app.md).

## Sources

- Apple: [Face ID and privacy](https://support.apple.com/en-us/102381), [Touch ID](https://support.apple.com/en-us/105095), [Face ID and Touch ID security](https://support.apple.com/guide/security/face-id-and-touch-id-security-sec067eb0c9e/web).
- Google: [face unlock on Pixel](https://support.google.com/pixelphone/answer/9517039?hl=en), [fingerprint on Pixel](https://support.google.com/pixelphone/answer/6285273?hl=en), [Android biometric levels](https://source.android.com/docs/security/features/biometric/measure), [verification in Google Wallet](https://support.google.com/wallet/answer/12059519?hl=en).
- Which?: [face unlock fooled by a photo](https://www.which.co.uk/news/article/why-using-face-unlock-on-a-phone-could-be-risking-your-data-az8fe2G7uX8n).
- Samsung: [Lockdown mode](https://www.samsung.com/us/support/answer/ANS10013057/).
