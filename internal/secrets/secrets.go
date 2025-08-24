package secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"regexp"
	"strings"
	"syscall"

	"golang.org/x/crypto/scrypt"
	"golang.org/x/term"
)

// Config holds all application secrets
type Config struct {
	DatabaseURL string
	NWSAgent    string
}

// KeyValidator validates encryption keys
type KeyValidator struct {
	MinLength      int
	RequireUpper   bool
	RequireLower   bool
	RequireDigits  bool
	RequireSymbols bool
	Blacklist      []string
}

// NewKeyValidator creates a default key validator
func NewKeyValidator() *KeyValidator {
	return &KeyValidator{
		MinLength:      12,
		RequireUpper:   true,
		RequireLower:   true,
		RequireDigits:  true,
		RequireSymbols: false,
		Blacklist: []string{
			"password", "password123", "secret", "secret123",
			"admin", "admin123", "weather", "weather123",
			"123456789012", "abcdef123456",
		},
	}
}

// ValidateKey validates an encryption key against security requirements
func (kv *KeyValidator) ValidateKey(key string) error {
	if len(key) < kv.MinLength {
		return fmt.Errorf("key must be at least %d characters long", kv.MinLength)
	}

	if kv.RequireUpper && !regexp.MustCompile(`[A-Z]`).MatchString(key) {
		return fmt.Errorf("key must contain at least one uppercase letter")
	}
	if kv.RequireLower && !regexp.MustCompile(`[a-z]`).MatchString(key) {
		return fmt.Errorf("key must contain at least one lowercase letter")
	}
	if kv.RequireDigits && !regexp.MustCompile(`[0-9]`).MatchString(key) {
		return fmt.Errorf("key must contain at least one digit")
	}
	if kv.RequireSymbols && !regexp.MustCompile(`[^A-Za-z0-9]`).MatchString(key) {
		return fmt.Errorf("key must contain at least one symbol")
	}

	keyLower := strings.ToLower(key)
	for _, forbidden := range kv.Blacklist {
		if strings.Contains(keyLower, forbidden) {
			return fmt.Errorf("key contains forbidden pattern: %s", forbidden)
		}
	}

	if len(key) > 0 && len(strings.TrimLeft(key, string(key[0]))) == 0 {
		return fmt.Errorf("key must not be all the same character")
	}

	return nil
}

// GetEncryptionKey retrieves the encryption key from various sources with validation
//
//	Priority order: CLI arg -> ENV var -> prompt
func GetEncryptionKey(cliKey string) (string, error) {
	validator := NewKeyValidator()
	var key string

	if cliKey != "" {
		key = cliKey
	} else if envKey := os.Getenv("WEATHER_API_ENCRYPTION_KEY"); envKey != "" {
		key = envKey
	} else {
		var err error
		key, err = promptForKey("Enter encryption key: ")
		if err != nil {
			return "", fmt.Errorf("failed to read key: %w", err)
		}
	}

	if err := validator.ValidateKey(key); err != nil {
		return "", fmt.Errorf("key validation failed: %w", err)
	}

	return key, nil
}

// LoadConfig loads the application configuration from environment or encrypted file
func LoadConfig() (*Config, error) {
	config := &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		NWSAgent:    os.Getenv("NWS_AGENT"),
	}

	if config.NWSAgent == "" {
		config.NWSAgent = "weather-api/1.0 (https://github.com/stormlight-labs/weather-api)"
	}

	return config, nil
}

// ValidateConfig validates the loaded configuration
func (c *Config) ValidateConfig() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}

	if !strings.HasPrefix(c.DatabaseURL, "postgres://") && !strings.HasPrefix(c.DatabaseURL, "postgresql://") {
		return fmt.Errorf("DATABASE_URL must be a valid PostgreSQL connection string")
	}

	if c.NWSAgent == "" {
		return fmt.Errorf("NWS_AGENT is required")
	}

	return nil
}

