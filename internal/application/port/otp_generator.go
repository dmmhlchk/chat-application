package port

type OTPGenerator interface {
	Generate(length int) (string, error)
}
