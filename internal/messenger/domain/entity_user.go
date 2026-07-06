package domain

import "time"

type User struct {
	ID             string
	Username       string
	Phone          string
	FirstName      string
	LastName       string
	BirdDate       time.Time
	BioDescription string
	ProfileURL     string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func NewUser(
	userID string,
	username string,
	phone string,
	firstName string,
	lastName string,
	birdDate time.Time,
	bioDescription string,
	profileURL string,
) *User {
	now := time.Now().UTC()

	return &User{
		ID:             userID,
		Username:       username,
		Phone:          phone,
		FirstName:      firstName,
		LastName:       lastName,
		BirdDate:       birdDate,
		BioDescription: bioDescription,
		ProfileURL:     profileURL,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}
