It's a beta. It has worked well on several of my own playthroughs, but you will find things that didn't. Please [open an issue](https://github.com/dstrus/hermaeus/issues) if and when something looks wrong.

## Downloads

- **Windows:** `Hermaeus-windows-x64.zip`. Put `Hermaeus.exe` in your MO2 folder (beside `ModOrganizer.exe`), or in your Skyrim folder (beside `Skyrim.exe` or `SkyrimVR.exe`) if you use Vortex or nothing, and run it.
- **macOS:** `Hermaeus-macos.zip`, for Apple silicon and Intel.
- **Linux:** `Hermaeus-linux-x64.tar.gz`.

`SHA256SUMS` lists each download's checksum.

## Before you start

- Hermaeus won't save anything while Skyrim is running, because SkyrimNet keeps its database open. You can still browse.
- Before the first change to each database in a session, it copies that database to `hermaeus-backups` in the SkyrimNet folder. Every save is one transaction, so a failed save changes nothing.
- Keyboard shortcuts are shown in a shortcut bar in the lower-right. Those are also clickable buttons.
- The builds aren't code-signed yet, so your system will warn you the first time:
  - **Windows:** if SmartScreen says "Windows protected your PC," choose **More info**, then **Run anyway**.
  - **macOS:** if it says Apple couldn't verify Hermaeus, open **System Settings → Privacy & Security** and choose **Open Anyway**.
