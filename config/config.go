package config

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"github.com/mexirica/decentralized-file-signature/signer"
	"os"
)

const settingsFileName = "settings.json"

func createDefaultSettingsFile(downloadpath string) (string, error) {
	err := signer.InitKeys()
	if err != nil {
		return "", err
	}

	privateKeyBytes, err := json.Marshal(&signer.PrivateKey)
	publicKeyBytes, err := json.Marshal(&signer.PublicKey)

	defaultData := map[string]string{
		"downloadpath": downloadpath,
		"privateKey":   string(privateKeyBytes),
		"publicKey":    string(publicKeyBytes),
	}

	// Converte os dados para JSON
	jsonData, err := json.MarshalIndent(defaultData, "", "    ")
	if err != nil {
		return "", err
	}

	// Cria o arquivo settings.json
	if err := os.WriteFile(settingsFileName, jsonData, 0644); err != nil {
		return "", err
	}

	return "", nil
}

func ReadSettingsFile() (string, error) {
	data, err := os.ReadFile(settingsFileName)
	if err != nil {
		return "", err
	}

	var settings map[string]string
	if err := json.Unmarshal(data, &settings); err != nil {
		return "", err
	}

	// Verifica a existência da chave downloadpath
	downloadPath, ok := settings["downloadpath"]
	if !ok {
		return "", fmt.Errorf("downloadpath key not found in settings file")
	}

	// Lê as chaves privadas e públicas
	privateKeyStr, privateKeyOk := settings["privateKey"]
	publicKeyStr, publicKeyOk := settings["publicKey"]
	if privateKeyOk && publicKeyOk {
		// Desserializa as chaves
		var privateKey rsa.PrivateKey
		var publicKey rsa.PublicKey
		json.Unmarshal([]byte(privateKeyStr), &privateKey)
		json.Unmarshal([]byte(publicKeyStr), &publicKey)

		// Atribui as chaves ao pacote signer
		signer.PrivateKey = &privateKey
		signer.PublicKey = &publicKey
	}

	return downloadPath, nil
}

func UpdateDownloadPath(newPath string) error {
	data, err := os.ReadFile(settingsFileName)
	if err != nil {
		return err
	}

	var settings map[string]string
	if err := json.Unmarshal(data, &settings); err != nil {
		return err
	}

	// Atualiza o caminho de download
	settings["downloadpath"] = newPath

	// Converte de volta para JSON
	jsonData, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	// Escreve de volta no arquivo
	if err := os.WriteFile(settingsFileName, jsonData, 0644); err != nil {
		return err
	}

	fmt.Println("Download path updated successfully.")
	return nil
}

// CheckSettingsFileIntegrity checks if the settings file exists and contains all required keys.
func CheckSettingsFileIntegrity() error {
	// Verifica se o arquivo settings.json existe
	if _, err := os.Stat(settingsFileName); os.IsNotExist(err) {
		// Se o arquivo não existir, cria com valores padrão
		_, err := createDefaultSettingsFile("")
		return err
	}

	// Se o arquivo existe, verificamos as chaves
	data, err := os.ReadFile(settingsFileName)
	if err != nil {
		return err
	}

	var settings map[string]string
	if err := json.Unmarshal(data, &settings); err != nil {
		return err
	}

	// Checa se todas as chaves obrigatórias estão presentes
	requiredKeys := []string{"downloadpath", "privateKey", "publicKey"}
	for _, key := range requiredKeys {
		if _, ok := settings[key]; !ok {
			_, err := createDefaultSettingsFile(settings["downloadpath"])
			return err
		}
	}

	return nil
}
