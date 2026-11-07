# Before your device is stolen

[Español](../es/theft-before.md)

Most of what decides how bad a theft will be happens **before** it. A phone or laptop that is locked, encrypted, findable and backed up is just an object you lost. One without those things is a door into your email, your bank and your contacts.

This guide is a checklist. Do it once, calmly, at home. It takes about an hour, and you can do it in pieces.

**Who needs each step:**

- **Everyone**: do it, whatever your situation.
- **If you carry it outside**: the device leaves home (commute, travel, cafés, school).
- **If it holds work or banking data**: work email, company files, bank or payment apps.
- **Elevated risk**: you are a journalist, activist, public figure, or someone is targeting you personally.

> Tip: run `hardenme scan` on your Mac to see which of these are already done. On phones, use this page as a manual checklist for now.

---

## 1. Lock the screen (everyone)

A thief who can open your device in the first minutes can do almost anything.

- [ ] **Use a real screen lock.** On phones, a 6-digit code at least; an alphanumeric password is better if it holds work or banking data. Avoid patterns, birthdays and `123456`.
- [ ] **Short auto-lock.** Phones: 30 seconds to 1 minute. Laptops: 5 minutes or less, and a password required immediately after sleep or screen saver.
- [ ] **Use Face ID, Touch ID or a fingerprint.** It makes a long passcode bearable, so you are not tempted to choose a weak one.
- [ ] **Watch who sees you type it.** Many phone thefts start with someone watching you enter your code in a bar or on the bus. In public, prefer Face ID or a fingerprint, and cover the screen when you type the code.

Where to find it:

- iPhone: Settings > Face ID & Passcode, and Settings > Display & Brightness > Auto-Lock.
- Android: Settings > Security (or Security & privacy) > Screen lock, and Settings > Display > Screen timeout.
- Mac: System Settings > Lock Screen.
- Windows: Settings > Accounts > Sign-in options, and Settings > System > Power & battery > Screen and sleep.

## 2. Turn on theft protection on your phone (if you carry it outside)

Modern phones have features built for exactly this situation. They are often off by default.

**iPhone (iOS 17.3 or later)**

- [ ] **Stolen Device Protection**: Settings > Face ID & Passcode > Stolen Device Protection. With it on, changing your Apple Account password or turning off Find My needs Face ID or Touch ID, and some changes add a one-hour delay when you are away from familiar places. A thief who saw your code cannot use it to take over your account.
- [ ] **Turn off Control Center on the lock screen**: Settings > Face ID & Passcode > Allow Access When Locked > Control Center off. Otherwise a thief can switch on Airplane Mode without unlocking and your phone disappears from Find My.

**Android (Android 10 or later with Google Play services; names vary by brand)**

- [ ] **Theft Detection Lock**: locks the screen if someone snatches the phone and runs, rides or drives away.
- [ ] **Offline Device Lock**: locks the screen if the phone goes offline for a while after being stolen.
- [ ] **Identity Check** (Android 15 or later, some devices): asks for your fingerprint or face to change sensitive settings when you are away from trusted places.
- [ ] **Remote Lock**: lets you lock the phone from android.com/lock using only your phone number.

Look for them in Settings > Google > All services > Theft protection, or search "theft" in Settings.

## 3. Make sure you can find, lock and erase it (everyone)

- [ ] **iPhone, iPad and Mac:** turn on **Find My** and **Send Last Location** (Settings > [your name] > Find My). On a Mac: System Settings > [your name] > iCloud > Find My Mac.
- [ ] **Android:** turn on **Find My Device** (Settings > Google > Find My Device, or Settings > Security).
- [ ] **Windows:** turn on **Find my device** (Settings > Privacy & security > Find my device). It needs a Microsoft account.
- [ ] **Test it now.** Open icloud.com/find, android.com/find or account.microsoft.com/devices from another device and check that yours shows up on the map. A feature you never tested is a feature you cannot trust.

## 4. Encrypt the disk (everyone with a laptop; optional on a desktop that never leaves home)

Encryption means a thief who removes the disk sees only noise.

- [ ] **Mac:** turn on **FileVault** (System Settings > Privacy & Security > FileVault).
- [ ] **Windows:** turn on **BitLocker** or **Device encryption** (Settings > Privacy & security > Device encryption). Save the recovery key somewhere other than the laptop itself.
- [ ] **Phones:** iPhones and modern Android phones are encrypted automatically **once you set a screen lock**. Step 1 covers this.

## 5. Keep the lock screen quiet (everyone)

