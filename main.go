package main

import (
	"fmt"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/mexirica/decentralized-file-signature/config"
	"github.com/mexirica/decentralized-file-signature/ipfs"
)

var downloadPath string

func main() {
	ipfsClient := shell.NewLocalShell()

	err := config.CheckSettingsFileIntegrity()
	if err != nil {
		fmt.Println("Error checking settings file integrity:", err)
		return
	}

	path, err := config.ReadSettingsFile()
	if err != nil {
		return
	}

	if path == "" {
		fmt.Println("Enter the path where the files will be downloaded:")
		fmt.Scanln(&downloadPath)

		// Atualize o caminho no arquivo de configuração
		if err := config.UpdateDownloadPath(downloadPath); err != nil {
			fmt.Println("Error updating settings file:", err)
			return
		}
	} else {
		downloadPath = path
	}

	menu := `
Choose an option:
1. Add file to IPFS
2. List added files
3. Get file information
4. Retrieve file content
5. Download file locally
6. Clear terminal
7. Verify file integrity
8. Exit
`
	for {
		fmt.Println(menu)
		var choice int
		fmt.Scanln(&choice)

		switch choice {
		case 1:
			ipfs.AddFile(ipfsClient)
		case 2:
			ipfs.ListFiles()
		case 3:
			ipfs.GetFileInfo()
		case 4:
			ipfs.RetrieveFileContent(ipfsClient)
		case 5:
			ipfs.DownloadFile(ipfsClient, downloadPath)
		case 6:
			ipfs.ClearScreen()
		case 7:
			ipfs.VerifyIpfsFileIntegrity(ipfsClient)
		case 8:
			fmt.Println("Goodbye!")
			return
		default:
			fmt.Println("Invalid option. Please choose again.")
		}

		fmt.Println()
	}
}
