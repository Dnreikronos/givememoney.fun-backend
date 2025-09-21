package utils

import (
	"errors"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

const (
	// BcryptCost defines the cost factor for bcrypt hashing
	// Higher values provide better security but slower performance
	BcryptCost = 12

	// Password validation constants
	MinPasswordLength = 8
	MaxPasswordLength = 128
)

var (
	// ErrPasswordTooShort is returned when password is too short
	ErrPasswordTooShort = errors.New("password must be at least 8 characters long")

	// ErrPasswordTooLong is returned when password is too long
	ErrPasswordTooLong = errors.New("password must be less than 128 characters long")

	// ErrPasswordTooWeak is returned when password doesn't meet strength requirements
	ErrPasswordTooWeak = errors.New("password must contain at least one lowercase letter, one uppercase letter, and one number")

	// ErrInvalidPassword is returned when password verification fails
	ErrInvalidPassword = errors.New("invalid password")
)

// Password strength validation patterns
var (
	hasLowerCase = regexp.MustCompile(`[a-z]`)
	hasUpperCase = regexp.MustCompile(`[A-Z]`)
	hasNumber    = regexp.MustCompile(`[0-9]`)
)

// HashPassword generates a bcrypt hash of the given password
func HashPassword(password string) (string, error) {
	if err := ValidatePasswordStrength(password); err != nil {
		return "", err
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return "", err
	}

	return string(hashedBytes), nil
}

// VerifyPassword verifies a password against its hash
func VerifyPassword(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrInvalidPassword
		}
		return err
	}
	return nil
}

// ValidatePasswordStrength checks if password meets security requirements
func ValidatePasswordStrength(password string) error {
	// Check length
	if len(password) < MinPasswordLength {
		return ErrPasswordTooShort
	}

	if len(password) > MaxPasswordLength {
		return ErrPasswordTooLong
	}

	// Check for required character types
	if !hasLowerCase.MatchString(password) {
		return ErrPasswordTooWeak
	}

	if !hasUpperCase.MatchString(password) {
		return ErrPasswordTooWeak
	}

	if !hasNumber.MatchString(password) {
		return ErrPasswordTooWeak
	}

	return nil
}

// IsPasswordSecure performs comprehensive password validation
func IsPasswordSecure(password string) bool {
	return ValidatePasswordStrength(password) == nil
}