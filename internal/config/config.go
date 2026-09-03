package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"onebuild/internal/crypto"
)

const dirName = ".onebuild"

func BaseDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, dirName)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return dir, nil
}

func sessionPath() (string, error) {
	dir, err := BaseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "session.json"), nil
}

func keyPath() (string, error) {
	dir, err := BaseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "local.key"), nil
}

type Session struct {
	Login          string `json:"login"`
	Name           string `json:"name"`
	TokenEncrypted string `json:"token_encrypted"`
}

func SaveSession(login, name, token string) error {
	kp, err := keyPath()
	if err != nil {
		return err
	}
	key, err := crypto.LoadOrCreateKey(kp)
	if err != nil {
		return err
	}
	encrypted, err := crypto.Encrypt(key, token)
	if err != nil {
		return err
	}
	sess := Session{Login: login, Name: name, TokenEncrypted: encrypted}
	data, err := json.MarshalIndent(sess, "", "  ")
	if err != nil {
		return err
	}
	sp, err := sessionPath()
	if err != nil {
		return err
	}
	return os.WriteFile(sp, data, 0600)
}

func LoadSession() (*Session, error) {
	sp, err := sessionPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(sp)
	if err != nil {
		return nil, err
	}
	var sess Session
	if err := json.Unmarshal(data, &sess); err != nil {
		return nil, err
	}
	return &sess, nil
}

func LoadToken() (string, error) {
	sess, err := LoadSession()
	if err != nil {
		return "", err
	}
	kp, err := keyPath()
	if err != nil {
		return "", err
	}
	key, err := crypto.LoadOrCreateKey(kp)
	if err != nil {
		return "", err
	}
	token, err := crypto.Decrypt(key, sess.TokenEncrypted)
	if err != nil {
		return "", errors.New("could not decrypt saved token, please login again")
	}
	return token, nil
}

func ClearSession() error {
	sp, err := sessionPath()
	if err != nil {
		return err
	}
	if _, err := os.Stat(sp); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(sp)
}

func HasSession() bool {
	sp, err := sessionPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(sp)
	return err == nil
}
