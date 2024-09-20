package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/mexirica/decentralized-file-signature/cmd/ipfs"
	"github.com/mexirica/decentralized-file-signature/pkg/signer"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"strings"
)

const settingsFileName = "settings.json"

var DownloadPath string

// init initializes the configuration by checking the settings file integrity
// and loading necessary values such as the download path and cryptographic keys.
// If the settings file does not exist or is incomplete, it prompts the user to set the download path.
func init() {
	err := checkSettingsFileIntegrity()
	if err != nil {
		fmt.Println("Error checking settings file integrity:", err)
		return
	}

	path, err := readSettingsFile()
	if err != nil {
		return
	}

	if path == "" {
		for {
			fmt.Println("Enter the path where the files will be downloaded:")
			fmt.Scanln(&DownloadPath)
			err := isValidPath(DownloadPath)
			if err != nil {
				fmt.Println("Error:", err)
			} else {
				break
			}
		}

		// Update the download path in the configuration file
		if err = updateDownloadPath(DownloadPath); err != nil {
			fmt.Println("Error updating settings file:", err)
			return
		}
		ipfs.ClearScreen()
	} else {
		DownloadPath = path
	}
}

// CheckSettingsFileIntegrity verifies the existence and integrity of the settings file.
// If the file is missing or required keys are absent, it creates a default settings file.
// It checks for keys such as "downloadpath", "privateKey", and "publicKey".
func checkSettingsFileIntegrity() error {
	if _, err := os.Stat(settingsFileName); os.IsNotExist(err) {
		// Create the settings file with default values
		_, err := createDefaultSettingsFile("")
		return err
	}

	// Check if all required keys are present in the settings file
	data, err := os.ReadFile(settingsFileName)
	if err != nil {
		return err
	}

	var settings map[string]string
	if err := json.Unmarshal(data, &settings); err != nil {
		return err
	}

	requiredKeys := []string{"downloadpath", "privateKey", "publicKey"}
	for _, key := range requiredKeys {
		if _, ok := settings[key]; !ok {
			_, err := createDefaultSettingsFile(settings["downloadpath"])
			return err
		}
	}

	return nil
}

// createDefaultSettingsFile creates a new settings file with default values,
// including the provided download path and newly generated RSA private and public keys.
func createDefaultSettingsFile(downloadpath string) (string, error) {
	err := signer.InitKeys()
	if err != nil {
		return "", err
	}

	privateKeyPEM, err := marshalPrivateKey(signer.PrivateKey)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal private key")
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: x509.MarshalPKCS1PublicKey(signer.PublicKey)})

	defaultData := map[string]string{
		"downloadpath": downloadpath,
		"privateKey":   string(privateKeyPEM),
		"publicKey":    string(publicKeyPEM),
	}

	jsonData, err := json.MarshalIndent(defaultData, "", "    ")
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal JSON")
	}

	if err := os.WriteFile(settingsFileName, jsonData, 0644); err != nil {
		return "", errors.Wrap(err, "failed to write settings file")
	}

	return "", nil
}

// ReadSettingsFile reads the settings file and returns the configured download path.
// It also loads the cryptographic keys (private and public) if they exist in the file.
func readSettingsFile() (string, error) {
	data, err := os.ReadFile(settingsFileName)
	if err != nil {
		return "", fmt.Errorf("failed to read settings file")
	}

	var settings map[string]string
	err = json.Unmarshal(data, &settings)
	if err != nil {
		return "", errors.Wrap(err, "failed to unmarshal JSON")
	}

	downloadPath, ok := settings["downloadpath"]
	if !ok {
		return "", fmt.Errorf("downloadpath key not found in settings file")
	}

	privateKeyStr, privateKeyOk := settings["privateKey"]
	publicKeyStr, publicKeyOk := settings["publicKey"]
	if privateKeyOk && publicKeyOk {
		err = loadKeysFromPEM(privateKeyStr, publicKeyStr)
		if err != nil {
			return "", errors.Wrap(err, "failed to load keys from PEM")
		}
	}

	return downloadPath, nil
}

// UpdateDownloadPath updates the download path in the settings file with the provided new path.
func updateDownloadPath(newPath string) error {
	data, err := os.ReadFile(settingsFileName)
	if err != nil {
		return err
	}

	var settings map[string]string
	if err := json.Unmarshal(data, &settings); err != nil {
		return err
	}

	// Update the download path
	settings["downloadpath"] = newPath

	// Convert the updated settings back to JSON
	jsonData, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	// Write the updated settings back to the file
	if err := os.WriteFile(settingsFileName, jsonData, 0644); err != nil {
		return err
	}

	return nil
}

// marshalPrivateKey converts an RSA private key to PEM format.
func marshalPrivateKey(privateKey *rsa.PrivateKey) ([]byte, error) {
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privateKeyBytes}), nil
}

// loadKeysFromPEM loads the private and public RSA keys from PEM strings and sets them in the signer package.
func loadKeysFromPEM(privateKeyStr, publicKeyStr string) error {
	privateKeyBlock, _ := pem.Decode([]byte(privateKeyStr))
	if privateKeyBlock == nil || privateKeyBlock.Type != "RSA PRIVATE KEY" {
		return fmt.Errorf("invalid private key PEM")
	}

	publicKeyBlock, _ := pem.Decode([]byte(publicKeyStr))
	if publicKeyBlock == nil || publicKeyBlock.Type != "RSA PUBLIC KEY" {
		return fmt.Errorf("invalid public key PEM")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(privateKeyBlock.Bytes)
	if err != nil {
		return errors.Wrap(err, "failed to parse private key")
	}

	var publicKey *rsa.PublicKey
	// Try parsing the public key as PKIX first, and fallback to PKCS1 if it fails
	publicKeyInterface, err := x509.ParsePKIXPublicKey(publicKeyBlock.Bytes)
	if err == nil {
		var ok bool
		publicKey, ok = publicKeyInterface.(*rsa.PublicKey)
		if !ok {
			return fmt.Errorf("not an RSA public key (PKIX)")
		}
	} else {
		publicKey, err = x509.ParsePKCS1PublicKey(publicKeyBlock.Bytes)
		if err != nil {
			return errors.Wrap(err, "failed to parse public key")
		}
	}

	// Set the keys in the signer package
	signer.PrivateKey = privateKey.(*rsa.PrivateKey)
	signer.PublicKey = publicKey

	return nil
}

// isValidPath checks if a given path is a valid directory with write permissions.
func isValidPath(path string) error {
	// Remove leading/trailing spaces
	path = strings.TrimSpace(path)

	// Check if the path is empty
	if path == "" {
		return fmt.Errorf("the path cannot be empty")
	}

	// Check if the path exists
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("the path does not exist")
	}

	// Check if the path is a directory
	if !info.IsDir() {
		return fmt.Errorf("the path is not a directory")
	}

	// Check if the directory is writable
	testFile := filepath.Join(path, "testfile")
	file, err := os.Create(testFile)
	if err != nil {
		return fmt.Errorf("cannot write to the directory")
	}
	file.Close()
	os.Remove(testFile) // Remove the test file

	return nil
}
