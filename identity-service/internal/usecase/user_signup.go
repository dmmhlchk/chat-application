package usecase

import (
	"context"
	"errors"
	"fmt"
	"identity-service/internal/domain"
)

// 1. Determine Input and Output data
// 1.1. SignUpInput defines the data required to register a new user
type SignUpInput struct {
	Username string
	Phone    string
	Password string
}

// 1.2. SignUpOutput defines what data we return to the delivery layer upon success
type SignUpOutput struct {
	UserId int
}

// 2. Determine the dependencies
// 2.1. SignUpUseCase coordinates domain layer to register a user
type SignUpUseCase struct {
	userRepo  domain.UserRepo
	pwdHasher domain.PasswordHasher
}

// 2.2. NewSignUpUseCase is a constructor that handles dependency injection
func NewSignUpUseCase(userRepo domain.UserRepo, hasher domain.PasswordHasher) *SignUpUseCase {
	return &SignUpUseCase{
		userRepo:  userRepo,
		pwdHasher: hasher,
	}
}

// 3. Execute runs the actual step-by-step registration business flow
func (uc *SignUpUseCase) Execute(ctx context.Context, input SignUpInput) (*SignUpOutput, error) {
	// 1. Validate input basic constraints
	if input.Username == "" || input.Phone == "" || input.Password == "" {
		return nil, errors.New("required fields were not filled")
	}

	// 2. Check the domain repository to see if this user already exists
	exists, err := uc.userRepo.ExistsByPhoneOrUsername(ctx, input.Phone, input.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to verify account uniqueness: %w", err)
	}
	if exists {
		return nil, errors.New("phone number or username is already registered")
	}

	// 3. Secure the password using our domain's abstract PasswordHasher interface
	hashedPassword, err := uc.pwdHasher.Hash(input.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to process password: %w", err)
	}

	// 4. Register a new user
	newUser := domain.NewUser(0, input.Username, input.Phone, hashedPassword)
	err = uc.userRepo.Create(ctx, newUser)
	if err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	// 5. Return success
	return &SignUpOutput{
		UserId: newUser.ID,
	}, nil
}
