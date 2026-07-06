package repository

type ParticipantReader interface {
}

type ParticipantWriter interface {
}

type ParticipantRepository interface {
	ParticipantReader
	ParticipantWriter
}
