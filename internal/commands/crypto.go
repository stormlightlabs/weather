package commands

import (
	"bufio"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/charmbracelet/log"
	"github.com/urfave/cli/v3"
	"golang.org/x/crypto/scrypt"
	"golang.org/x/term"
)

func encryptEnvFile(_ context.Context, cmd *cli.Command, logger *log.Logger) error {
	filePath := cmd.String("file")
	key := cmd.String("key")

	if key == "" {
		var err error
		key, err = promptForKey("Enter encryption key: ")
		if err != nil {
			return fmt.Errorf("failed to read key: %w", err)
		}
	}

	logger.Info("Encrypting environment file", "file", filePath)
	return processEnvFile(filePath, key, true, logger)
}

func decryptEnvFile(_ context.Context, cmd *cli.Command, logger *log.Logger) error {
	filePath := cmd.String("file")
	key := cmd.String("key")

	if key == "" {
		var err error
		key, err = promptForKey("Enter decryption key: ")
		if err != nil {
			return fmt.Errorf("failed to read key: %w", err)
		}
	}

	logger.Info("Decrypting environment file", "file", filePath)
	return processEnvFile(filePath, key, false, logger)
}

func processEnvFile(filePath, key string, encrypt bool, logger *log.Logger) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "=") && !strings.HasPrefix(line, "#") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				if encrypt {
					encrypted, err := encryptValue(parts[1], key)
					if err != nil {
						return fmt.Errorf("failed to encrypt value for %s: %w", parts[0], err)
					}
					lines = append(lines, fmt.Sprintf("%s=%s", parts[0], encrypted))
				} else {
					decrypted, err := decryptValue(parts[1], key)
					if err != nil {
						return fmt.Errorf("failed to decrypt value for %s: %w", parts[0], err)
					}
					lines = append(lines, fmt.Sprintf("%s=%s", parts[0], decrypted))
				}
			} else {
				lines = append(lines, line)
			}
		} else {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	backupFile := filePath + ".backup"
	if err := os.Rename(filePath, backupFile); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	outFile, err := os.Create(filePath)
	if err != nil {
		os.Rename(backupFile, filePath) // Restore backup
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	for _, line := range lines {
		if _, err := fmt.Fprintln(outFile, line); err != nil {
			return fmt.Errorf("failed to write line: %w", err)
		}
	}

	os.Remove(backupFile) // Remove backup on success
	operation := "Encryption"
	if !encrypt {
		operation = "Decryption"
	}
	logger.Info(operation+" completed successfully", "file", filePath)
	return nil
}

func encryptValue(value, key string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	derivedKey, err := scrypt.Key([]byte(key), salt, 32768, 8, 1, 32)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(derivedKey)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nil, nonce, []byte(value), nil)

	// Format: salt:nonce:ciphertext (all hex encoded)
	return fmt.Sprintf("%s:%s:%s",
		hex.EncodeToString(salt),
		hex.EncodeToString(nonce),
		hex.EncodeToString(ciphertext)), nil
}

func decryptValue(encryptedValue, key string) (string, error) {
	parts := strings.Split(encryptedValue, ":")
	if len(parts) != 3 {
		// If it's not encrypted format, return as-is
		return encryptedValue, nil
	}

	salt, err := hex.DecodeString(parts[0])
	if err != nil {
		return encryptedValue, nil // Not encrypted format
	}

	nonce, err := hex.DecodeString(parts[1])
	if err != nil {
		return encryptedValue, nil // Not encrypted format
	}

	ciphertext, err := hex.DecodeString(parts[2])
	if err != nil {
		return encryptedValue, nil // Not encrypted format
	}

	derivedKey, err := scrypt.Key([]byte(key), salt, 32768, 8, 1, 32)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(derivedKey)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
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
