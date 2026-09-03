<div align="center">

# OneBuild

**Build your Flutter app's Android, iOS, Web, Windows, Linux and macOS
outputs from any computer — no Mac required for iOS.**

[![Go Report](https://img.shields.io/badge/Go-1.21%2B-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License: MIT](https://img.shields.io/github/license/ghaderi0x/onebuild)](LICENSE)
[![Latest release](https://img.shields.io/github/v/release/ghaderi0x/onebuild?include_prereleases)](https://github.com/ghaderi0x/onebuild/releases/latest)
[![Platforms](https://img.shields.io/badge/platform-macOS%20%7C%20Linux%20%7C%20Windows-informational)](#install)
[![Zero dependencies](https://img.shields.io/badge/dependencies-stdlib%20only-brightgreen)](#why)

**[English](README.md) · [فارسی](README.fa.md)**

</div>

---

OneBuild uploads your project to a GitHub repository (yours, on your own
account), generates a GitHub Actions workflow tailored to the platforms you
picked, triggers the build, waits for it, downloads the resulting artifacts,
and keeps a local history of everything you've built. All the actual
compiling happens on GitHub's own runners (including real macOS runners for
iOS) — OneBuild is just the remote control.

Written in Go using only the standard library. The compiled binary is all
you need — no Flutter, no Xcode, and no extra runtime has to be installed
on your machine to run OneBuild itself. (git is used automatically if it's
already on your system, purely as a faster upload path — it's not required.)

Made by **A.M.Ghaderi** · issues & PRs: https://github.com/ghaderi0x/onebuild

---

## Table of contents

- [Why](#why)
- [Requirements](#requirements)
- [Install](#install)
- [Quick start](#quick-start)
- [Step-by-step guide](#step-by-step-guide)
  - [1. Create a GitHub token](#1-create-a-github-token)
  - [2. Run the build wizard](#2-run-the-build-wizard)
  - [3. Pick your project source](#3-pick-your-project-source)
  - [4. Pick your build targets](#4-pick-your-build-targets)
  - [5. Signed iOS builds (optional)](#5-signed-ios-builds-optional)
  - [Getting a certificate without a Mac](#getting-a-certificate-without-a-mac)
  - [6. Watching the build](#6-watching-the-build)
  - [7. Getting your files](#7-getting-your-files)
  - [8. When a build fails](#8-when-a-build-fails)
- [The `history` command](#the-history-command)
- [The `doctor` command](#the-doctor-command)
- [Keeping OneBuild up to date](#keeping-onebuild-up-to-date)
- [All commands](#all-commands)
- [Where things are stored on your machine](#where-things-are-stored-on-your-machine)
- [FAQ / Troubleshooting](#faq--troubleshooting)
- [Extending to other frameworks](#extending-to-other-frameworks)
- [License](#license)

---

## Why

Flutter developers on Windows or Linux can't produce an iOS build locally,
since Xcode only runs on macOS. Buying a Mac just to ship iOS builds is a
real barrier for a lot of solo developers and small teams. OneBuild works
around this by letting GitHub's hosted macOS runners (which Actions gives
you access to for free, within GitHub's usage limits) do the iOS build for
you — along with every other platform Flutter supports, from the same
command.

## Requirements

- A GitHub account (free tier works — Actions minutes may be limited on the
  free plan, see [FAQ](#faq--troubleshooting)).
- Nothing else. OneBuild does **not** need Flutter, Xcode, Android Studio,
  or git installed locally to run — it only needs those to exist on the
  GitHub Actions runner, which GitHub already provides.

## Install

**Option A — download a binary (recommended, no Go needed)**

Grab the file for your platform from the
**[latest release](https://github.com/ghaderi0x/onebuild/releases/latest)**:

| Platform | File |
|---|---|
| macOS (Apple Silicon) | `onebuild-macos-arm64` |
| macOS (Intel) | `onebuild-macos-intel` |
| Linux (x86_64) | `onebuild-linux-amd64` |
| Linux (arm64) | `onebuild-linux-arm64` |
| Windows | `onebuild-windows-amd64.exe` |

macOS/Linux:
```bash
chmod +x onebuild-*
./onebuild-* version
```
On macOS you may need to allow it once under **System Settings → Privacy &
Security → "Allow Anyway"**, since it isn't notarized.

Windows: just run the `.exe` from PowerShell or a terminal.

**Option B — build it yourself** (needs Go 1.21+ installed just for this
one step):

```bash
git clone https://github.com/ghaderi0x/onebuild
cd onebuild
go build -o onebuild .
```

Either way, you get a single self-contained file — move it anywhere you
like, including a folder on your `PATH` so you can run `onebuild` from
anywhere without typing the full path.

## Quick start

```bash
onebuild auth login     # one-time: paste a GitHub token
onebuild build          # answer a few questions, get your builds
onebuild history        # see everything you've built before
```

That's the whole workflow. Everything below explains each step in detail.

---

## Step-by-step guide

### 1. Create a GitHub token

The first time you run `onebuild build` (or if you run `onebuild auth
login` directly), OneBuild will ask for a **GitHub Personal Access
Token**. This is how it creates repositories and starts builds on your
behalf.

1. Go to **https://github.com/settings/tokens/new**
2. Give it any name, e.g. `onebuild`.
3. Set an expiration you're comfortable with (or "No expiration" if you'd
   rather not repeat this step later).
4. Under **scopes**, check:
   - `repo` (full control of private repositories)
   - `workflow` (update GitHub Action workflows)
5. Click **Generate token**, then copy it — GitHub only shows it once.
6. Paste it into OneBuild when asked.

OneBuild encrypts this token and stores it in `~/.onebuild/` on your own
machine. You won't be asked again on future runs. To remove it at any
time, run `onebuild logout`.

> Prefer a fine-grained token instead of a classic one? That works too, as
> long as it has read/write access to Contents, Actions, and Secrets, and
> is allowed to create new repositories (fine-grained tokens need
> "All repositories" access with Administration: write for that last
> part). Classic tokens with `repo` + `workflow` are simpler and are what
> the prompts above assume.

### 2. Run the build wizard

```bash
onebuild build
```

You'll see the OneBuild banner, then a short series of questions. Example
of what the first part looks like:

```
  ? Where is your Flutter project?
      1) A local folder on this computer
      2) An existing GitHub repository (already pushed)
  > Enter number: 1

  ? Path to your Flutter project folder [.]: ~/projects/my_app
  ? App name (used for labels and history) [my_app]: My App
```

### 3. Pick your project source

- **Local folder** — point OneBuild at your Flutter project's root folder
  (the one containing `pubspec.yaml`). OneBuild will:
  - create a new GitHub repository for you (you choose the name and
    whether it's private or public),
  - add a `.github/workflows/onebuild.yml` file to your project,
  - upload everything (skipping `build/`, `.dart_tool/`, `Pods/`,
    `.gradle/`, `node_modules/`, and similar folders that don't belong in
    version control).
- **Existing GitHub repository** — if your project is already pushed to
  GitHub, just give OneBuild the URL (or `owner/repo`). It won't touch
  your files; it only adds/updates the workflow file and triggers a run.

### 4. Pick your build targets

```
  ? Which outputs do you want to build? (comma separated numbers, e.g. 1,3)
      1) Android (.apk)
      2) Android App Bundle (.aab)
      3) iOS - unsigned build (.ipa, needs resigning)
      4) iOS - signed with your certificate (.ipa)
      5) Web
      6) Windows desktop
      7) Linux desktop
      8) macOS desktop
  > Enter numbers: 1,4,5
```

You can pick as many as you want in one run — each becomes its own job in
the generated workflow, and they all build in parallel on GitHub's side.

### 5. Signed iOS builds (optional)

If you selected the **signed iOS** target, OneBuild first asks for your
Apple **Team ID** and the export method (`ad-hoc`, `app-store`,
`development`, or `enterprise`).

Then it checks whether your repository already has four required GitHub
Actions secrets, and if any are missing, it prints exactly what to add and
where:

```
  ⚠ This repository is missing 4 required secret(s) for signed iOS builds:
     - IOS_CERTIFICATE_BASE64
     - IOS_CERTIFICATE_PASSWORD
     - IOS_PROVISIONING_PROFILE_BASE64
     - KEYCHAIN_PASSWORD

  Add them at:
  https://github.com/you/your-repo/settings/secrets/actions/new
```

How to get each value:

| Secret | How to get it |
|---|---|
| `IOS_CERTIFICATE_BASE64` | See [Getting a certificate without a Mac](#getting-a-certificate-without-a-mac) below. |
| `IOS_CERTIFICATE_PASSWORD` | The password you choose while running `onebuild ios-cert package`. |
| `IOS_PROVISIONING_PROFILE_BASE64` | Download the matching `.mobileprovision` file from **https://developer.apple.com/account/resources/profiles/list**, then run `onebuild ios-cert encode path/to/profile.mobileprovision`. |
| `KEYCHAIN_PASSWORD` | Any password you make up — it's only used to protect a temporary keychain created during the CI run, and is never used anywhere else. |

Once the secrets are in place, go back to the OneBuild prompt and choose
**"I've added them, check again."** OneBuild will re-check and continue.
You can also choose to skip the signed iOS target and continue with the
rest of your selected outputs, or cancel entirely.

> The unsigned iOS target doesn't need any of this — it needs no Apple
> account at all, but the resulting `.ipa` **cannot be installed on a
> device as-is**. It needs to be re-signed afterwards with a tool like
> AltStore, Sideloadly, or TrollStore — this is a limitation of unsigned
> iOS builds in general, not something OneBuild can work around.

### Getting a certificate without a Mac

Getting an Apple *distribution certificate* normally means opening
Keychain Access on a Mac to generate a Certificate Signing Request (CSR).
That's not actually an Apple requirement — it's just what Keychain Access
happens to automate. A CSR is a standard, well-defined file format (PKCS#10),
and OneBuild can generate one itself, on any OS:

```bash
onebuild ios-cert csr
```

This asks for your Apple ID email, your name, and a country code, then
creates a private key and a `.certSigningRequest` file locally — nothing
is sent anywhere at this step.

Next:

1. Go to **https://developer.apple.com/account/resources/certificates/add**
2. Choose **Apple Distribution** (or **iOS Distribution**).
3. Upload the `.certSigningRequest` file OneBuild just created.
4. Download the certificate Apple gives you back (a `.cer` file).

Then package it:

```bash
onebuild ios-cert package
```

Give it the path to the downloaded `.cer` file and the private key from
the first step, pick a password, and OneBuild builds the `.p12` file,
base64-encodes it, and saves both — ready to paste as
`IOS_CERTIFICATE_BASE64` and `IOS_CERTIFICATE_PASSWORD`.

This step shells out to `openssl` (it's preinstalled on macOS and Linux;
on Windows it's included with Git for Windows or WSL). If `openssl` isn't
found, OneBuild prints the exact two commands to run yourself instead —
nothing about this requires a Mac.

You'll still need a **provisioning profile** tied to that certificate and
your app's bundle ID — download it from
**https://developer.apple.com/account/resources/profiles/list** (also
possible from a browser on any OS), then:

```bash
onebuild ios-cert encode path/to/profile.mobileprovision
```

to get the base64 value for `IOS_PROVISIONING_PROFILE_BASE64`.

### 6. Watching the build

Once everything is uploaded and the workflow is committed, OneBuild finds
the resulting GitHub Actions run automatically and waits for it, showing a
live status and elapsed time:

```
  ✔ Workflow started: https://github.com/you/your-repo/actions/runs/123456
  ⠙ Building on GitHub Actions... status: in_progress (3m12s elapsed)
```

Multi-platform builds can take anywhere from a few minutes (Android only)
to 20–30 minutes (several platforms including iOS/macOS/Windows/Linux
together, since each hosted runner needs to set up its own toolchain from
scratch — this is normal and expected, not a sign anything is stuck). You
can safely leave the terminal running in the background.

### 7. Getting your files

When the run finishes, OneBuild lists each job with its result:

```
  ✔ Android (.apk)  (https://github.com/you/your-repo/actions/runs/.../job/...)
  ✔ Web  (...)
  ✖ iOS - signed with your certificate (.ipa)  (...)
```

Then it downloads every successful artifact into:

```
~/OneBuild-output/<app-name>-<timestamp>/
```

with one subfolder per artifact (e.g. `app-android-apk/`, `app-web/`),
containing the actual `.apk`, `.aab`, `.ipa`, or platform bundle GitHub
Actions produced.

### 8. When a build fails

For every failed job, OneBuild fetches that job's own log (not the whole
run — so an Android failure doesn't drown out an iOS one) and shows you:

- which job failed, and a direct link to it,
- any structured **GitHub annotations** on that job (for example from
  `flutter analyze` with a problem matcher, or `::error
  file=...,line=...::message` workflow commands) — these come straight
  from GitHub's own Checks API, so they're exact, not guesses,
- the last lines of that job's raw log, right in your terminal.

OneBuild deliberately does **not** try to guess the cause or suggest a
fix — build failures are too varied and context-dependent for a simple
keyword match to get right reliably, and a wrong guess is worse than no
guess. You get the real log; you (or a search engine, or the Flutter/
Gradle/Xcode error message itself) are the best judge of what it means.

If you'd like a copy to keep or share, OneBuild can save a PDF with the
failed jobs, their annotations, and log tails — saved straight to your
**Desktop** for easy access:

```
  ✔ PDF saved to /Users/you/Desktop/onebuild-error-report-20260901-111652-038.pdf
```

---

## The `history` command

```bash
onebuild history
```

Shows every past build: app name, repository link, run link, date, and
where each downloaded artifact ended up locally.

```
  1. [✔] My App
     Repo:   https://github.com/you/my-app
     Run:    https://github.com/you/my-app/actions/runs/123456
     Date:   2026-08-31 10:15
     Artifacts:
       - app-android-apk: /home/you/OneBuild-output/my-app-20260831-101512/app-android-apk
       - app-web: /home/you/OneBuild-output/my-app-20260831-101512/app-web
```

## The `doctor` command

```bash
onebuild doctor
```

A quick environment check — useful before your first run or when
something isn't working:

```
  Info: OS/Arch: darwin/arm64
  ✔ git is installed (will be used for faster uploads)
  ✔ Local config directory is writable (~/.onebuild)
  ✔ api.github.com is reachable
  ✔ A GitHub session is saved
```

## Keeping OneBuild up to date

Every time you run a command like `onebuild build`, OneBuild does a quick
(a few seconds, silently skipped if you're offline) check against this
repository's latest release. If a newer version exists, you'll see:

```
  ⚠ A newer version (v1.1.0) is available. Run 'onebuild update' to update.
```

To update:

```bash
onebuild update
```

This downloads the correct binary for your OS/architecture from the
latest GitHub release and replaces the currently running one in place —
no reinstalling, no re-downloading manually.

## All commands

```
onebuild build            Start the interactive build wizard
onebuild history           Show past builds
onebuild auth login        Save a GitHub token for future runs
onebuild auth logout        Remove the saved GitHub token
onebuild logout               Shortcut for 'onebuild auth logout'
onebuild auth status         Show who is currently logged in
onebuild ios-cert csr        Generate an Apple certificate request (no Mac needed)
onebuild ios-cert package    Package a downloaded certificate into a .p12
onebuild ios-cert encode     Base64-encode a file (e.g. a provisioning profile)
onebuild doctor             Check your local environment
onebuild update              Update OneBuild to the latest version
onebuild version            Print the version number
onebuild help                Show this list
```

## Where things are stored on your machine

| Path | Contents |
|---|---|
| `~/.onebuild/session.json` | Your GitHub login name and encrypted token |
| `~/.onebuild/local.key` | The local encryption key used for the token above |
| `~/.onebuild/history.json` | Your build history |
| `~/OneBuild-output/ios-cert/` | Private key, CSR, certificate and provisioning profile files from `onebuild ios-cert` |
| `~/OneBuild-output/` | Downloaded build artifacts |
| `~/Desktop/` | PDF failure reports, when you ask for one |

Nothing here is ever sent anywhere except direct HTTPS calls to
`api.github.com` (and, only during `onebuild update`, to download a new
binary from this repository's GitHub Releases).

> **A note on the local token storage**: the saved GitHub token is
> encrypted at rest with a locally generated key stored right next to it
> (`~/.onebuild/local.key`). This protects the token from casual
> inspection (e.g. opening the file in a text editor) but isn't a
> substitute for full disk encryption — anyone with the same level of
> access to your user account that could read the encrypted file could
> also read the key next to it. Treat `~/.onebuild/` with the same care
> you'd give any other credential on your machine.

## FAQ / Troubleshooting

**Do I need a paid GitHub plan?**
No. Public repositories get unlimited free Actions minutes; private
repositories on the free plan get a monthly quota (2,000 minutes/month as
of this writing) which is generally plenty for personal projects. Building
many platforms at once, especially macOS/iOS jobs, uses minutes faster
than Android/Web alone.

**Why did my first build take 13 minutes for just an Android APK?**
That's normal, and even a bit on the fast side. GitHub's hosted runners
start from a clean machine every time — installing the Flutter SDK,
downloading the Gradle wrapper, the Android SDK components, and your
project's dependencies all happen from scratch on that first run. Later
builds of the same project are usually a bit faster since pub packages get
cached, though Gradle itself isn't cached between separate runs yet.

**Can I use this for a company/organization repository?**
Yes — when asked for the repository, choose "existing GitHub repository"
and point OneBuild at it, as long as your token has access to it.

**Where does the actual Flutter/Xcode/Gradle version come from?**
From whatever `subosito/flutter-action` installs on the GitHub-hosted
runner at the time of the build (`stable` channel by default) — the exact
same tool most Flutter CI pipelines use.

**Can I edit the generated workflow file afterwards?**
Yes, it's a normal file at `.github/workflows/onebuild.yml` in your repo.
Re-running `onebuild build` against the same project will overwrite it
with a freshly generated version based on your latest answers, so keep
that in mind if you've hand-edited it.

**My push/upload failed with a permissions error.**
Your token has probably expired or doesn't have the right scopes. Run
`onebuild logout` then `onebuild auth login` again with a fresh token that
has `repo` and `workflow` scopes.

## Extending to other frameworks

Version 1.0.0 focuses entirely on Flutter, but the design keeps that
assumption isolated: all of the target definitions and the GitHub Actions
YAML generation live in `internal/workflow/`, separate from the
GitHub/upload/history/UI code. Adding support for another
framework (React Native, plain native Android/iOS, etc.) means adding a
new set of targets there, not touching the rest of the tool.

## License

MIT — see [LICENSE](LICENSE).

Made by **A.M.Ghaderi**.
