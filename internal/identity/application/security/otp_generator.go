package security

type OTPGenerator interface {
	Generate(length int) (string, error)
}
