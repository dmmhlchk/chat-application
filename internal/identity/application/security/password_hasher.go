package security

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hashedPassword string, password string) (bool, error)
}
