package ipfs

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
)

func CheckIPFSInstalled() {
	_, err := exec.LookPath("ipfs")
	if err != nil {
		fmt.Println("IPFS not found, attempting to install...")
		installIPFS()
	} else {
		fmt.Println("IPFS is already installed.")
	}
}

func installIPFS() {
	fmt.Println("Installing IPFS...")

	// Adjust download URL and filenames based on OS
	var downloadURL, downloadFile, binaryPath, extractedPath string

	if runtime.GOOS == "windows" {
		downloadURL = "https://dist.ipfs.io/go-ipfs/v0.7.0/go-ipfs_v0.7.0_windows-amd64.zip"
		downloadFile = "go-ipfs.zip"
		binaryPath = "C:/Users/rmecheri/Documents/go-ipfs"
		extractedPath = "go-ipfs"
	} else {
		downloadURL = "https://dist.ipfs.io/go-ipfs/v0.7.0/go-ipfs_v0.7.0_linux-amd64.tar.gz"
		downloadFile = "go-ipfs.tar.gz"
		binaryPath = "/usr/local/bin/go-ipfs"
		extractedPath = "/tmp/go-ipfs"
	}

	err := downloadFileFromURL(downloadURL, downloadFile)
	if err != nil {
		fmt.Printf("Error downloading IPFS: %v\n", err)
		return
	}

	if runtime.GOOS == "windows" {
		// Handle ZIP extraction for Windows
		extractZip(downloadFile)
	} else {
		// Handle tar.gz extraction for Linux/Mac
		extractTarGZ(downloadFile)
	}

	if err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		return
	}

	err = moveBinary(extractedPath, binaryPath)
	if err != nil {
		fmt.Printf("Error moving binary: %v\n", err)
		return
	}

	fmt.Println("IPFS installed!")
	fmt.Scanln()
}

func downloadFileFromURL(url, filename string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = os.WriteFile(filename, body, 0644)
	if err != nil {
		return err
	}

	return nil
}

func extractTarGZ(targzFile string) {
	cmd := exec.Command("tar", "-xzvf", targzFile)
	cmd.Run()
}

func extractZip(zipFile string) {
	cmd := exec.Command("powershell", "Expand-Archive", "-Path", zipFile, "-DestinationPath", ".")
	cmd.Run()
}

func moveBinary(srcPath, dstPath string) error {
	err := os.Rename(srcPath, dstPath)
	if err != nil {
		return fmt.Errorf("Error moving the IPFS binary: %v", err)
	}

	if runtime.GOOS == "windows" {
		cmd := exec.Command("attrib", "+r", dstPath) // Make sure the file is readable
		cmd.Run()
	} else {
		cmd := exec.Command("chmod", "+x", dstPath)
		cmd.Run()
	}

	return nil
}
