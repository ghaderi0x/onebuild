<div align="center">

# OneBuild

**خروجی Android، iOS، Web، Windows، Linux و macOS اپلیکیشن فلاترت رو از روی هر سیستمی بگیر —
بدون نیاز به مک برای iOS.**

[![Go Report](https://img.shields.io/badge/Go-1.21%2B-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License: MIT](https://img.shields.io/github/license/ghaderi0x/onebuild)](LICENSE)
[![Latest release](https://img.shields.io/github/v/release/ghaderi0x/onebuild?include_prereleases)](https://github.com/ghaderi0x/onebuild/releases/latest)
[![Platforms](https://img.shields.io/badge/platform-macOS%20%7C%20Linux%20%7C%20Windows-informational)](#نصب)
[![Zero dependencies](https://img.shields.io/badge/dependencies-stdlib%20only-brightgreen)](#چرا)

**[English](README.md) · [فارسی](README.fa.md)**

</div>

---

OneBuild پروژه‌ت رو توی یه ریپازیتوری گیت‌هاب (روی اکانت خودت) آپلود می‌کنه، یه
GitHub Actions workflow متناسب با پلتفرم‌هایی که انتخاب کردی می‌سازه، بیلد رو
اجرا می‌کنه، منتظرش می‌مونه، فایل‌های خروجی رو دانلود می‌کنه، و یه تاریخچه‌ی
محلی از همه‌ی build هایی که گرفتی نگه می‌داره. کل کامپایل واقعی روی سرورهای
خود گیت‌هاب انجام میشه (شامل runner واقعی macOS برای iOS) — OneBuild فقط
ریموت‌کنترلشه.

با Go و فقط با کتابخونه‌ی استاندارد نوشته شده. فایل باینری کامپایل‌شده تنها
چیزیه که لازم داری — نه Flutter، نه Xcode، و هیچ runtime اضافه‌ای لازم نیست
روی سیستم خودت نصب کنی. (اگه `git` از قبل روی سیستمت باشه، خودکار ازش برای
آپلود سریع‌تر استفاده میشه — ولی اجباری نیست.)

ساخته‌ی **A.M.Ghaderi** · گزارش باگ و مشارکت: https://github.com/ghaderi0x/onebuild

---

## فهرست مطالب

- [چرا](#چرا)
- [پیش‌نیازها](#پیش‌نیازها)
- [نصب](#نصب)
- [شروع سریع](#شروع-سریع)
- [راهنمای قدم‌به‌قدم](#راهنمای-قدم‌به‌قدم)
  - [۱. ساخت توکن گیت‌هاب](#۱-ساخت-توکن-گیت‌هاب)
  - [۲. اجرای wizard ساخت](#۲-اجرای-wizard-ساخت)
  - [۳. انتخاب منبع پروژه](#۳-انتخاب-منبع-پروژه)
  - [۴. انتخاب خروجی‌ها](#۴-انتخاب-خروجی‌ها)
  - [۵. خروجی iOS امضاشده (اختیاری)](#۵-خروجی-ios-امضاشده-اختیاری)
  - [گرفتن گواهی iOS بدون مک](#گرفتن-گواهی-ios-بدون-مک)
  - [۶. رصد کردن build](#۶-رصد-کردن-build)
  - [۷. دریافت فایل‌ها](#۷-دریافت-فایل‌ها)
  - [۸. وقتی build فیل میشه](#۸-وقتی-build-فیل-میشه)
- [دستور `history`](#دستور-history)
- [دستور `doctor`](#دستور-doctor)
- [به‌روز نگه‌داشتن OneBuild](#به‌روز-نگه‌داشتن-onebuild)
- [همه‌ی دستورات](#همه‌ی-دستورات)
- [این ابزار چی و کجا روی سیستمت ذخیره می‌کنه](#این-ابزار-چی-و-کجا-روی-سیستمت-ذخیره-می‌کنه)
- [سوالات متداول / رفع اشکال](#سوالات-متداول--رفع-اشکال)
- [توسعه برای زبان‌های دیگه](#توسعه-برای-زبان‌های-دیگه)
- [لایسنس](#لایسنس)

---

## چرا

توسعه‌دهنده‌های فلاتر روی ویندوز یا لینوکس نمی‌تونن خروجی iOS رو محلی بگیرن،
چون Xcode فقط روی macOS نصب میشه. خرید یه مک فقط برای همین کار، برای خیلی از
توسعه‌دهنده‌های مستقل و تیم‌های کوچیک یه مانع واقعیه. OneBuild این مشکل رو با
استفاده از runner های macOS خود گیت‌هاب (که Actions به‌صورت رایگان و در حد
سقف مصرف در اختیارت می‌ذاره) حل می‌کنه — همراه با تمام پلتفرم‌های دیگه‌ای که
فلاتر پشتیبانی می‌کنه، همه از یه دستور.

## پیش‌نیازها

- یه اکانت گیت‌هاب (پلن رایگان کافیه — ممکنه دقیقه‌های Actions محدود باشه،
  به بخش [سوالات متداول](#سوالات-متداول--رفع-اشکال) نگاه کن).
- همین و بس. OneBuild نیازی به نصب Flutter، Xcode، Android Studio یا حتی
  git روی سیستم خودت **نداره** — این‌ها فقط باید روی runner گیت‌هاب باشن، که
  از قبل توسط خودشون فراهم شده.

## نصب

**روش الف — دانلود باینری (پیشنهادی، بدون نیاز به Go)**

فایل مخصوص سیستم خودت رو از **[آخرین release](https://github.com/ghaderi0x/onebuild/releases/latest)**
دانلود کن:

| پلتفرم | فایل |
|---|---|
| مک (Apple Silicon) | `onebuild-macos-arm64` |
| مک (Intel) | `onebuild-macos-intel` |
| لینوکس (x86_64) | `onebuild-linux-amd64` |
| لینوکس (arm64) | `onebuild-linux-arm64` |
| ویندوز | `onebuild-windows-amd64.exe` |

مک/لینوکس:
```bash
chmod +x onebuild-*
./onebuild-* version
```
روی مک ممکنه یه‌بار لازم باشه از مسیر **System Settings → Privacy & Security
→ "Allow Anyway"** اجازه‌ی اجرا بدی، چون فایل notarize نشده.

ویندوز: کافیه فایل `.exe` رو از PowerShell یا ترمینال اجرا کنی.

**روش ب — خودت build کن** (فقط برای همین یه قدم به Go 1.21+ نیاز داری):

```bash
git clone https://github.com/ghaderi0x/onebuild
cd onebuild
go build -o onebuild .
```

توی هر دو روش، یه فایل مستقل و تک‌تکه می‌گیری — هرجا خواستی جابه‌جاش کن،
حتی توی یه پوشه‌ای که توی `PATH` سیستمته تا بتونی از هرجا فقط با نوشتن
`onebuild` اجراش کنی.

## شروع سریع

```bash
onebuild auth login     # یه‌بار: توکن گیت‌هابت رو بده
onebuild build          # به چندتا سوال جواب بده، خروجی بگیر
onebuild history        # همه‌ی build های قبلیت رو ببین
```

همه‌ی جریان کار همینه. بقیه‌ی این راهنما هر قدم رو کامل توضیح می‌ده.

---

## راهنمای قدم‌به‌قدم

### ۱. ساخت توکن گیت‌هاب

اولین باری که `onebuild build` رو می‌زنی (یا مستقیم `onebuild auth login`)،
OneBuild یه **GitHub Personal Access Token** ازت می‌خواد. با همین توکنه که
می‌تونه از طرف تو ریپو بسازه و build رو اجرا کنه.

1. برو **https://github.com/settings/tokens/new**
2. یه اسم بهش بده، مثلاً `onebuild`
3. یه expiration انتخاب کن (یا "No expiration" اگه نمی‌خوای دوباره این کار رو تکرار کنی)
4. زیر بخش **scopes**، این دوتا رو تیک بزن:
   - `repo` (دسترسی کامل به ریپوهای private)
   - `workflow` (آپدیت فایل‌های GitHub Action)
5. روی **Generate token** بزن، بعد کپیش کن — گیت‌هاب فقط یه‌بار نشونش می‌ده.
6. توی OneBuild پیستش کن.

OneBuild این توکن رو رمزنگاری می‌کنه و توی `~/.onebuild/` روی سیستم خودت
ذخیره می‌کنه. دیگه دفعه‌های بعد ازت نمی‌پرسه. برای حذفش هر وقت خواستی:
`onebuild logout`.

> ترجیح می‌دی از fine-grained token به‌جای classic استفاده کنی؟ اونم جواب
> می‌ده، فقط باید دسترسی خواندن/نوشتن به Contents، Actions و Secrets داشته
> باشه و اجازه‌ی ساخت ریپوی جدید بهش داده شده باشه (fine-grained token ها
> برای این کار به دسترسی "All repositories" با "Administration: write"
> نیاز دارن). ولی classic token با scope های `repo` + `workflow` ساده‌تره
> و همون چیزیه که این راهنما فرض کرده.

### ۲. اجرای wizard ساخت

```bash
onebuild build
```

بنر OneBuild رو می‌بینی، بعد چندتا سوال کوتاه. مثلاً بخش اولش این‌شکلیه:

```
  ? Where is your Flutter project?
      1) A local folder on this computer
      2) An existing GitHub repository (already pushed)
  > Enter number: 1

  ? Path to your Flutter project folder [.]: ~/projects/my_app
  ? App name (used for labels and history) [my_app]: My App
```

### ۳. انتخاب منبع پروژه

- **پوشه‌ی محلی** — آدرس ریشه‌ی پروژه‌ی فلاترت (همون‌جایی که `pubspec.yaml`
  توشه) رو بده. OneBuild:
  - یه ریپوی گیت‌هاب جدید برات می‌سازه (اسم و private/public بودنش دست خودته)،
  - یه فایل `.github/workflows/onebuild.yml` به پروژه‌ت اضافه می‌کنه،
  - همه‌چیز رو آپلود می‌کنه (به‌جز `build/`, `.dart_tool/`, `Pods/`,
    `.gradle/`, `node_modules/` و پوشه‌های مشابه که اصلاً نباید توی
    version control باشن).
- **ریپوی موجود روی گیت‌هاب** — اگه پروژه‌ت از قبل روی گیت‌هاب پوش شده، فقط
  URL (یا `owner/repo`) رو بده. فایل‌هاتو دست نمی‌زنه، فقط فایل workflow رو
  اضافه/آپدیت می‌کنه و اجراش می‌کنه.

### ۴. انتخاب خروجی‌ها

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

هر تعداد که بخوای می‌تونی همزمان انتخاب کنی — هرکدوم یه job جدا توی workflow
میشه و همه‌شون موازی روی گیت‌هاب اجرا میشن.

### ۵. خروجی iOS امضاشده (اختیاری)

اگه target «iOS signed» رو انتخاب کنی، اول ازت **Team ID** اپل و روش export
(`ad-hoc`, `app-store`, `development`, یا `enterprise`) رو می‌پرسه.

بعد چک می‌کنه که ریپوت از قبل ۴ تا secret لازم رو داره یا نه، و اگه چیزی کم
بود دقیقاً می‌گه چی باید اضافه بشه و کجا:

```
  ⚠ This repository is missing 4 required secret(s) for signed iOS builds:
     - IOS_CERTIFICATE_BASE64
     - IOS_CERTIFICATE_PASSWORD
     - IOS_PROVISIONING_PROFILE_BASE64
     - KEYCHAIN_PASSWORD

  Add them at:
  https://github.com/you/your-repo/settings/secrets/actions/new
```

چطور هرکدوم رو بسازی:

| Secret | چطور بسازیش |
|---|---|
| `IOS_CERTIFICATE_BASE64` | بخش [گرفتن گواهی iOS بدون مک](#گرفتن-گواهی-ios-بدون-مک) رو ببین. |
| `IOS_CERTIFICATE_PASSWORD` | همون پسوردی که موقع اجرای `onebuild ios-cert package` انتخاب می‌کنی. |
| `IOS_PROVISIONING_PROFILE_BASE64` | فایل `.mobileprovision` رو از **https://developer.apple.com/account/resources/profiles/list** دانلود کن، بعد `onebuild ios-cert encode path/to/profile.mobileprovision` رو بزن. |
| `KEYCHAIN_PASSWORD` | هر پسوردی که خودت بسازی — فقط برای محافظت از یه keychain موقتی توی CI استفاده میشه، جای دیگه‌ای کاربرد نداره. |

بعد از اضافه کردن secret ها، برگرد به OneBuild و گزینه‌ی **«اضافه‌شون کردم،
دوباره چک کن»** رو بزن. می‌تونی همچنین target iOS signed رو رد کنی و با
بقیه‌ی خروجی‌ها ادامه بدی، یا کل build رو کنسل کنی.

> target iOS unsigned به هیچ‌کدوم از این‌ها نیاز نداره، اصلاً اکانت اپل هم
> لازم نیست — ولی `.ipa` نهایی **مستقیم روی گوشی نصب نمیشه**. باید بعداً با
> ابزاری مثل AltStore، Sideloadly یا TrollStore دوباره امضاش کنی — این
> محدودیت ذاتی خروجی‌های iOS بدون امضاست، نه چیزی که OneBuild بتونه دورش
> بزنه.

### گرفتن گواهی iOS بدون مک

معمولاً برای گرفتن *distribution certificate* اپل باید Keychain Access رو
روی مک باز کنی تا یه CSR (Certificate Signing Request) بسازه. ولی این واقعاً
یه الزام اپل نیست — فقط کاریه که Keychain Access خودکارش می‌کنه. CSR یه
فرمت استاندارد (PKCS#10) هست، و OneBuild خودش می‌تونه روی هر سیستم‌عاملی
بسازدش:

```bash
onebuild ios-cert csr
```

ازت ایمیل اپل آیدی، اسم، و کد کشور می‌پرسه، بعد یه کلید خصوصی و یه فایل
`.certSigningRequest` محلی می‌سازه — توی این مرحله چیزی جایی فرستاده نمیشه.

بعدش:

1. برو **https://developer.apple.com/account/resources/certificates/add**
2. گزینه‌ی **Apple Distribution** (یا **iOS Distribution**) رو انتخاب کن
3. فایل `.certSigningRequest` که OneBuild ساخته رو آپلود کن
4. گواهی‌ای که اپل بهت می‌ده (یه فایل `.cer`) رو دانلود کن

بعد پکیجش کن:

```bash
onebuild ios-cert package
```

آدرس فایل `.cer` دانلودشده و کلید خصوصی مرحله‌ی اول رو بده، یه پسورد انتخاب
کن، و OneBuild فایل `.p12` رو می‌سازه، base64 هم می‌کنه، و هردو رو ذخیره
می‌کنه — آماده برای پیست کردن به‌عنوان `IOS_CERTIFICATE_BASE64` و
`IOS_CERTIFICATE_PASSWORD`.

این قدم به `openssl` نیاز داره (روی مک و لینوکس از قبل نصبه؛ روی ویندوز با
Git for Windows یا WSL میاد). اگه `openssl` پیدا نشه، OneBuild دقیقاً همون
دو دستوری که باید خودت بزنی رو نشونت می‌ده — هیچ‌کدوم از این مراحل به مک
نیاز نداره.

هنوز به یه **provisioning profile** متناظر با همون گواهی و bundle ID اپت
نیاز داری — از **https://developer.apple.com/account/resources/profiles/list**
دانلودش کن (این‌هم از هر مرورگری روی هر سیستمی قابل انجامه)، بعد:

```bash
onebuild ios-cert encode path/to/profile.mobileprovision
```

تا مقدار base64 برای `IOS_PROVISIONING_PROFILE_BASE64` رو بگیری.

### ۶. رصد کردن build

بعد از آپلود و commit شدن workflow، OneBuild خودش run مربوطه رو روی GitHub
Actions پیدا می‌کنه و منتظرش می‌مونه، با نمایش زنده‌ی وضعیت و زمان سپری‌شده:

```
  ✔ Workflow started: https://github.com/you/your-repo/actions/runs/123456
  ⠙ Building on GitHub Actions... status: in_progress (3m12s elapsed)
```

build های چندپلتفرمی می‌تونن از چند دقیقه (فقط Android) تا ۲۰-۳۰ دقیقه طول
بکشن (وقتی چندتا پلتفرم مثل iOS/macOS/Windows/Linux با هم انتخاب شده، چون
هر runner باید کل toolchain رو از صفر نصب کنه — این کاملاً عادیه، نشونه‌ی
گیر کردن چیزی نیست). می‌تونی ترمینال رو باز بذاری و بری سراغ کار دیگه.

### ۷. دریافت فایل‌ها

وقتی run تموم شد، OneBuild نتیجه‌ی هر job رو نشون می‌ده:

```
  ✔ Android (.apk)  (https://github.com/you/your-repo/actions/runs/.../job/...)
  ✔ Web  (...)
  ✖ iOS - signed with your certificate (.ipa)  (...)
```

بعد هر artifact موفق رو توی این مسیر دانلود می‌کنه:

```
~/OneBuild-output/<app-name>-<timestamp>/
```

با یه زیرپوشه برای هر artifact (مثلاً `app-android-apk/`, `app-web/`) که
فایل واقعی `.apk`, `.aab`, `.ipa` یا bundle پلتفرم موردنظر توشه.

### ۸. وقتی build فیل میشه

برای هر job که فیل بشه، OneBuild لاگ همون job رو می‌گیره (نه کل run — تا
خطای Android لاگ iOS رو گم نکنه) و نشونت می‌ده:

- کدوم job فیل شده، با لینک مستقیمش،
- هر annotation ساختاریافته‌ای که خود گیت‌هاب روی اون job ثبت کرده (مثلاً
  از `flutter analyze` با problem matcher، یا دستورات
  `::error file=...,line=...::message`) — این‌ها مستقیم از GitHub Checks
  API میان، یعنی دقیق‌ان، نه حدس،
- خط‌های آخر لاگ خام همون job، مستقیم توی ترمینالت.

OneBuild **عمداً** سعی نمی‌کنه دلیل خطا رو حدس بزنه یا راه‌حل پیشنهاد بده —
خطاهای build خیلی متنوع و وابسته به context هستن که یه تطبیق ساده‌ی کلمه‌ای
بتونه قابل‌اعتماد تشخیصشون بده، و یه حدس غلط بدتر از هیچ‌حدسیه. لاگ واقعی رو
بهت می‌ده؛ خودت (یا یه جستجوی ساده، یا پیام خطای خود Flutter/Gradle/Xcode)
بهترین کسیه که می‌تونه معنیش رو بفهمه.

اگه بخوای یه نسخه برای نگه‌داری یا اشتراک‌گذاری داشته باشی، OneBuild می‌تونه
یه PDF از job های فیل‌شده، annotation ها و آخر لاگ‌شون بسازه — مستقیم روی
**Desktop** ذخیره میشه تا راحت پیداش کنی:

```
  ✔ PDF saved to /Users/you/Desktop/onebuild-error-report-20260901-111652-038.pdf
```

---

## دستور `history`

```bash
onebuild history
```

همه‌ی build های قبلیت رو نشون می‌ده: اسم اپ، لینک ریپو، لینک run، تاریخ، و
اینکه هر artifact دانلودشده کجای سیستمت هست.

```
  1. [✔] My App
     Repo:   https://github.com/you/my-app
     Run:    https://github.com/you/my-app/actions/runs/123456
     Date:   2026-08-31 10:15
     Artifacts:
       - app-android-apk: /home/you/OneBuild-output/my-app-20260831-101512/app-android-apk
       - app-web: /home/you/OneBuild-output/my-app-20260831-101512/app-web
```

## دستور `doctor`

```bash
onebuild doctor
```

یه چک سریع محیط سیستمت — قبل از اولین اجرا یا هروقت چیزی درست کار نکرد
مفیده:

```
  Info: OS/Arch: darwin/arm64
  ✔ git is installed (will be used for faster uploads)
  ✔ Local config directory is writable (~/.onebuild)
  ✔ api.github.com is reachable
  ✔ A GitHub session is saved
```

## به‌روز نگه‌داشتن OneBuild

هر بار که یه دستور مثل `onebuild build` رو می‌زنی، OneBuild یه چک سریع
(چند ثانیه، و اگه آفلاین باشی بی‌صدا ردش می‌کنه) نسبت به آخرین release این
ریپو انجام می‌ده. اگه نسخه‌ی جدیدتری بود، این رو می‌بینی:

```
  ⚠ A newer version (v1.1.0) is available. Run 'onebuild update' to update.
```

برای آپدیت:

```bash
onebuild update
```

این دستور باینری مناسب OS/معماری سیستمت رو از آخرین release گیت‌هاب دانلود
می‌کنه و باینری در حال اجرا رو جاش می‌ذاره — بدون نیاز به نصب مجدد یا دانلود
دستی.

## همه‌ی دستورات

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

## این ابزار چی و کجا روی سیستمت ذخیره می‌کنه

| مسیر | محتوا |
|---|---|
| `~/.onebuild/session.json` | اسم کاربری گیت‌هابت و توکن رمزنگاری‌شده |
| `~/.onebuild/local.key` | کلید رمزنگاری محلی که برای توکن بالا استفاده میشه |
| `~/.onebuild/history.json` | تاریخچه‌ی build هات |
| `~/OneBuild-output/ios-cert/` | کلید خصوصی، CSR، گواهی و provisioning profile از `onebuild ios-cert` |
| `~/OneBuild-output/` | فایل‌های خروجی دانلودشده |
| `~/Desktop/` | گزارش‌های PDF خطا، هروقت درخواستشون بدی |

هیچی از این‌ها جایی فرستاده نمیشه، به‌جز تماس مستقیم HTTPS با
`api.github.com` (و فقط موقع `onebuild update`، دانلود باینری جدید از
Releases همین ریپو).

> **یه نکته درباره‌ی ذخیره‌ی محلی توکن**: توکن گیت‌هاب ذخیره‌شده با یه کلید
> که خودش هم محلی و کنارش ساخته میشه (`~/.onebuild/local.key`) رمزنگاری
> شده. این از دیدن تصادفی توکن (مثلاً باز کردن فایل توی ادیتور متن) جلوگیری
> می‌کنه، ولی جایگزین رمزنگاری کامل دیسک نیست — هرکسی که به همون سطح از
> اکانت کاربری‌ت دسترسی داشته باشه که بتونه فایل رمزشده رو بخونه، می‌تونه
> کلید کنارش رو هم بخونه. با `~/.onebuild/` مثل هر credential دیگه‌ای روی
> سیستمت رفتار کن.

## سوالات متداول / رفع اشکال

**نیاز به پلن پولی گیت‌هاب دارم؟**
نه. ریپوهای public دقیقه‌ی Actions رایگان نامحدود دارن؛ ریپوهای private توی
پلن رایگان یه سقف ماهانه دارن (در حال حاضر ۲۰۰۰ دقیقه در ماه) که معمولاً
برای پروژه‌های شخصی کافیه. بیلد چندتا پلتفرم همزمان، خصوصاً macOS/iOS، سریع‌تر
از Android/Web دقیقه مصرف می‌کنه.

**چرا اولین build فقط برای APK اندروید ۱۳ دقیقه طول کشید؟**
کاملاً عادیه، حتی نسبتاً سریع. runner های گیت‌هاب هر بار از یه سیستم کاملاً
تازه شروع می‌کنن — نصب Flutter SDK، دانلود Gradle wrapper، اجزای Android SDK،
و وابستگی‌های پروژه‌ت همه توی همون اولین اجرا از صفر انجام میشن. build های
بعدی همون پروژه معمولاً یه‌کم سریع‌تره چون پکیج‌های pub کش میشن، هرچند خود
Gradle هنوز بین اجراهای جدا کش نمیشه.

**می‌تونم برای ریپوی یه شرکت/سازمان استفاده کنم؟**
بله — موقع پرسیدن ریپو، گزینه‌ی «ریپوی موجود روی گیت‌هاب» رو انتخاب کن و
آدرسش رو بده، به‌شرطی که توکنت بهش دسترسی داشته باشه.

**نسخه‌ی واقعی Flutter/Xcode/Gradle از کجا میاد؟**
از هرچی که `subosito/flutter-action` روی runner گیت‌هاب موقع build نصب کنه
(پیش‌فرض کانال `stable`) — همون ابزاری که اکثر pipeline های CI فلاتر ازش
استفاده می‌کنن.

**می‌تونم فایل workflow ساخته‌شده رو دستی ادیت کنم؟**
بله، یه فایل عادیه توی `.github/workflows/onebuild.yml`. ولی اگه دوباره
`onebuild build` رو روی همون پروژه بزنی، این فایل رو با نسخه‌ی تازه‌ای که
از جواب‌های جدیدت ساخته میشه overwrite می‌کنه — اگه دستی ادیتش کرده باشی
این رو یادت باشه.

**آپلودم با خطای دسترسی fail شد.**
احتمالاً توکنت منقضی شده یا scope درست رو نداره. `onebuild logout` بزن،
بعد دوباره `onebuild auth login` با یه توکن تازه که scope های `repo` و
`workflow` رو داشته باشه.

## توسعه برای زبان‌های دیگه

نسخه‌ی ۱.۰.۰ کاملاً روی فلاتر تمرکز داره، ولی طراحی طوریه که این فرض جداست:
تعریف target ها و تولید YAML گیت‌هاب اکشن همه توی `internal/workflow/`
هستن، جدا از کد GitHub/آپلود/تاریخچه/UI. اضافه کردن پشتیبانی از یه فریم‌ورک
دیگه (React Native، اندروید/iOS native خالص، و...) یعنی یه مجموعه target
جدید اونجا اضافه کنی، بدون اینکه به بقیه‌ی ابزار دست بزنی.

## لایسنس

MIT — به [LICENSE](LICENSE) نگاه کن.

ساخته‌ی **A.M.Ghaderi**.
