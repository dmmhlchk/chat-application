package domain

type ChatType string

const (
	ChatTypeDirect ChatType = "direct"
	ChatTypeGroup  ChatType = "group"
)

func (c ChatType) IsValid() bool {
	switch c {
	case ChatTypeDirect, ChatTypeGroup:
		return true
	}

	return false
}
