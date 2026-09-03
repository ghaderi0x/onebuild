package workflow

import (
	"fmt"
	"strings"
)

const WorkflowPath = ".github/workflows/onebuild.yml"

const (
	TargetAndroidAPK    = "android-apk"
	TargetAndroidBundle = "android-appbundle"
	TargetIOSUnsigned   = "ios-unsigned"
	TargetIOSSigned     = "ios-signed"
	TargetWeb           = "web"
	TargetWindows       = "windows"
	TargetLinux         = "linux"
	TargetMacOS         = "macos"
)

var TargetLabels = map[string]string{
	TargetAndroidAPK:    "Android (.apk)",
	TargetAndroidBundle: "Android App Bundle (.aab)",
	TargetIOSUnsigned:   "iOS - unsigned build (.ipa, needs resigning)",
	TargetIOSSigned:     "iOS - signed with your certificate (.ipa)",
	TargetWeb:           "Web",
	TargetWindows:       "Windows desktop",
	TargetLinux:         "Linux desktop",
	TargetMacOS:         "macOS desktop",
}

var TargetOrder = []string{
	TargetAndroidAPK,
	TargetAndroidBundle,
	TargetIOSUnsigned,
	TargetIOSSigned,
	TargetWeb,
	TargetWindows,
	TargetLinux,
	TargetMacOS,
}

var IOSSignedRequiredSecrets = []string{
	"IOS_CERTIFICATE_BASE64",
	"IOS_CERTIFICATE_PASSWORD",
	"IOS_PROVISIONING_PROFILE_BASE64",
	"KEYCHAIN_PASSWORD",
}

type Options struct {
	Branch         string
	FlutterVersion string
	Targets        []string
	IOSTeamID      string
	IOSExportMethod string
}

func flutterVersionOrDefault(v string) string {
	if strings.TrimSpace(v) == "" {
		return "stable"
	}
	return v
}

func setupSteps(b *strings.Builder, flutterVersion string) {
	b.WriteString("      - uses: actions/checkout@v4\n")
	b.WriteString("      - uses: subosito/flutter-action@v2\n")
	b.WriteString("        with:\n")
	if flutterVersion == "stable" || flutterVersion == "beta" || flutterVersion == "master" {
		b.WriteString(fmt.Sprintf("          channel: %s\n", flutterVersion))
	} else {
		b.WriteString(fmt.Sprintf("          flutter-version: %s\n", flutterVersion))
	}
	b.WriteString("          cache: true\n")
	b.WriteString("      - run: flutter pub get\n")
}

func uploadStep(b *strings.Builder, name, path string) {
	b.WriteString("      - uses: actions/upload-artifact@v4\n")
	b.WriteString("        if: always()\n")
	b.WriteString(fmt.Sprintf("        with:\n          name: %s\n          path: %s\n          if-no-files-found: warn\n", name, path))
}

func writeJobHeader(b *strings.Builder, jobKey, name, runsOn string) {
	b.WriteString(fmt.Sprintf("  %s:\n", jobKey))
	b.WriteString(fmt.Sprintf("    name: %s\n", name))
	b.WriteString(fmt.Sprintf("    runs-on: %s\n", runsOn))
	b.WriteString("    steps:\n")
}

