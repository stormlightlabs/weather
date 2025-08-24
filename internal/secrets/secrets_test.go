package secrets

import (
	"os"
	"strings"
	"testing"
)

func TestKeyValidator_ValidateKey(t *testing.T) {
	validator := NewKeyValidator()

	tests := []struct {
		name        string
		key         string
		expectError bool
		errorMsg    string
	}{
		{name: "valid key", key: "MySecureKey123", expectError: false},
		{
			name:        "too short",
			key:         "short",
			expectError: true,
			errorMsg:    "key must be at least 12 characters long",
		},
		{
			name:        "missing uppercase",
			key:         "lowercase123",
			expectError: true,
			errorMsg:    "key must contain at least one uppercase letter",
		},
		{
			name:        "missing lowercase",
			key:         "UPPERCASE123",
			expectError: true,
			errorMsg:    "key must contain at least one lowercase letter",
		},
		{
			name:        "missing digits",
			key:         "NoDigitsHere",
			expectError: true,
			errorMsg:    "key must contain at least one digit",
		},
		{
			name:        "blacklisted word",
			key:         "password123Test",
			expectError: true,
			errorMsg:    "key contains forbidden pattern: password",
		},
		{name: "weak entropy", key: "Aaa1aaa1aaa1", expectError: false},
		{name: "valid with symbols", key: "MySecure!Key123", expectError: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validator.ValidateKey(test.key)
			if test.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if !strings.Contains(err.Error(), test.errorMsg) {
					t.Errorf("expected error containing '%s', got '%s'", test.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestKeyValidator_CustomValidation(t *testing.T) {
	// Test custom validator settings
	validator := &KeyValidator{
		MinLength:      8,
		RequireUpper:   false,
		RequireLower:   true,
		RequireDigits:  false,
		RequireSymbols: true,
		Blacklist:      []string{"test"},
	}

	tests := []struct {
		name        string
		key         string
		expectError bool
		errorMsg    string
	}{
		{name: "valid custom key", key: "lowercase!", expectError: false},
		{
			name:        "missing symbol (required)",
			key:         "lowercaseonly",
			expectError: true,
			errorMsg:    "key must contain at least one symbol",
		},
		{
			name:        "custom blacklist",
			key:         "testkey!",
			expectError: true,
			errorMsg:    "key contains forbidden pattern: test",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validator.ValidateKey(test.key)
			if test.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if !strings.Contains(err.Error(), test.errorMsg) {
					t.Errorf("expected error containing '%s', got '%s'", test.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestGetEncryptionKey(t *testing.T) {
	originalEnvKey := os.Getenv("WEATHER_API_ENCRYPTION_KEY")
	defer func() {
		if originalEnvKey == "" {
			os.Unsetenv("WEATHER_API_ENCRYPTION_KEY")
		} else {
			os.Setenv("WEATHER_API_ENCRYPTION_KEY", originalEnvKey)
		}
	}()

	tests := []struct {
		name        string
		cliKey      string
		envKey      string
		expectError bool
		expectedKey string
	}{
		{
			name:        "valid CLI key takes precedence",
			cliKey:      "CliKey123Valid",
			envKey:      "EnvKey123Valid",
			expectError: false,
			expectedKey: "CliKey123Valid",
		},
		{
			name:        "valid env key used when no CLI key",
			cliKey:      "",
			envKey:      "EnvKey123Valid",
			expectError: false,
			expectedKey: "EnvKey123Valid",
		},
		{
			name:        "invalid CLI key",
			cliKey:      "short",
			envKey:      "",
			expectError: true,
		},
		{
			name:        "invalid env key",
			cliKey:      "",
			envKey:      "invalid",
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Set env key
			if test.envKey == "" {
				os.Unsetenv("WEATHER_API_ENCRYPTION_KEY")
			} else {
				os.Setenv("WEATHER_API_ENCRYPTION_KEY", test.envKey)
			}

			// Skip prompt tests for now as they require user interaction
			if test.cliKey == "" && test.envKey == "" {
				t.Skip("Skipping prompt test - requires user interaction")
			}

			key, err := GetEncryptionKey(test.cliKey)
			if test.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
				if key != test.expectedKey {
					t.Errorf("expected key '%s', got '%s'", test.expectedKey, key)
				}
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	// Save original env vars
	originalDBURL := os.Getenv("DATABASE_URL")
	originalNWSAgent := os.Getenv("NWS_AGENT")
	defer func() {
		if originalDBURL == "" {
			os.Unsetenv("DATABASE_URL")
		} else {
			os.Setenv("DATABASE_URL", originalDBURL)
		}
		if originalNWSAgent == "" {
			os.Unsetenv("NWS_AGENT")
		} else {
			os.Setenv("NWS_AGENT", originalNWSAgent)
		}
	}()

	tests := []struct {
		name             string
		dbURL            string
		nwsAgent         string
		expectedDBURL    string
		expectedNWSAgent string
	}{
		{
			name:             "both env vars set",
			dbURL:            "postgres://user:pass@localhost/db",
			nwsAgent:         "custom-agent/1.0",
			expectedDBURL:    "postgres://user:pass@localhost/db",
			expectedNWSAgent: "custom-agent/1.0",
		},
		{
			name:             "default NWS agent",
			dbURL:            "postgres://user:pass@localhost/db",
			nwsAgent:         "",
			expectedDBURL:    "postgres://user:pass@localhost/db",
			expectedNWSAgent: "weather-api/1.0 (https://github.com/stormlight-labs/weather-api)",
		},
		{
			name:             "empty values",
			dbURL:            "",
			nwsAgent:         "",
			expectedDBURL:    "",
			expectedNWSAgent: "weather-api/1.0 (https://github.com/stormlight-labs/weather-api)",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Set env vars
			if test.dbURL == "" {
				os.Unsetenv("DATABASE_URL")
			} else {
				os.Setenv("DATABASE_URL", test.dbURL)
			}
			if test.nwsAgent == "" {
				os.Unsetenv("NWS_AGENT")
			} else {
				os.Setenv("NWS_AGENT", test.nwsAgent)
			}

			config, err := LoadConfig()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if config.DatabaseURL != test.expectedDBURL {
				t.Errorf("expected DatabaseURL '%s', got '%s'", test.expectedDBURL, config.DatabaseURL)
			}
			if config.NWSAgent != test.expectedNWSAgent {
				t.Errorf("expected NWSAgent '%s', got '%s'", test.expectedNWSAgent, config.NWSAgent)
			}
		})
	}
}

func TestConfig_ValidateConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: Config{
				DatabaseURL: "postgres://user:pass@localhost/db",
				NWSAgent:    "weather-api/1.0",
			},
			expectError: false,
		},
		{
			name: "missing DATABASE_URL",
			config: Config{
				DatabaseURL: "",
				NWSAgent:    "weather-api/1.0",
			},
			expectError: true,
			errorMsg:    "DATABASE_URL is required",
		},
		{
			name: "invalid DATABASE_URL format",
			config: Config{
				DatabaseURL: "mysql://user:pass@localhost/db",
				NWSAgent:    "weather-api/1.0",
			},
			expectError: true,
			errorMsg:    "DATABASE_URL must be a valid PostgreSQL connection string",
		},
		{
			name: "missing NWS_AGENT",
			config: Config{
				DatabaseURL: "postgres://user:pass@localhost/db",
				NWSAgent:    "",
			},
			expectError: true,
			errorMsg:    "NWS_AGENT is required",
		},
		{
			name: "postgresql:// prefix valid",
			config: Config{
				DatabaseURL: "postgresql://user:pass@localhost/db",
				NWSAgent:    "weather-api/1.0",
			},
			expectError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.config.ValidateConfig()
			if test.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if !strings.Contains(err.Error(), test.errorMsg) {
					t.Errorf("expected error containing '%s', got '%s'", test.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestEncryptDecryptValue(t *testing.T) {
	key := "TestKey123Valid"
	originalValue := "sensitive-database-url"

	encryptedValue, err := EncryptValue(originalValue, key)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	if encryptedValue == originalValue {
		t.Error("encrypted value should be different from original")
	}

	parts := strings.Split(encryptedValue, ":")
	if len(parts) != 3 {
		t.Errorf("expected encrypted value to have 3 parts, got %d", len(parts))
	}

	decryptedValue, err := DecryptValue(encryptedValue, key)
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}

	if decryptedValue != originalValue {
		t.Errorf("expected decrypted value '%s', got '%s'", originalValue, decryptedValue)
	}

	wrongKey := "WrongKey123Valid"
	_, err = DecryptValue(encryptedValue, wrongKey)
	if err == nil {
		t.Error("expected decryption to fail with wrong key")
	}

	plainValue := "not-encrypted"
	result, err := DecryptValue(plainValue, key)
	if err != nil {
		t.Errorf("unexpected error for non-encrypted value: %v", err)
	}
	if result != plainValue {
		t.Errorf("expected non-encrypted value to be returned as-is, got '%s'", result)
	}
}

func TestIsEncrypted(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{
			name:     "encrypted format",
			value:    "abcdef123456:deadbeef:cafebabe",
			expected: true,
		},
		{
			name:     "plain text",
			value:    "plain-text-value",
			expected: false,
		},
		{
			name:     "wrong number of parts",
			value:    "abc:def",
			expected: false,
		},
		{
			name:     "non-hex parts",
			value:    "abc:def:ghi",
			expected: false,
		},
		{
			name:     "empty string",
			value:    "",
			expected: false,
		},
		{
			name:     "valid hex format",
			value:    "deadbeef:cafebabe:feedface",
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := IsEncrypted(test.value)
			if result != test.expected {
				t.Errorf("expected %v, got %v for value '%s'", test.expected, result, test.value)
			}
		})
	}
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key := "RoundTripKey123"
	values := []string{
		"postgres://user:pass@localhost:5432/weather",
		"weather-api/1.0 (mailto:admin@example.com)",
		"simple-value",
		"value with spaces and symbols!@#$",
		"", // empty string
	}

	for _, originalValue := range values {
		t.Run("value_"+originalValue, func(t *testing.T) {
			encrypted, err := EncryptValue(originalValue, key)
			if err != nil {
				t.Fatalf("encryption failed: %v", err)
			}

			if !IsEncrypted(encrypted) {
				t.Error("encrypted value doesn't look encrypted")
			}

			decrypted, err := DecryptValue(encrypted, key)
			if err != nil {
				t.Fatalf("decryption failed: %v", err)
			}

			if decrypted != originalValue {
				t.Errorf("round trip failed: expected '%s', got '%s'", originalValue, decrypted)
			}
		})
	}
}

func TestEncryptionUniqueness(t *testing.T) {
	key := "UniquenessTestKey123"
	value := "same-value"

	encrypted1, err := EncryptValue(value, key)
	if err != nil {
		t.Fatalf("first encryption failed: %v", err)
	}

	encrypted2, err := EncryptValue(value, key)
	if err != nil {
		t.Fatalf("second encryption failed: %v", err)
	}

	if encrypted1 == encrypted2 {
		t.Error("multiple encryptions of same value should produce different results")
	}

	decrypted1, err := DecryptValue(encrypted1, key)
	if err != nil {
		t.Fatalf("first decryption failed: %v", err)
	}

	decrypted2, err := DecryptValue(encrypted2, key)
	if err != nil {
		t.Fatalf("second decryption failed: %v", err)
	}

	if decrypted1 != value || decrypted2 != value {
		t.Error("decrypted values don't match original")
	}
}

func TestGenerateSecureKey(t *testing.T) {
	tests := []struct {
		name      string
		length    int
		expectMin int // minimum expected length
	}{
		{"default length", 16, 16},
		{"minimum length", 12, 12},
		{"short input gets extended", 8, 16}, // Should default to 16
		{"long length", 32, 32},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			key, err := GenerateSecureKey(test.length)
			if err != nil {
				t.Fatalf("GenerateSecureKey failed: %v", err)
			}

			// Check length
			if len(key) < test.expectMin {
				t.Errorf("expected key length >= %d, got %d", test.expectMin, len(key))
			}

			// Validate the key passes our validation
			validator := NewKeyValidator()
			if err := validator.ValidateKey(key); err != nil {
				t.Errorf("generated key failed validation: %v", err)
			}

			// Check character diversity
			hasUpper := false
			hasLower := false
			hasDigit := false
			for _, r := range key {
				if r >= 'A' && r <= 'Z' {
					hasUpper = true
				}
				if r >= 'a' && r <= 'z' {
					hasLower = true
				}
				if r >= '0' && r <= '9' {
					hasDigit = true
				}
			}

			if !hasUpper {
				t.Error("generated key missing uppercase letters")
			}
			if !hasLower {
				t.Error("generated key missing lowercase letters")
			}
			if !hasDigit {
				t.Error("generated key missing digits")
			}
		})
	}
}

func TestGenerateSecureKeyUniqueness(t *testing.T) {
	// Generate multiple keys and ensure they're unique
	keys := make(map[string]bool)
	for i := 0; i < 100; i++ {
		key, err := GenerateSecureKey(16)
		if err != nil {
			t.Fatalf("GenerateSecureKey failed on iteration %d: %v", i, err)
		}

		if keys[key] {
			t.Errorf("duplicate key generated: %s", key)
		}
		keys[key] = true
	}
}

func TestWriteKeyToFile(t *testing.T) {
	tempDir := t.TempDir()
	keyFile := tempDir + "/test.key"
	testKey := "TestKeyForFile123"

	// Test writing key to file
	err := WriteKeyToFile(testKey, keyFile)
	if err != nil {
		t.Fatalf("WriteKeyToFile failed: %v", err)
	}

	// Check file exists and has correct content
	content, err := os.ReadFile(keyFile)
	if err != nil {
		t.Fatalf("failed to read key file: %v", err)
	}

	if string(content) != testKey {
		t.Errorf("expected file content '%s', got '%s'", testKey, string(content))
	}

	// Check file permissions
	info, err := os.Stat(keyFile)
	if err != nil {
		t.Fatalf("failed to stat key file: %v", err)
	}

	// Check that permissions are restrictive (0600)
	mode := info.Mode()
	if mode&0777 != 0600 {
		t.Errorf("expected file permissions 0600, got %o", mode&0777)
	}
}

func TestWriteKeyToFileOverwrite(t *testing.T) {
	tempDir := t.TempDir()
	keyFile := tempDir + "/overwrite.key"

	// Write first key
	firstKey := "FirstKey123Test"
	err := WriteKeyToFile(firstKey, keyFile)
	if err != nil {
		t.Fatalf("first WriteKeyToFile failed: %v", err)
	}

	// Write second key (should overwrite)
	secondKey := "SecondKey456Test"
	err = WriteKeyToFile(secondKey, keyFile)
	if err != nil {
		t.Fatalf("second WriteKeyToFile failed: %v", err)
	}

	// Check file has second key content
	content, err := os.ReadFile(keyFile)
	if err != nil {
		t.Fatalf("failed to read key file: %v", err)
	}

	if string(content) != secondKey {
		t.Errorf("expected file content '%s', got '%s'", secondKey, string(content))
	}
}

func TestEnsureGitIgnore(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// Change to temp directory
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("failed to change to temp directory: %v", err)
	}

	filename := "test.key"

	// Test adding to non-existent .gitignore
	err = EnsureGitIgnore(filename)
	if err != nil {
		t.Fatalf("ensureGitIgnore failed: %v", err)
	}

	// Check .gitignore was created and contains the file
	content, err := os.ReadFile(".gitignore")
	if err != nil {
		t.Fatalf("failed to read .gitignore: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, filename) {
		t.Errorf(".gitignore does not contain '%s'. Content: %s", filename, contentStr)
	}

	// Test adding same file again (should not duplicate)
	err = EnsureGitIgnore(filename)
	if err != nil {
		t.Fatalf("second ensureGitIgnore failed: %v", err)
	}

	// Check file is still only mentioned once
	content2, err := os.ReadFile(".gitignore")
	if err != nil {
		t.Fatalf("failed to read .gitignore second time: %v", err)
	}

	lines := strings.Split(string(content2), "\n")
	count := 0
	for _, line := range lines {
		if strings.TrimSpace(line) == filename {
			count++
		}
	}

	if count != 1 {
		t.Errorf("expected filename to appear once in .gitignore, found %d times", count)
	}
}

func TestEnsureGitIgnoreExisting(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// Change to temp directory
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("failed to change to temp directory: %v", err)
	}

	// Create existing .gitignore with some content
	existingContent := "node_modules/\n*.log\n"
	err = os.WriteFile(".gitignore", []byte(existingContent), 0644)
	if err != nil {
		t.Fatalf("failed to create existing .gitignore: %v", err)
	}

	filename := "secret.key"
	err = EnsureGitIgnore(filename)
	if err != nil {
		t.Fatalf("ensureGitIgnore failed: %v", err)
	}

	// Check .gitignore contains both old and new content
	content, err := os.ReadFile(".gitignore")
	if err != nil {
		t.Fatalf("failed to read .gitignore: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "node_modules/") {
		t.Error(".gitignore missing original content")
	}
	if !strings.Contains(contentStr, filename) {
		t.Errorf(".gitignore missing new filename '%s'", filename)
	}
}

func TestMustRandomInt(t *testing.T) {
	// Test valid cases
	for n := 1; n <= 100; n++ {
		result := MustRandomInt(n)
		if result < 0 || result >= n {
			t.Errorf("mustRandomInt(%d) = %d, expected in range [0, %d)", n, result, n)
		}
	}

	// Test that it produces different values (randomness test)
	results := make(map[int]bool)
	for i := 0; i < 50; i++ {
		result := MustRandomInt(10)
		results[result] = true
	}

	// We should get at least some variety (not just one value)
	if len(results) < 2 {
		t.Error("mustRandomInt appears to be producing non-random values")
	}
}

func TestMustRandomIntPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("mustRandomInt(0) should panic but didn't")
		}
	}()
	MustRandomInt(0)
}

func TestShuffleString(t *testing.T) {
	original := "abcdefghijklmnop"
	shuffled := ShuffleString(original)

	// Should have same length
	if len(shuffled) != len(original) {
		t.Errorf("shuffled string length %d, expected %d", len(shuffled), len(original))
	}

	// Should contain same characters (different order)
	originalRunes := []rune(original)
	shuffledRunes := []rune(shuffled)

	// Count character frequencies
	originalFreq := make(map[rune]int)
	shuffledFreq := make(map[rune]int)

	for _, r := range originalRunes {
		originalFreq[r]++
	}
	for _, r := range shuffledRunes {
		shuffledFreq[r]++
	}

	// Frequencies should match
	if len(originalFreq) != len(shuffledFreq) {
		t.Error("shuffled string has different character set")
	}

	for r, count := range originalFreq {
		if shuffledFreq[r] != count {
			t.Errorf("character '%c' frequency mismatch: expected %d, got %d", r, count, shuffledFreq[r])
		}
	}

	// Test multiple shuffles to ensure randomness
	same := 0
	for i := 0; i < 20; i++ {
		if ShuffleString(original) == original {
			same++
		}
	}

	// It's extremely unlikely (but not impossible) for all shuffles to be identical
	if same == 20 {
		t.Error("shuffleString appears to be non-random (all results identical to input)")
	}
}
