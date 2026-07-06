package domain

type MessageType string

const (
	MessageTypePlainText MessageType = "plain text"
	MessageTypePhoto     MessageType = "photo"
	MessageTypeVideo     MessageType = "video"
	MessageTypeAlbum     MessageType = "album"
)

func (m MessageType) IsValid() bool {
	switch m {
	case MessageTypePlainText, MessageTypePhoto, MessageTypeVideo, MessageTypeAlbum:
		return true
	}

	return false
}
