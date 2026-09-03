package ioscert

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func OpenSSLAvailable() bool {
	_, err := exec.LookPath("openssl")
	return err == nil
}

func ConvertDERToPEM(derCertPath, outDir string) (string, error) {
	pemPath := filepath.Join(outDir, "ios_distribution.cert.pem")
	cmd := exec.Command("openssl", "x509", "-inform", "DER", "-in", derCertPath, "-out", pemPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("openssl x509 failed: %s: %w", string(out), err)
	}
	return pemPath, nil
}

func ExportP12(certPEMPath, keyPEMPath, password, outDir string) (string, error) {
	p12Path := filepath.Join(outDir, "ios_distribution.p12")
	cmd := exec.Command("openssl", "pkcs12", "-export",
		"-inkey", keyPEMPath,
		"-in", certPEMPath,
		"-out", p12Path,
		"-passout", "pass:"+password,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("openssl pkcs12 failed: %s: %w", string(out), err)
	}
	return p12Path, nil
}

func Base64File(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

func ManualInstructions(derCertPath, keyPEMPath, outDir string) string {
	pemPath := filepath.Join(outDir, "ios_distribution.cert.pem")
	p12Path := filepath.Join(outDir, "ios_distribution.p12")
	return fmt.Sprintf(
		"openssl was not found on this machine. Install it (it's preinstalled on\n"+
			"macOS and Linux; on Windows, Git for Windows or WSL both include it),\n"+
			"then run these two commands yourself:\n\n"+
			"  openssl x509 -inform DER -in %q -out %q\n"+
			"  openssl pkcs12 -export -inkey %q -in %q -out %q -passout pass:YOUR_PASSWORD\n\n"+
			"Then base64-encode %q and use that as the IOS_CERTIFICATE_BASE64 secret,\n"+
			"and YOUR_PASSWORD as the IOS_CERTIFICATE_PASSWORD secret.",
		derCertPath, pemPath,
		keyPEMPath, pemPath, p12Path,
		p12Path,
	)
}
