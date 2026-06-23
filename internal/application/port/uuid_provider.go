package port

type UUIDProvider interface {
	Generate() string
}
