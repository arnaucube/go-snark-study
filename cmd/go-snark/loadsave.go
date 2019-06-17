package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
)

func loadFromReader(r io.Reader, obj interface{}) error {
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(r); err != nil {
		return err
	}
	return json.Unmarshal(buf.Bytes(), obj)
}

func loadFromFile(path string, obj interface{}) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return loadFromReader(f, obj)
}

func saveToWriter(w io.Writer, obj interface{}) error {
	b, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return err
	}
	if _, err := w.Write(b); err != nil {
		return err
	}
	return nil
}

func saveToFile(path string, obj interface{}) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return saveToWriter(f, obj)
}