func Generate(opts Options) string {
	version := flutterVersionOrDefault(opts.FlutterVersion)
	branch := opts.Branch
	if branch == "" {
		branch = "main"
	}

	var b strings.Builder
	b.WriteString("name: OneBuild\n\n")
	b.WriteString("on:\n")
	b.WriteString("  push:\n")
	b.WriteString(fmt.Sprintf("    branches: [ \"%s\" ]\n", branch))
	b.WriteString("  workflow_dispatch: {}\n\n")
	b.WriteString("jobs:\n")

	targets := map[string]bool{}
	for _, t := range opts.Targets {
		targets[t] = true
	}

	if targets[TargetAndroidAPK] {
		writeJobHeader(&b, "android_apk", TargetLabels[TargetAndroidAPK], "ubuntu-latest")
		setupSteps(&b, version)
		b.WriteString("      - run: flutter build apk --release\n")
		uploadStep(&b, "app-android-apk", "build/app/outputs/flutter-apk/*.apk")
		b.WriteString("\n")
	}

	if targets[TargetAndroidBundle] {
		writeJobHeader(&b, "android_appbundle", TargetLabels[TargetAndroidBundle], "ubuntu-latest")
		setupSteps(&b, version)
		b.WriteString("      - run: flutter build appbundle --release\n")
		uploadStep(&b, "app-android-appbundle", "build/app/outputs/bundle/release/*.aab")
		b.WriteString("\n")
	}

	if targets[TargetIOSUnsigned] {
		writeJobHeader(&b, "ios_unsigned", TargetLabels[TargetIOSUnsigned], "macos-latest")
		setupSteps(&b, version)
		b.WriteString("      - run: flutter build ios --release --no-codesign\n")
		b.WriteString("      - name: Package unsigned ipa\n")
		b.WriteString("        run: |\n")
		b.WriteString("          set -e\n")
		b.WriteString("          cd build/ios/iphoneos\n")
		b.WriteString("          mkdir -p Payload\n")
		b.WriteString("          cp -r Runner.app Payload/Runner.app\n")
		b.WriteString("          zip -r app-unsigned.ipa Payload > /dev/null\n")
		uploadStep(&b, "app-ios-unsigned", "build/ios/iphoneos/app-unsigned.ipa")
		b.WriteString("\n")
	}

	if targets[TargetIOSSigned] {
		writeJobHeader(&b, "ios_signed", TargetLabels[TargetIOSSigned], "macos-latest")
		setupSteps(&b, version)
		b.WriteString("      - name: Import signing certificate and provisioning profile\n")
		b.WriteString("        env:\n")
		b.WriteString("          IOS_CERTIFICATE_BASE64: ${{ secrets.IOS_CERTIFICATE_BASE64 }}\n")
		b.WriteString("          IOS_CERTIFICATE_PASSWORD: ${{ secrets.IOS_CERTIFICATE_PASSWORD }}\n")
		b.WriteString("          IOS_PROVISIONING_PROFILE_BASE64: ${{ secrets.IOS_PROVISIONING_PROFILE_BASE64 }}\n")
		b.WriteString("          KEYCHAIN_PASSWORD: ${{ secrets.KEYCHAIN_PASSWORD }}\n")
		b.WriteString("        run: |\n")
		b.WriteString("          set -e\n")
		b.WriteString("          echo \"$IOS_CERTIFICATE_BASE64\" | base64 --decode > certificate.p12\n")
		b.WriteString("          security create-keychain -p \"$KEYCHAIN_PASSWORD\" build.keychain\n")
		b.WriteString("          security default-keychain -s build.keychain\n")
		b.WriteString("          security unlock-keychain -p \"$KEYCHAIN_PASSWORD\" build.keychain\n")
		b.WriteString("          security set-keychain-settings -lut 21600 build.keychain\n")
		b.WriteString("          security list-keychains -d user -s build.keychain $(security list-keychains -d user | sed 's/\"//g')\n")
		b.WriteString("          security import certificate.p12 -k build.keychain -P \"$IOS_CERTIFICATE_PASSWORD\" -T /usr/bin/codesign -T /usr/bin/security\n")
		b.WriteString("          security set-key-partition-list -S apple-tool:,apple:,codesign: -s -k \"$KEYCHAIN_PASSWORD\" build.keychain\n")
		b.WriteString("          mkdir -p \"$HOME/Library/MobileDevice/Provisioning Profiles\"\n")
		b.WriteString("          echo \"$IOS_PROVISIONING_PROFILE_BASE64\" | base64 --decode > \"$HOME/Library/MobileDevice/Provisioning Profiles/profile.mobileprovision\"\n")
		b.WriteString("      - name: Create ExportOptions.plist\n")
		b.WriteString("        run: |\n")
		b.WriteString("          cat > ios/ExportOptions.plist << 'PLIST_EOF'\n")
		b.WriteString("          <?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
		b.WriteString("          <!DOCTYPE plist PUBLIC \"-//Apple//DTD PLIST 1.0//EN\" \"http://www.apple.com/DTDs/PropertyList-1.0.dtd\">\n")
		b.WriteString("          <plist version=\"1.0\">\n")
		b.WriteString("          <dict>\n")
		method := opts.IOSExportMethod
		if method == "" {
			method = "ad-hoc"
		}
		b.WriteString(fmt.Sprintf("            <key>method</key>\n            <string>%s</string>\n", method))
		b.WriteString(fmt.Sprintf("            <key>teamID</key>\n            <string>%s</string>\n", opts.IOSTeamID))
		b.WriteString("            <key>signingStyle</key>\n            <string>manual</string>\n")
		b.WriteString("            <key>compileBitcode</key>\n            <false/>\n")
		b.WriteString("          </dict>\n")
		b.WriteString("          </plist>\n")
		b.WriteString("          PLIST_EOF\n")
		b.WriteString("      - run: flutter build ipa --release --export-options-plist=ios/ExportOptions.plist\n")
		uploadStep(&b, "app-ios-signed", "build/ios/ipa/*.ipa")
		b.WriteString("\n")
	}

	if targets[TargetWeb] {
		writeJobHeader(&b, "web", TargetLabels[TargetWeb], "ubuntu-latest")
		setupSteps(&b, version)
		b.WriteString("      - run: flutter build web --release\n")
		uploadStep(&b, "app-web", "build/web")
		b.WriteString("\n")
	}

	if targets[TargetWindows] {
		writeJobHeader(&b, "windows", TargetLabels[TargetWindows], "windows-latest")
		setupSteps(&b, version)
		b.WriteString("      - run: flutter build windows --release\n")
		uploadStep(&b, "app-windows", "build/windows/**/runner/Release/**")
		b.WriteString("\n")
	}

	if targets[TargetLinux] {
		writeJobHeader(&b, "linux", TargetLabels[TargetLinux], "ubuntu-latest")
		b.WriteString("      - name: Install Linux build dependencies\n")
		b.WriteString("        run: |\n")
		b.WriteString("          sudo apt-get update\n")
		b.WriteString("          sudo apt-get install -y clang cmake ninja-build pkg-config libgtk-3-dev liblzma-dev\n")
		setupSteps(&b, version)
		b.WriteString("      - run: flutter config --enable-linux-desktop\n")
		b.WriteString("      - run: flutter build linux --release\n")
		uploadStep(&b, "app-linux", "build/linux/**/release/bundle/**")
		b.WriteString("\n")
	}

	if targets[TargetMacOS] {
		writeJobHeader(&b, "macos", TargetLabels[TargetMacOS], "macos-latest")
		setupSteps(&b, version)
		b.WriteString("      - run: flutter config --enable-macos-desktop\n")
		b.WriteString("      - run: flutter build macos --release\n")
		uploadStep(&b, "app-macos", "build/macos/Build/Products/Release/*.app")
		b.WriteString("\n")
	}

	return b.String()
}