// EncryptValue encrypts a single value using the provided key
func EncryptValue(value, key string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	derivedKey, err := scrypt.Key([]byte(key), salt, 32768, 8, 1, 32)
	if err != nil {
		return "", fmt.Errorf("key derivation failed: %w", err)
	}

	block, err := aes.NewCipher(derivedKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := aesGCM.Seal(nil, nonce, []byte(value), nil)

	// Format: salt:nonce:ciphertext (all hex encoded)
	return fmt.Sprintf("%s:%s:%s",
		hex.EncodeToString(salt),
		hex.EncodeToString(nonce),
		hex.EncodeToString(ciphertext)), nil
}

// DecryptValue decrypts a single value using the provided key
func DecryptValue(encryptedValue, key string) (string, error) {
	parts := strings.Split(encryptedValue, ":")
	if len(parts) != 3 {
		return encryptedValue, nil
	}

	salt, err := hex.DecodeString(parts[0])
	if err != nil {
		return encryptedValue, nil
	}

	nonce, err := hex.DecodeString(parts[1])
	if err != nil {
		return encryptedValue, nil
	}

	ciphertext, err := hex.DecodeString(parts[2])
	if err != nil {
		return encryptedValue, nil
	}

	derivedKey, err := scrypt.Key([]byte(key), salt, 32768, 8, 1, 32)
	if err != nil {
		return "", fmt.Errorf("key derivation failed: %w", err)
	}

	block, err := aes.NewCipher(derivedKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decryption failed: %w", err)
	}

	return string(plaintext), nil
}

// IsEncrypted checks if a value appears to be encrypted
func IsEncrypted(value string) bool {
	parts := strings.Split(value, ":")
	if len(parts) != 3 {
		return false
	}

	for _, part := range parts {
		if _, err := hex.DecodeString(part); err != nil {
			return false
		}
	}

	return true
}

// GenerateSecureKey generates a cryptographically secure key that passes validation
func GenerateSecureKey(length int) (string, error) {
	if length < 12 {
		length = 16
	}

	uppercase := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	lowercase := "abcdefghijklmnopqrstuvwxyz"
	digits := "0123456789"
	symbols := "!@#$%^&*()-_=+[]{}|;:,.<>?"

	var result strings.Builder
	result.WriteByte(uppercase[MustRandomInt(len(uppercase))])
	result.WriteByte(lowercase[MustRandomInt(len(lowercase))])
	result.WriteByte(digits[MustRandomInt(len(digits))])

	// Fill the rest with random characters from all sets
	allChars := uppercase + lowercase + digits + symbols
	for i := 3; i < length; i++ {
		result.WriteByte(allChars[MustRandomInt(len(allChars))])
	}

	key := ShuffleString(result.String())

	validator := NewKeyValidator()
	if err := validator.ValidateKey(key); err != nil {
		return GenerateSecureKey(length + 4)
	}

	return key, nil
}

// WriteKeyToFile writes a key to a file with proper permissions and gitignore setup
func WriteKeyToFile(key, filename string) error {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create key file: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(key); err != nil {
		return fmt.Errorf("failed to write key to file: %w", err)
	}

	if err := EnsureGitIgnore(filename); err != nil {
		fmt.Printf("Warning: Could not update .gitignore: %v\n", err)
	}

	return nil
}

// EnsureGitIgnore adds the filename to .gitignore if it's not already there
//
// TODO: move to shared utility package
func EnsureGitIgnore(filename string) error {
	gitignorePath := ".gitignore"

	content, err := os.ReadFile(gitignorePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read .gitignore: %w", err)
	}

	lines := strings.SplitSeq(string(content), "\n")
	for line := range lines {
		if strings.TrimSpace(line) == filename {
			return nil
		}
	}

	file, err := os.OpenFile(gitignorePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open .gitignore: %w", err)
	}
	defer file.Close()

	if _, err := fmt.Fprintf(file, "\n# Weather API encryption key\n%s\n", filename); err != nil {
		return fmt.Errorf("failed to write to .gitignore: %w", err)
	}

	return nil
}

// MustRandomInt generates a cryptographically secure random integer in [0, n)
//
// TODO: move to shared utility package
func MustRandomInt(n int) int {
	if n <= 0 {
		panic("mustRandomInt: n must be positive")
	}

	max := big.NewInt(int64(n))
	randomInt, err := rand.Int(rand.Reader, max)
	if err != nil {
		panic(fmt.Sprintf("mustRandomInt: failed to generate random number: %v", err))
	}

	return int(randomInt.Int64())
}

// ShuffleString randomly shuffles the characters in a string
//
// TODO: move to shared utility package
func ShuffleString(s string) string {
	runes := []rune(s)
	for i := len(runes) - 1; i > 0; i-- {
		j := MustRandomInt(i + 1)
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func promptForKey(prompt string) (string, error) {
	fmt.Print(prompt)
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	fmt.Println()
	return string(bytePassword), nil
}
