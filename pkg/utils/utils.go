package utils

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

// FileExists checks if file exists
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// DirExists checks if directory exists
func DirExists(dirname string) bool {
	_, err := os.Stat(dirname)
	return err == nil
}

// ReadJSON reads json file
func ReadJSON(file string) ([]byte, error) {
	fd, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	buf, err := ioutil.ReadAll(fd)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

// WriteJSON writes content into json file
func WriteJSON(file string, data interface{}) error {
	content, err := json.Marshal(data)
	if err != nil {
		return err
	}
	fd, err := os.Create(file)
	if err != nil {
		return err
	}
	defer fd.Close()
	if _, err = fd.Write(content); err != nil {
		return err
	}
	return nil
}
