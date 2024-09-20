package ipfs

import (
	"encoding/json"
	"fmt"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/mexirica/decentralized-file-signature/signer"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

type FileInfo struct {
	Name      string `json:"name"`
	Size      int64  `json:"size"`
	CID       string `json:"cid"`
	Signature string `json:"signature"`
}

func AddFile(ipfs *shell.Shell) {
	fmt.Print("Enter the file path: ")
	var filePath string
	fmt.Scanln(&filePath)

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening the file:", err)
		return
	}
	defer file.Close()

	// Get file information
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Println("Error getting file information:", err)
		return
	}

	content, err := io.ReadAll(file)

	signature, err := signer.Sign(content)

	// Add the file to IPFS
	cid, err := ipfs.Add(file)
	if err != nil {
		fmt.Println("Error adding file to IPFS:", err)
		return
	}

	fmt.Printf("File successfully added! CID: %s\n", cid)

	// Save the file info to the JSON
	saveFileInfo(fileInfo.Name(), fileInfo.Size(), cid, signature)
}

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

func DownloadFile(ipfs *shell.Shell, downloadPath string) {
	fmt.Print("Enter the file CID: ")
	var cid string
	fmt.Scanln(&cid)

	// Retrieve the file content from IPFS
	r, err := ipfs.Cat(cid)
	if err != nil {
		fmt.Println("Error retrieving file content:", err)
		return
	}
	defer r.Close()

	// Get the file information from the stored JSON
	fileInfoList, err := loadFileInfo("files.json")
	if err != nil {
		fmt.Println("Error loading file info:", err)
		return
	}

	// Find the file name by CID
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

	// Create the destination file
	fullPath := filepath.Join(downloadPath, fileName)
	outFile, err := os.Create(fullPath)
	if err != nil {
		fmt.Println("Error creating the local file:", err)
		return
	}
	defer outFile.Close()

	// Copy the content from IPFS to the local file
	_, err = io.Copy(outFile, r)
	if err != nil {
		fmt.Println("Error saving file to local system:", err)
		return
	}

	fmt.Printf("File downloaded successfully to %s\n", fullPath)
}

func addFileInfo(filename string, newFileInfo FileInfo) error {
	// Load existing JSON
	fileInfoList, err := loadFileInfo(filename)
	if err != nil {
		return err
	}

	// Append the new record
	fileInfoList = append(fileInfoList, newFileInfo)

	// Save updated JSON
	return saveFileInfos(filename, fileInfoList)
}

func saveFileInfos(filename string, fileInfoList []FileInfo) error {
	// Encode the FileInfo list to JSON
	data, err := json.MarshalIndent(fileInfoList, "", "  ")
	if err != nil {
		return err
	}

	// Write the encoded data to the file
	return os.WriteFile(filename, data, 0644)
}

func loadFileInfo(filename string) ([]FileInfo, error) {
	// Check if the file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// Return an empty list if the file doesn't exist
		return []FileInfo{}, nil
	}

	// Read the file content
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Decode the JSON into a FileInfo slice
	var fileInfoList []FileInfo
	err = json.Unmarshal(data, &fileInfoList)
	if err != nil {
		return nil, err
	}

	return fileInfoList, nil
}

func ClearScreen() {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux", "darwin": // Linux and MacOS
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

func VerifyIpfsFileIntegrity(ipfs *shell.Shell) {
	var cid string
	var signature string

	fmt.Println("Enter the file CID:")
	fmt.Scanln(&cid)

	fmt.Println("Enter the file signature:")
	fmt.Scanln(&signature)
	// Retrieve the file content from IPFS
	r, err := ipfs.Cat(cid)
	if err != nil {
		fmt.Println("Error retrieving file content:", err)
	}
	defer r.Close()

	// Read the content
	content, err := io.ReadAll(r)
	if err != nil {
		fmt.Println("Error reading file content:", err)
	}

	// Verify the signature
	original := signer.Verify(content, signature)
	if original {
		fmt.Println("File integrity confirmed.")
	} else {
		fmt.Println("File integrity refuted.")
	}
}