Your lock screen is visible to whoever holds the phone. Security codes sent by SMS or email often appear there.

- [ ] **iPhone:** Settings > Notifications > Show Previews > When Unlocked.
- [ ] **Android:** Settings > Notifications > Notifications on lock screen > Hide sensitive content (or "Don't show notifications").
- [ ] **Mac and Windows:** hide notification previews on the lock screen too.

## 6. Protect your SIM (if you carry it outside)

With your SIM in another phone, a thief can receive your SMS codes and reset your accounts.

- [ ] **Set a SIM PIN.** iPhone: Settings > Mobile Data (Cellular) > SIM PIN. Android: Settings > Security > SIM card lock. Your carrier's default PIN is usually printed on the card holder; change it.
- [ ] **Ask your carrier for a port-out or SIM-swap lock**, or at least an extra password for any change to your line. Ask them what they offer: names differ.
- [ ] **Have your carrier's support number** written down somewhere other than your phone.

## 7. Write down your IMEI and serial numbers NOW (everyone)

When a phone is stolen, the police and the carrier will ask for the IMEI. You will not be able to look it up on a phone you no longer have.

- [ ] **Phone IMEI:** dial `*#06#`, or look in Settings > General > About (iPhone) or Settings > About phone (Android). Phones with two SIMs have two IMEIs: note both.
- [ ] **Laptop serial number:** Mac: Apple menu > About This Mac. Windows: Settings > System > About, or the sticker under the laptop.
- [ ] **Keep the purchase invoice** (a photo or PDF is fine). Insurers and the police often ask for it.
- [ ] **Store all this off the device**: in your password manager, in an email to yourself, or on paper at home.

## 8. Secure your accounts (everyone)

The device is replaceable. Your accounts are what a thief really wants.

- [ ] **Use a password manager.** One strong, unique password per site. Make sure you can open it from another device.
- [ ] **Turn on two-factor authentication (2FA)** on email, Apple Account, Google, Microsoft, bank and social networks. Prefer **passkeys** or an **authenticator app** over SMS: SMS codes follow your SIM.
- [ ] **Save recovery codes offline**: printed or written down, kept at home. Not only on the phone that might be stolen.
- [ ] **Set up a recovery contact**: Apple Account (Settings > [your name] > Sign-In & Security > Account Recovery) and a recovery email or phone on Google and Microsoft that is not the stolen phone itself.
- [ ] **If it holds work or banking data:** check that your bank app has its own PIN or biometric lock, and tell your IT team which device you use for work.

## 9. Back up (everyone)

A backup turns "I lost everything" into "I lost a device".

- [ ] **iPhone:** iCloud Backup or an encrypted backup on your computer.
- [ ] **Android:** Settings > Google > Backup.
- [ ] **Mac:** Time Machine to an external disk, plus iCloud Drive for documents.
- [ ] **Windows:** OneDrive folder backup or File History.
- [ ] **WhatsApp and other messaging apps:** turn on their chat backup; it is separate from the phone backup.
- [ ] **Check that the last backup is recent.**

## 10. Make your "accounts to secure first" list (everyone)

If your phone is stolen, you will be stressed and in a hurry. Write this list now, keep it off the phone, and you will thank yourself later. Put them in order:

1. **Main email** (it can reset everything else).
2. **Apple Account / Google account / Microsoft account.**
3. **Bank and payment apps** (with the bank's phone number to block cards).
4. **WhatsApp, Telegram, Signal** and other messaging apps.
5. **Social networks** (Instagram, Facebook, X, TikTok, LinkedIn).
6. **Work accounts** and your IT team's contact, if any.
7. **Carrier** customer service number.

Then read [My device was stolen: what to do now](theft-after.md), so you already know the plan.

## Extra steps for elevated risk

- [ ] **iPhone:** Settings > Face ID & Passcode > Erase Data after 10 failed attempts (not on devices used by kids). Turn on **Advanced Data Protection** for iCloud and consider **Lockdown Mode**.
- [ ] **Turn off USB accessories when locked** (iPhone: Allow Access When Locked > USB Accessories off).
- [ ] **Use an alphanumeric passcode** on every device.
- [ ] **Mac:** consider Lockdown Mode (System Settings > Privacy & Security).
- [ ] **Run Safety Check** on iPhone if you share or shared accounts with someone you no longer trust.
- [ ] If you suspect targeted spyware, look at the Mobile Verification Toolkit (MVT) by Amnesty International, or ask a digital-security helpline for help.
