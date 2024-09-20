package ipfs

import (
	"bytes"
	"encoding/json"
	"fmt"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/mexirica/decentralized-file-signature/config"
	"github.com/mexirica/decentralized-file-signature/pkg/signer"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// FileInfo represents metadata for a file, including its name, size, CID, and cryptographic signature.
type FileInfo struct {
	Name      string `json:"name"`
	Size      int64  `json:"size"`
	CID       string `json:"cid"`
	Signature string `json:"signature"`
}

// AddFile prompts the user for a file path, signs the file content, and uploads it to IPFS.
// It saves the file's metadata (name, size, CID, and signature) to a JSON file.
func AddFile(ipfs *shell.Shell) {
	fmt.Print("Enter the file path: ")
	var filePath string
	fmt.Scanln(&filePath)

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening the file:", err)
		return
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Println("Error getting file information:", err)
		return
	}
	content, err := io.ReadAll(file)

	signature, err := signer.Sign(content)
	if err != nil {
		fmt.Println("Error signing the file:", err)
		return
	}

	// Add the signed content to IPFS
	cid, err := ipfs.Add(bytes.NewReader(content))
	if err != nil {
		fmt.Println("Error adding file to IPFS:", err)
		return
	}

	fmt.Printf("File successfully added! CID: %s\n", cid)

	saveFileInfo(fileInfo.Name(), fileInfo.Size(), cid, signature)
}

// saveFileInfo stores the metadata of the uploaded file (name, size, CID, and signature)
// into a JSON file named `files.json`.
func saveFileInfo(name string, size int64, cid string, signature string) {
	newFileInfo := FileInfo{
		Name:      name,
		Size:      size,
		CID:       cid,
		Signature: signature,
	}

	filename := "files.json"
	err := addFileInfo(filename, newFileInfo)
	if err != nil {
		fmt.Println("Error saving file info:", err)
	}
}

// ListFiles reads and displays the metadata (name, size, CID, and signature)
// of all files stored in the `files.json` file.
func ListFiles() {
	fileInfoList, err := loadFileInfo("files.json")
	if err != nil {
		fmt.Println("Error loading files:", err)
		return
	}

	if len(fileInfoList) == 0 {
		fmt.Println("No files have been added yet.")
		return
	}

	for _, fileInfo := range fileInfoList {
		fmt.Printf("Name: %s, Size: %dB, CID: %s, Signature: %s\n", fileInfo.Name, fileInfo.Size, fileInfo.CID, fileInfo.Signature)
	}
}

// GetFileInfo prompts the user for a file's CID and retrieves the corresponding metadata
// (name, size, and signature) from the `files.json` file.
func GetFileInfo() {
	fmt.Print("Enter the file CID: ")
	var cid string
	fmt.Scanln(&cid)

	fileInfoList, err := loadFileInfo("files.json")
	if err != nil {
		fmt.Println("Error loading files:", err)
		return
	}

	for _, fileInfo := range fileInfoList {
		if fileInfo.CID == cid {
			fmt.Printf("Name: %s, Size: %dB, Signature: %s\n", fileInfo.Name, fileInfo.Size, fileInfo.Signature)
			return
		}
	}

	fmt.Println("File not found.")
}

// RetrieveFileContent retrieves the content of a file from IPFS using its CID and displays it.
func RetrieveFileContent(ipfs *shell.Shell) {
	fmt.Print("Enter the file CID: ")
	var cid string
	fmt.Scanln(&cid)

	r, err := ipfs.Cat(cid)
	if err != nil {
		fmt.Println("Error retrieving file content:", err)
		return
	}

	content, err := io.ReadAll(r)
	if err != nil {
		fmt.Println("Error reading file content:", err)
		return
	}

	fmt.Printf("File content:\n%s\n", string(content))
}

// DownloadFile downloads a file from IPFS using its CID and saves it locally in the download path
// specified in the configuration settings.
func DownloadFile(ipfs *shell.Shell) {
	fmt.Print("Enter the file CID: ")
	var cid string
	fmt.Scanln(&cid)

	r, err := ipfs.Cat(cid)
	if err != nil {
		fmt.Println("Error retrieving file content:", err)
		return
	}
	defer r.Close()

	fileInfoList, err := loadFileInfo("files.json")
	if err != nil {
		fmt.Println("Error loading file info:", err)
		return
	}

	var fileName string
	for _, fileInfo := range fileInfoList {
		if fileInfo.CID == cid {
			fileName = fileInfo.Name
			break
		}
	}

	if fileName == "" {
		fmt.Println("File information not found")
		return
	}

	fullPath := filepath.Join(filepath.Dir(config.DownloadPath), fileName)
	outFile, err := os.Create(fullPath)
	if err != nil {
		fmt.Println("Error creating the local file:", err)
		return
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, r)
	if err != nil {
		fmt.Println("Error saving file to local system:", err)
		return
	}

	fmt.Printf("File downloaded successfully to %s\n", fullPath)
}

// addFileInfo appends a new FileInfo record to the existing `files.json` file,
// ensuring that previous records are preserved.
func addFileInfo(filename string, newFileInfo FileInfo) error {
	fileInfoList, err := loadFileInfo(filename)
	if err != nil {
		return err
	}

	fileInfoList = append(fileInfoList, newFileInfo)
	return saveFileInfos(filename, fileInfoList)
}

// saveFileInfos writes a list of FileInfo records to a JSON file.
func saveFileInfos(filename string, fileInfoList []FileInfo) error {
	data, err := json.MarshalIndent(fileInfoList, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

// loadFileInfo reads and returns the list of FileInfo records from the specified JSON file.
// If the file does not exist, it returns an empty list.
func loadFileInfo(filename string) ([]FileInfo, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return []FileInfo{}, nil
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var fileInfoList []FileInfo
	err = json.Unmarshal(data, &fileInfoList)
	if err != nil {
		return nil, err
	}

	return fileInfoList, nil
}

// ClearScreen clears the terminal screen based on the operating system (Windows, Linux, or MacOS).
func ClearScreen() {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux", "darwin":
		cmd = exec.Command("clear")
	case "windows":
		cmd = exec.Command("cmd", "/c", "cls")
	default:
		fmt.Println("Unsupported platform, cannot clear screen.")
		return
	}

	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		fmt.Println("Failed to clear terminal:", err)
	}
}

// VerifyIpfsFileIntegrity verifies the integrity of a file in IPFS by comparing its content's cryptographic signature.
func VerifyIpfsFileIntegrity(ipfs *shell.Shell) {
	var cid string
	var signature string

	fmt.Println("Enter the file CID:")
	fmt.Scanln(&cid)

	fmt.Println("Enter the file signature:")
	fmt.Scanln(&signature)

	cid = strings.TrimSpace(cid)
	signature = strings.TrimSpace(signature)

	r, err := ipfs.Cat(cid)
	if err != nil {
		fmt.Println("Error retrieving file content:", err)
		return
	}
	defer r.Close()

	content, err := io.ReadAll(r)
	if err != nil {
		fmt.Println("Error reading file content:", err)
		return
	}

	original := signer.Verify(content, signature)
	if original {
		fmt.Println("File integrity confirmed.")
	} else {
		fmt.Println("File integrity not confirmed.")
	}
}
