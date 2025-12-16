package service

import (
	"context"
	"errors"
	"expense-tracker/auth-service/internal/model"
	"expense-tracker/auth-service/internal/repository"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// AuthService handles authentication business logic
// This is the "service layer" - it contains the core business rules
// It uses the repository (data access) and JWT service (token generation)
type AuthService struct {
	userRepo       repository.UserRepository
	jwtService     *JWTService
	eventPublisher *EventPublisher // Optional - can be nil if not configured
}

// NewAuthService creates a new authentication service
// Dependency injection: we pass dependencies as parameters
// This makes testing easier and follows clean architecture
func NewAuthService(userRepo repository.UserRepository, jwtService *JWTService) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		jwtService: jwtService,
	}
}

// SetEventPublisher sets the event publisher (optional)
func (s *AuthService) SetEventPublisher(publisher *EventPublisher) {
	s.eventPublisher = publisher
}

// Register creates a new user account
// Steps:
// 1. Check if user already exists
// 2. Hash the password (NEVER store plain passwords!)
// 3. Create the user in database
// 4. Generate a JWT token
// 5. Return user info and token
func (s *AuthService) Register(ctx context.Context, req *model.RegisterRequest) (*model.AuthResponse, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		// User already exists - return an error
		// In Go, we use errors.New() or fmt.Errorf() to create errors
		return nil, errors.New("user with this email already exists")
	}

	// Hash the password using bcrypt
	// bcrypt automatically adds a salt and uses a cost factor (how many rounds)
	// Cost 10-12 is recommended for production (higher = more secure but slower)
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create the user
	user := model.NewUser(req.Email, string(passwordHash), req.Name)

	// Save to database
	err = s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	// Publish event (non-blocking, async)
	if s.eventPublisher != nil {
		event := &Event{
			EventType: "user.registered",
			UserID:    user.ID,
			UserEmail: user.Email,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"user_id": user.ID,
				"email":   user.Email,
				"name":    user.Name,
			},
		}
		s.eventPublisher.PublishEventAsync(ctx, event)
	}

	// Generate JWT token
	token, err := s.jwtService.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	// Return the response
	return &model.AuthResponse{
		UserID: user.ID,
		Email:  user.Email,
		Name:   user.Name,
		Token:  token,
	}, nil
}

// Login authenticates a user and returns a token
// Steps:
// 1. Find user by email
// 2. Compare provided password with stored hash
// 3. If match, generate token
// 4. Return user info and token
func (s *AuthService) Login(ctx context.Context, req *model.LoginRequest) (*model.AuthResponse, error) {
	// Find user by email
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		// User not found - don't reveal if email exists (security best practice)
		return nil, errors.New("invalid email or password")
	}

	// Compare password with hash
	// bcrypt.CompareHashAndPassword returns nil if passwords match
	// This is secure because it uses constant-time comparison (prevents timing attacks)
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		// Password doesn't match
		return nil, errors.New("invalid email or password")
	}

	// Generate JWT token
	token, err := s.jwtService.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	// Return the response
	return &model.AuthResponse{
		UserID: user.ID,
		Email:  user.Email,
		Name:   user.Name,
		Token:  token,
	}, nil
}

// ValidateToken checks if a JWT token is valid
// This is used by middleware to protect routes
func (s *AuthService) ValidateToken(ctx context.Context, tokenString string) (*model.ValidateResponse, error) {
	// Validate the token using JWT service
	claims, err := s.jwtService.ValidateToken(tokenString)
	if err != nil {
		return &model.ValidateResponse{
			Valid: false,
		}, nil // Return valid=false, but no error (token is just invalid)
	}

	// Optionally, verify user still exists in database
	// This ensures the user wasn't deleted after token was issued
	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return &model.ValidateResponse{
			Valid: false,
		}, nil
	}
	if user == nil {
		return &model.ValidateResponse{
			Valid: false,
		}, nil
	}

	// Token is valid!
	return &model.ValidateResponse{
		Valid:  true,
		UserID: claims.UserID,
		Email:  claims.Email,
	}, nil
}
