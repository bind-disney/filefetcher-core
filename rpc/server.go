package rpc

import (
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/bind-disney/filefetcher-core/cli"
)

type (
	Server struct {
		Clients   *clientManager
		logger    *log.Logger
		directory string
	}

	ClientRequest struct {
		ClientAddress string
	}

	FileSystemRequest struct {
		ClientRequest
		FileSystemEntry string
	}

	DownloadResponse struct {
		FileSize int64
		Address  string
	}
)

const bufferSize int64 = 4096

func NewServer(directory string, logger *log.Logger) (*Server, error) {
	fullPath, err := filepath.Abs(directory)
	if err != nil {
		return nil, err
	}

	server := &Server{
		directory: fullPath,
		logger:    logger,
	}
	server.Clients = newClientManager(server)

	return server, nil
}

func NewClientRequest(address string) ClientRequest {
	return ClientRequest{ClientAddress: address}
}

func NewFileSystemRequest(address, fileSystemEntry string) FileSystemRequest {
	return FileSystemRequest{ClientRequest: NewClientRequest(address), FileSystemEntry: fileSystemEntry}
}

func (server *Server) CurrentDirectory(request ClientRequest, directory *string) error {
	client, exists := server.Clients.get(request.ClientAddress)
	if !exists {
		return clientDoesNotExistsError(request.ClientAddress)
	}

	path, err := filepath.Rel(server.directory, client.directory)
	if err != nil {
		return err
	}

	*directory = strings.Replace(path, ".", "/", 1)

	return nil
}

func (server *Server) ChangeDirectory(request FileSystemRequest, newDirectory *string) error {
	connectedClient, exists := server.Clients.get(request.ClientAddress)
	if !exists {
		return clientDoesNotExistsError(request.ClientAddress)
	}

	directory, err := server.cleanFileSystemEntry(connectedClient, request.FileSystemEntry)
	if err != nil {
		return err
	}

	err = os.Chdir(directory)
	if err != nil {
		return invalidFileSystemEntryError(directory)
	}

	relativeDirectory, err := filepath.Rel(server.directory, directory)
	if err != nil {
		return err
	}

	*newDirectory = string(filepath.Separator) + strings.Replace(relativeDirectory, ".", "", 1)
	connectedClient.directory = directory

	return nil
}

func (server *Server) ListFiles(request FileSystemRequest, files *[]string) error {
	connectedClient, exists := server.Clients.get(request.ClientAddress)
	if !exists {
		return clientDoesNotExistsError(request.ClientAddress)
	}

	directory, err := server.cleanFileSystemEntry(connectedClient, request.FileSystemEntry)
	if err != nil {
		return err
	}

	entries, err := ioutil.ReadDir(directory)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		entryName := entry.Name()
		if entry.IsDir() {
			entryName += string(filepath.Separator)
		}

		*files = append(*files, entryName)
	}

	return nil
}

func (server *Server) GetFile(request FileSystemRequest, response *DownloadResponse) error {
	connectedClient, exists := server.Clients.get(request.ClientAddress)
	if !exists {
		return clientDoesNotExistsError(request.ClientAddress)
	}

	fileSystemEntry, err := server.cleanFileSystemEntry(connectedClient, request.FileSystemEntry)
	if err != nil {
		return err
	}

	file, err := os.Open(fileSystemEntry)
	if err != nil {
		return invalidFileSystemEntryError(fileSystemEntry)
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	socket, err := net.Listen("tcp", ":0") // Choose random free port if available
	if err != nil {
		return err
	}

	response.FileSize = fileInfo.Size()
	response.Address = socket.Addr().String()

	go func() {
		connection, err := socket.Accept()
		if err != nil {
			cli.LogError("Download file", err)
			return
		}
		defer connection.Close()
		defer file.Close()

		log.Printf("Opened connection for file '%s' on %v\n", request.FileSystemEntry, connection.LocalAddr())

		sendBuffer := make([]byte, bufferSize)

		for {
			if _, err = file.Read(sendBuffer); err == io.EOF {
				break
			}

			if _, err = connection.Write(sendBuffer); err != nil {
				cli.LogError("Download file", err)
				return
			}
		}

		log.Printf("Closed connection for file '%s' on %v, download successful\n", request.FileSystemEntry, connection.LocalAddr())
	}()

	return nil
}

func (server *Server) BufferSize(request ClientRequest, size *int64) error {
	*size = bufferSize
	return nil
}

func (server *Server) cleanFileSystemEntry(connectedClient *client, fileSystemEntry string) (string, error) {
	fileSystemEntry, err := filepath.Abs(filepath.Join(connectedClient.directory, fileSystemEntry))

	if err != nil || !strings.HasPrefix(fileSystemEntry, server.directory) {
		return "", invalidFileSystemEntryError(fileSystemEntry)
	}

	return fileSystemEntry, nil
}
