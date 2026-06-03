package crypto

import (
	"crypto/rand"
	"fmt"
	"identity-service/internal/domain"
	"math/big"
)

// Compile-time interface guard
var _ domain.OTPGenerator = (*SecureOTPGenerator)(nil)

type SecureOTPGenerator struct{}

func NewSecureOTPGenerator() *SecureOTPGenerator {
	return &SecureOTPGenerator{}
}

// Generate builds a numeric string of the specified length (e.g., length=6 -> "482019")
func (g *SecureOTPGenerator) Generate(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("invalid OTP length: %d", length)
	}

	digits := "0123456789"
	otp := make([]byte, length)
	maxIdx := big.NewInt(int64(len(digits)))

	for i := 0; i < length; i++ {
		idx, err := rand.Int(rand.Reader, maxIdx)
		if err != nil {
			return "", fmt.Errorf("failed to generate random secure digit: %w", err)
		}
		otp[i] = digits[idx.Int64()]
	}

	return string(otp), nil
}
