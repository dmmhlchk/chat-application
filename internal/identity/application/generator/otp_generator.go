package generator

type OTPGenerator interface {
	Generate(length int) (string, error)
}
