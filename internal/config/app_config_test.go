package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAppConfigStruct(t *testing.T) {
	// Test that the AppConfig struct has the expected fields
	config := AppConfig{
		Port:       "8080",
		DBName:     "test.db",
		DeviceName: "eth0",
	}

	if config.Port != "8080" {
		t.Errorf("expected Port '8080', got '%s'", config.Port)
	}
	if config.DBName != "test.db" {
		t.Errorf("expected DBName 'test.db', got '%s'", config.DBName)
	}
	if config.DeviceName != "eth0" {
		t.Errorf("expected DeviceName 'eth0', got '%s'", config.DeviceName)
	}
}

func TestNewAppConfig_LoadsFromEnvFile(t *testing.T) {
	// Save existing env vars
	originalPort := os.Getenv("PORT")
	originalDBName := os.Getenv("DB_NAME")
	originalDeviceName := os.Getenv("DEVICE_NAME")
	defer func() {
		os.Setenv("PORT", originalPort)
		os.Setenv("DB_NAME", originalDBName)
		os.Setenv("DEVICE_NAME", originalDeviceName)
	}()

	// Clear the env vars so godotenv will load them from file
	os.Unsetenv("PORT")
	os.Unsetenv("DB_NAME")
	os.Unsetenv("DEVICE_NAME")

	// Create a temporary env file
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, "local.env")
	envContent := `PORT=9090
DB_NAME=testdb.sqlite
DEVICE_NAME=wlan0`

	if err := os.WriteFile(envFile, []byte(envContent), 0644); err != nil {
		t.Fatalf("failed to create test env file: %v", err)
	}

	// Change to the temp directory so local.env can be found
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Load the config
	config := NewAppConfig()

	if config.Port != "9090" {
		t.Errorf("expected Port '9090', got '%s'", config.Port)
	}
	if config.DBName != "testdb.sqlite" {
		t.Errorf("expected DBName 'testdb.sqlite', got '%s'", config.DBName)
	}
	if config.DeviceName != "wlan0" {
		t.Errorf("expected DeviceName 'wlan0', got '%s'", config.DeviceName)
	}
}

func TestNewAppConfig_UsesEnvironmentVariables(t *testing.T) {
	// Save existing env vars
	originalPort := os.Getenv("PORT")
	originalDBName := os.Getenv("DB_NAME")
	originalDeviceName := os.Getenv("DEVICE_NAME")
	defer func() {
		os.Setenv("PORT", originalPort)
		os.Setenv("DB_NAME", originalDBName)
		os.Setenv("DEVICE_NAME", originalDeviceName)
	}()

	// Clear the env vars so godotenv will load them from file
	os.Unsetenv("PORT")
	os.Unsetenv("DB_NAME")
	os.Unsetenv("DEVICE_NAME")

	// Create a temporary env file
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, "local.env")
	envContent := `PORT=3000
DB_NAME=env.db
DEVICE_NAME=lo`

	if err := os.WriteFile(envFile, []byte(envContent), 0644); err != nil {
		t.Fatalf("failed to create test env file: %v", err)
	}

	// Change to the temp directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	config := NewAppConfig()

	// Verify the env file was loaded
	if config.Port != "3000" {
		t.Errorf("expected Port '3000', got '%s'", config.Port)
	}
}

func TestNewAppConfig_EmptyValues(t *testing.T) {
	// Create a temporary env file with empty values
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, "local.env")
	envContent := `PORT=
DB_NAME=
DEVICE_NAME=`

	if err := os.WriteFile(envFile, []byte(envContent), 0644); err != nil {
		t.Fatalf("failed to create test env file: %v", err)
	}

	// Save and clear existing env vars
	originalPort := os.Getenv("PORT")
	originalDBName := os.Getenv("DB_NAME")
	originalDeviceName := os.Getenv("DEVICE_NAME")
	defer func() {
		os.Setenv("PORT", originalPort)
		os.Setenv("DB_NAME", originalDBName)
		os.Setenv("DEVICE_NAME", originalDeviceName)
	}()

	os.Unsetenv("PORT")
	os.Unsetenv("DB_NAME")
	os.Unsetenv("DEVICE_NAME")

	// Change to the temp directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	config := NewAppConfig()

	// Empty values are valid
	if config.Port != "" {
		t.Errorf("expected empty Port, got '%s'", config.Port)
	}
	if config.DBName != "" {
		t.Errorf("expected empty DBName, got '%s'", config.DBName)
	}
	if config.DeviceName != "" {
		t.Errorf("expected empty DeviceName, got '%s'", config.DeviceName)
	}
}

func TestAppConfig_AllFieldsAccessible(t *testing.T) {
	config := &AppConfig{
		Port:       "8080",
		DBName:     "packets.db",
		DeviceName: "eth0",
	}

	// Test that all fields are accessible
	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"Port", config.Port, "8080"},
		{"DBName", config.DBName, "packets.db"},
		{"DeviceName", config.DeviceName, "eth0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("expected %s to be '%s', got '%s'", tt.name, tt.expected, tt.got)
			}
		})
	}
}

func TestAppConfig_PointerReceiver(t *testing.T) {
	// Save existing env vars
	originalPort := os.Getenv("PORT")
	originalDBName := os.Getenv("DB_NAME")
	originalDeviceName := os.Getenv("DEVICE_NAME")
	defer func() {
		os.Setenv("PORT", originalPort)
		os.Setenv("DB_NAME", originalDBName)
		os.Setenv("DEVICE_NAME", originalDeviceName)
	}()

	// Clear the env vars so godotenv will load them from file
	os.Unsetenv("PORT")
	os.Unsetenv("DB_NAME")
	os.Unsetenv("DEVICE_NAME")

	// Verify that NewAppConfig returns a pointer
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, "local.env")
	envContent := `PORT=8080
DB_NAME=test.db
DEVICE_NAME=eth0`

	if err := os.WriteFile(envFile, []byte(envContent), 0644); err != nil {
		t.Fatalf("failed to create test env file: %v", err)
	}

	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	config := NewAppConfig()

	if config == nil {
		t.Fatal("NewAppConfig returned nil")
	}

	// Modifying through the pointer should work
	config.Port = "9999"
	if config.Port != "9999" {
		t.Errorf("expected modified Port '9999', got '%s'", config.Port)
	}
}
