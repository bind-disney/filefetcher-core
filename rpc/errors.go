package rpc

import "fmt"

func clientAlreadyExistsError(address string) error {
	return fmt.Errorf("client with address %s already exists", address)
}

func clientDoesNotExistsError(address string) error {
	return fmt.Errorf("client with address %s is not exists", address)
}

func invalidFileSystemEntryError(fileSystemEntry string) error {
	return fmt.Errorf("filesystem entry '%s' is invalid or doesn't exists", fileSystemEntry)
}
