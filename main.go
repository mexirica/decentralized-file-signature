package main

import (
	"fmt"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/mexirica/decentralized-file-signature/ipfs"
)

var downloadPath string

func main() {
	//ipfs.CheckIPFSInstalled()
	ipfsClient := shell.NewLocalShell()

	path, _ := ipfs.CheckSettingsFile()

	if path == "" {
		fmt.Println("Enter the path where the files will be downloaded:")
		fmt.Scanln(&downloadPath)

		// Atualize o caminho no arquivo de configuração
		if err := ipfs.UpdateDownloadPath(downloadPath); err != nil {
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
6. Exit
`
	for {
		fmt.Print(menu)
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
			fmt.Println("Goodbye!")
			return
		default:
			fmt.Println("Invalid option. Please choose again.")
		}

		fmt.Println()
	}
}
