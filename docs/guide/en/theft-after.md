# My device was stolen: what to do now

[Español](../es/theft-after.md)

Take a breath. This page is an ordered plan: follow it from top to bottom. You do not need to understand everything, just do the next step.

You will need **another device** (a friend's phone, a computer, a laptop at work) and, ideally, the list of "accounts to secure first" and your IMEI from [Before your device is stolen](theft-before.md). If you never made that list, no problem: the steps below cover the usual accounts.

> **Do not chase the thief.** If the map shows where your device is, give that to the police. Your safety is worth more than any phone.

---

## First 10 minutes: lock it down

### 1. Mark it as lost and lock it

Use the official service for your device from another device or browser:

- **iPhone, iPad, Mac:** icloud.com/find, or the Find My app on another Apple device (a family member's device works through Family Sharing). Choose **Mark As Lost** (iPhone/iPad) or **Lock** (Mac). Add a contact phone number that is **not** the stolen one.
- **Android:** android.com/find (Google Find My Device). Choose **Secure device** to lock it and show a message. If you cannot sign in because your 2FA codes go to the stolen phone, try android.com/lock, which only needs your phone number (if Remote Lock was enabled).
- **Windows laptop:** account.microsoft.com/devices > Find my device > **Lock**.

Marking as lost also suspends Apple Pay and Google Wallet cards on the device.

### 2. Decide: erase now, or wait?

- **Wait (a few hours) if** the device is locked with a passcode, it shows up on the map, and there is a realistic chance of getting it back (you left it in a taxi, a shop, an office). Once you erase it, you usually can no longer see where it is.
- **Erase right away if** it was snatched violently, you think the thief saw your passcode, the device had no passcode, or it holds work or very sensitive data. Erasing is safer than hoping.
- **Do not remove the device from your account** after erasing it. On iPhone and Mac, keeping it linked keeps **Activation Lock** on, so the thief cannot reuse or resell it easily.

### 3. Call your carrier to block the SIM

- Ask them to **suspend the line and block the SIM**, so nobody can receive your SMS codes or make calls.
- Ask for a **replacement SIM or eSIM** in your name. You will need it to receive codes again.
- Ask for an **extra lock against SIM swaps and port-outs** if you do not already have one.

---

## First hour: protect your accounts

Work through these in order. Each one may log the thief out.

### 4. Change your main email password first

Your email can reset almost every other account, so it comes first.

- Change the password from a trusted device.
- Check the recovery email and phone number; remove any you do not recognise.
- Check forwarding rules and filters: attackers sometimes add a rule that silently forwards your mail to them.

### 5. Then Apple, Google and Microsoft accounts

- **Apple Account:** account.apple.com > Sign-In & Security > change the password. Under Devices, check what is still signed in.
- **Google:** myaccount.google.com > Security > change the password, then **Your devices** > sign out of the stolen device.
- **Microsoft:** account.microsoft.com > Security > change the password; under Devices, review and remove access.

### 6. Sign out everywhere and revoke devices

Changing a password does not always log out a session that is already open. Revoke them explicitly:

- [ ] **WhatsApp:** install it on your new SIM to take the account back. Then Settings > Linked devices > log out of anything you do not recognise. If you cannot do it yet, email WhatsApp support to deactivate the account temporarily.
- [ ] **Telegram, Signal:** Settings > Devices / Linked devices > terminate other sessions.
- [ ] **Instagram, Facebook:** Accounts Center > Password and security > Where you're logged in > log out of the stolen device.
- [ ] **X, TikTok, LinkedIn** and others: look for "sessions", "devices" or "where you're logged in" in the security settings.
- [ ] **Password manager:** sign out the stolen device and, if your master password could have been seen, change it.
- [ ] **Work accounts:** tell your IT team now. They can wipe company data and revoke access.

### 7. Call your bank

- Tell them the device was stolen. Ask them to **block the cards** linked to the phone or used in it, and to watch for unusual transfers.
- Remove the stolen device from Apple Pay / Google Wallet if marking as lost did not already do it.
- In the bank app (on a new device), sign out other sessions and review trusted devices.

---

## First day: report and warn

### 8. Report it to the police

- Bring your **IMEI** (both, if it has two SIMs), the **serial number**, the make and model, and the purchase invoice if you have it.
- Ask for a copy of the report. Your carrier and your insurer will need it.
- With the report, ask your carrier to **block the IMEI** so the phone cannot be used on mobile networks.

Where is the IMEI if you did not write it down? On the original box, the purchase invoice, your carrier's bill or customer area, or in your account: Apple (account.apple.com > Devices) and Google (myaccount.google.com > Devices, or google.com/android/find) usually show it.

### 9. Warn your contacts

Thieves and scammers often use a stolen phone or a hijacked WhatsApp to message your contacts: "I changed my number", "can you send me money", "send me the code I just texted you".

- Post a short message from another account or ask a friend to spread it: "My phone was stolen. If you get any strange message from me asking for money or codes, it's not me."

### 10. Beware of fake "we found your iPhone" messages

A few days after the theft you may get an SMS, email or WhatsApp saying your device was found, with a link to "see its location". **It is a trap** to steal your Apple Account or Google password so they can unlock the device.

- Apple and Google will **never** send you a link asking for your password to show a found device.
- Do not click. Open icloud.com/find or android.com/find by typing the address yourself.

---

## Following weeks: watch and close

### 11. Watch for suspicious activity

- [ ] **Sign-in alerts** from your email, Apple, Google, Microsoft and social networks. If something looks wrong, change the password again and review sessions.
- [ ] **Bank movements**: check them daily for a couple of weeks.
- [ ] **SIM swap signs**: if your phone suddenly loses signal for no reason, call your carrier right away. Someone may be trying to get a duplicate of your SIM.
- [ ] **Password reset emails** you did not ask for.

### 12. Move 2FA off SMS

If your accounts used SMS codes, now is the moment to switch to passkeys or an authenticator app, and to save new recovery codes offline. See [Before your device is stolen](theft-before.md), section 8.

### 13. Insurance claim

- If the device was insured (home insurance, card insurance, the carrier's or the manufacturer's plan), file the claim with the police report, the IMEI or serial number, and the invoice.
- Check the deadline: many insurers give only a few days.

### 14. Set up the replacement properly

When you get a new device, restore from your backup and go through [Before your device is stolen](theft-before.md) again. If you have a Mac, run `hardenme scan` on it.

---

## Quick checklist

- [ ] Marked as lost / locked
- [ ] Decided whether to erase
- [ ] SIM blocked, replacement requested
- [ ] Main email password changed
- [ ] Apple / Google / Microsoft password changed
- [ ] Sessions and devices revoked (WhatsApp, social networks, password manager, work)
- [ ] Bank told, cards blocked
- [ ] Police report with IMEI, IMEI blocked by carrier
- [ ] Contacts warned
- [ ] Watching for phishing, SIM swaps and strange logins
- [ ] Insurance claim filed
