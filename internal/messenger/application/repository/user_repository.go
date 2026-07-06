package repository

type UserReader interface {
}

type UserWriter interface {
}

type UserRepository interface {
	UserReader
	UserWriter
}
