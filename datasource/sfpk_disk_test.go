package datasource

import (
	"github.com/xen-echo/go-repository/service"
	"os"
	"strings"
	"testing"
	"time"
)

type data struct {
	Test string
}

func setupWithEncryption(t *testing.T, enc service.EncryptionService) SFPKDiskDS[data] {
	if err := os.Setenv("SFPK_ROOT", "../sfpk"); err != nil {
		t.Errorf("Error setting SFPK_ROOT: %v", err)
	}

	var ds SFPKDiskDS[data]

	if enc == nil {
		ds = NewSFPKDiskDS[data]("test-ds")
	} else {
		ds = NewSFPKDiskDSWithEncryption[data]("test-ds", enc)
	}

	if err := ds.Wipe(); err != nil {
		t.Errorf("Error wiping disk: %v", err)
	}

	return ds
}

func setup(t *testing.T) SFPKDiskDS[data] {
	return setupWithEncryption(t, nil)
}

func TestWritingToDisk(t *testing.T) {

	ds := setup(t)
	defer ds.Wipe()
	defer os.Unsetenv("SFPK_ROOT")

	dataFile, err := ds.GetDataFile("datafile")

	if err != nil {
		t.Errorf("Error getting data file: %v", err)
	}

	if dataFile == nil {
		t.Errorf("Data file is nil")
	}

	if dataFile.Item.Key != "datafile" {
		t.Errorf("Data file key is not correct")
	}

	if dataFile.Item.Value != nil {
		t.Errorf("Data file value is not nil")
	}

	if dataFile.Item.TTLSeconds != 0 {
		t.Errorf("Data file TTLSeconds is not 0")
	}

	if dataFile.Item.ModifiedAtSeconds != 0 {
		t.Errorf("Data file ModifiedAtSeconds is not 0")
	}

	test := data{"test"}

	dataFile.Item.Value = &test

	if err := ds.SaveDataFile(dataFile); err != nil {
		t.Errorf("Error saving data file: %v", err)
	}

	dataFile.Unlock()

	dataFile, err = ds.GetDataFile("datafile")

	if err != nil {
		t.Errorf("Error getting data file: %v", err)
	}

	if dataFile.Item.Key != "datafile" {
		t.Errorf("Data file key is not correct")
	}

	if dataFile.Item.Value == nil {
		t.Errorf("Data file value is nil")
	}

	if dataFile.Item.Value.Test != "test" {
		t.Errorf("Data file value is not correct")
	}

	if dataFile.Item.TTLSeconds != 0 {
		t.Errorf("Data file TTLSeconds is not 0")
	}

	if dataFile.Item.ModifiedAtSeconds == 0 {
		t.Errorf("Data file ModifiedAtSeconds is 0")
	}

}

func TestCheckTTL(t *testing.T) {

	ds := setup(t)
	defer ds.Wipe()
	defer os.Unsetenv("SFPK_ROOT")

	dataFile, _ := ds.GetDataFile("datafile")
	dataFile.Item.Value = &data{"test"}
	dataFile.Item.TTLSeconds = 1
	ds.SaveDataFile(dataFile)
	dataFile.Unlock()

	time.Sleep(2 * time.Second)

	dataFile, _ = ds.GetDataFile("datafile")
	if dataFile.Item.Value != nil {
		t.Errorf("Data file should have been removed")
	}
	defer dataFile.Unlock()

}

func TestLocking(t *testing.T) {

	ds := setup(t)
	defer ds.Wipe()
	defer os.Unsetenv("SFPK_ROOT")

	dataFile, err := ds.GetDataFile("datafile")
	if err != nil {
		t.Errorf("Error getting data file: %v", err)
	}

	done := make(chan bool)
	go func() {
		_, err := ds.GetDataFile("datafile")
		if err != nil {
			t.Errorf("Error getting data file in goroutine: %v", err)
		}
		done <- true
	}()

	// Wait for goroutine to get data file
	time.Sleep(1 * time.Second)

	// Unlock data file
	dataFile.Unlock()

	// Wait for goroutine to finish
	select {
	case <-done:
		break
	case <-time.After(2 * time.Second):
		t.Errorf("Goroutine did not finish in time")
	}

}

func TestWithEncryption(t *testing.T) {

	ds := setupWithEncryption(t, service.NewEncryptionService("password"))
	defer ds.Wipe()
	defer os.Unsetenv("SFPK_ROOT")

	dataFile, err := ds.GetDataFile("datafile")
	if err != nil {
		t.Errorf("Error getting data file: %v", err)
	}

	dataFile.Item.Value = &data{"test"}

	if err := ds.SaveDataFile(dataFile); err != nil {
		t.Errorf("Error saving data file: %v", err)
	}

	dataFile.Unlock()

	file, err := os.ReadFile(dataFile.FilePath)

	if err != nil {
		t.Errorf("Error reading file: %v", err)
	}

	if string(file) == "" {
		t.Errorf("File is empty")
	}

	// Check that the file is encrypted
	if strings.Contains(string(file), "\"test\"") {
		t.Errorf("File is not encrypted")
	}

	dataFile, err = ds.GetDataFile("datafile")
	if err != nil {
		t.Errorf("Error getting data file: %v", err)
	}

	if dataFile.Item.Value == nil {
		t.Errorf("Data file value is nil")
	}

	if dataFile.Item.Value.Test != "test" {
		t.Errorf("Data file value is not correct")
	}

}
