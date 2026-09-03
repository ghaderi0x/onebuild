package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"onebuild/internal/ioscert"
	"onebuild/internal/ui"
)

func iosCertOutDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	dir := filepath.Join(home, "OneBuild-output", "ios-cert")
	os.MkdirAll(dir, 0700)
	return dir
}

func runIOSCert(args []string) {
	if len(args) == 0 {
		ui.Warn("Specify a subcommand: csr, or package")
		return
	}
	switch args[0] {
	case "csr":
		runIOSCertCSR()
	case "package":
		runIOSCertPackage()
	case "encode":
		runIOSCertEncode(args[1:])
	default:
		ui.Warn("Unknown ios-cert subcommand: %s", args[0])
	}
}

func runIOSCertEncode(args []string) {
	ui.Banner()
	ui.Step("Base64-encode a file for a GitHub secret")

	var path string
	if len(args) > 0 {
		path = args[0]
	} else {
		path = ui.AskString("Path to the file (e.g. your .mobileprovision)", "")
	}
	if _, err := os.Stat(path); err != nil {
		ui.Error("Could not find that file: %s", path)
		os.Exit(1)
	}

	b64, err := ioscert.Base64File(path)
	if err != nil {
		ui.Error("Could not read/encode the file: %s", err.Error())
		os.Exit(1)
	}

	outDir := iosCertOutDir()
	outPath := filepath.Join(outDir, filepath.Base(path)+".base64.txt")
	if err := os.WriteFile(outPath, []byte(b64), 0600); err != nil {
		ui.Error("Could not save the base64 file: %s", err.Error())
		os.Exit(1)
	}

	ui.Success("base64 saved to: %s", outPath)
	fmt.Println()
	fmt.Println("  Paste the contents of that file as the matching GitHub secret")
	fmt.Println("  (for a .mobileprovision file, that's IOS_PROVISIONING_PROFILE_BASE64).")
}

func runIOSCertCSR() {
	ui.Banner()
	ui.Step("Generate an Apple certificate request (no Mac needed)")

	fmt.Println("  This creates a private key and a Certificate Signing Request")
	fmt.Println("  (CSR) locally, using the same standard everyone else uses —")
	fmt.Println("  including Keychain Access on a Mac. Apple can't tell the")
	fmt.Println("  difference.")
	fmt.Println()

	email := ui.AskString("Your Apple ID email", "")
	name := ui.AskString("Your full name (as it should appear on the certificate)", "")
	country := ui.AskString("Two-letter country code", "US")

	outDir := iosCertOutDir()
	existingKey := filepath.Join(outDir, "ios_distribution.key.pem")
	if _, statErr := os.Stat(existingKey); statErr == nil {
		ui.Warn("A private key already exists at %s", existingKey)
		fmt.Println("  Overwriting it means any certificate Apple already issued for the")
		fmt.Println("  old key won't be usable anymore (the key and certificate must match).")
		if !ui.AskYesNo("Generate a new key and CSR anyway?", false) {
			return
		}
	}

	result, err := ioscert.GenerateKeyAndCSR(outDir, email, name, country)
	if err != nil {
		ui.Error("Could not generate the CSR: %s", err.Error())
		os.Exit(1)
	}

	ui.Success("Private key saved to:  %s", result.KeyPath)
	ui.Success("CSR saved to:          %s", result.CSRPath)
	fmt.Println()
	fmt.Println("  Keep the private key file safe and don't share it — you'll")
	fmt.Println("  need it again in the next step.")
	fmt.Println()
	fmt.Println("  Next:")
	fmt.Println("   1. Go to https://developer.apple.com/account/resources/certificates/add")
	fmt.Println("   2. Choose \"Apple Distribution\" (or \"iOS Distribution\")")
	fmt.Println("   3. Upload the CSR file above when asked for one")
	fmt.Println("   4. Download the resulting certificate (a .cer file)")
	fmt.Println("   5. Run: onebuild ios-cert package")
}

func runIOSCertPackage() {
	ui.Banner()
	ui.Step("Package your certificate into a .p12")

	outDir := iosCertOutDir()
	defaultKey := filepath.Join(outDir, "ios_distribution.key.pem")

	certPath := ui.AskString("Path to the .cer file you downloaded from Apple", "")
	if _, err := os.Stat(certPath); err != nil {
		ui.Error("Could not find that file: %s", certPath)
		os.Exit(1)
	}

	keyPath := ui.AskString("Path to your private key", defaultKey)
	if _, err := os.Stat(keyPath); err != nil {
		ui.Error("Could not find that file: %s", keyPath)
		os.Exit(1)
	}

	if !ioscert.OpenSSLAvailable() {
		ui.Warn("openssl was not found on this machine.")
		fmt.Println()
		fmt.Println(ioscert.ManualInstructions(certPath, keyPath, outDir))
		return
	}

	password := ui.AskSecret("Choose a password to protect the .p12 (this becomes IOS_CERTIFICATE_PASSWORD)")
	if password == "" {
		ui.Error("A password is required")
		os.Exit(1)
	}

	spinner := ui.NewSpinner()
	spinner.Start("Converting certificate...")
	pemPath, convErr := ioscert.ConvertDERToPEM(certPath, outDir)
	spinner.Stop()
	if convErr != nil {
		ui.Error("Could not convert the certificate: %s", convErr.Error())
		ui.Info("If your certificate is already in PEM format, this step normally isn't needed — you can adapt the manual openssl pkcs12 command yourself.")
		os.Exit(1)
	}

	spinner.Start("Building the .p12...")
	p12Path, expErr := ioscert.ExportP12(pemPath, keyPath, password, outDir)
	spinner.Stop()
	if expErr != nil {
		ui.Error("Could not build the .p12: %s", expErr.Error())
		os.Exit(1)
	}

	b64, b64Err := ioscert.Base64File(p12Path)
	if b64Err != nil {
		ui.Error("Could not base64-encode the .p12: %s", b64Err.Error())
		os.Exit(1)
	}

	b64Path := filepath.Join(outDir, "ios_distribution.p12.base64.txt")
	if err := os.WriteFile(b64Path, []byte(b64), 0600); err != nil {
		ui.Error("Could not save the base64 file: %s", err.Error())
		os.Exit(1)
	}

	ui.Success(".p12 saved to:        %s", p12Path)
	ui.Success("base64 saved to:      %s", b64Path)
	fmt.Println()
	fmt.Println("  Add these two GitHub secrets on your repository:")
	fmt.Println("    IOS_CERTIFICATE_BASE64    -> the contents of the base64 file above")
	fmt.Println("    IOS_CERTIFICATE_PASSWORD  -> the password you just chose")
	fmt.Println()
	fmt.Println("  You'll still need a provisioning profile matching this certificate")
	fmt.Println("  (download it from developer.apple.com and base64-encode it the same")
	fmt.Println("  way) for IOS_PROVISIONING_PROFILE_BASE64, plus any value for")
	fmt.Println("  KEYCHAIN_PASSWORD. onebuild build will tell you the exact names")
	fmt.Println("  and check they're all present before starting a signed iOS build.")
}
