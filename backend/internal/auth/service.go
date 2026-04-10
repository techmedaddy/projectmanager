package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"

	"taskflow/backend/internal/users"
)

var (
	// ErrEmailTaken is returned when a registration email is already in use.
	ErrEmailTaken = errors.New("email already in use")
	// ErrInvalidCredentials is returned when login credentials do not match a
	// user record.
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrInvalidToken is returned when an access token is missing, malformed, or
	// no longer maps to a valid user.
	ErrInvalidToken = errors.New("invalid token")
)

type userRepository interface {
	Create(ctx context.Context, params users.CreateParams) (users.User, error)
	GetByEmail(ctx context.Context, email string) (users.User, error)
	GetByID(ctx context.Context, id string) (users.User, error)
}

// Claims represents the JWT access token claims used by the API.
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// Service contains authentication business logic independent of HTTP.
type Service struct {
	usersRepo      userRepository
	jwtSecret      []byte
	jwtExpiryHours int
	bcryptCost     int
}

// NewService constructs an auth service with the dependencies needed for
// registration and login.
func NewService(usersRepo userRepository, jwtSecret string, jwtExpiryHours, bcryptCost int) *Service {
	return &Service{
		usersRepo:      usersRepo,
		jwtSecret:      []byte(jwtSecret),
		jwtExpiryHours: jwtExpiryHours,
		bcryptCost:     bcryptCost,
	}
}

// Register creates a new user with a normalized email and bcrypt-hashed
// password.
func (s *Service) Register(ctx context.Context, name, email, password string) (users.User, error) {
	name = strings.TrimSpace(name)
	email = NormalizeEmail(email)

	_, err := s.usersRepo.GetByEmail(ctx, email)
	switch {
	case err == nil:
		return users.User{}, ErrEmailTaken
	case !errors.Is(err, pgx.ErrNoRows):
		return users.User{}, fmt.Errorf("check existing user: %w", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), s.bcryptCost)
	if err != nil {
		return users.User{}, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.usersRepo.Create(ctx, users.CreateParams{
		Name:     name,
		Email:    email,
		Password: string(hashedPassword),
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return users.User{}, ErrEmailTaken
		}

		return users.User{}, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

// Login verifies a user's credentials and returns a signed JWT access token.
func (s *Service) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.usersRepo.GetByEmail(ctx, NormalizeEmail(email))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrInvalidCredentials
		}

		return "", fmt.Errorf("get user for login: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	token, err := s.generateAccessToken(user)
	if err != nil {
		return "", fmt.Errorf("generate access token: %w", err)
	}

	return token, nil
}

func (s *Service) generateAccessToken(user users.User) (string, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(time.Duration(s.jwtExpiryHours) * time.Hour)

	claims := Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// ValidateAccessToken parses and validates a JWT access token.
func (s *Service) ValidateAccessToken(tokenString string) (Claims, error) {
	claims := Claims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		&claims,
		func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, ErrInvalidToken
			}

			return s.jwtSecret, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil || !token.Valid {
		return Claims{}, ErrInvalidToken
	}

	if claims.UserID == "" || claims.Email == "" {
		return Claims{}, ErrInvalidToken
	}

	return claims, nil
}

// Authenticate resolves a bearer token into the current authenticated user.
func (s *Service) Authenticate(ctx context.Context, tokenString string) (CurrentUser, error) {
	claims, err := s.ValidateAccessToken(tokenString)
	if err != nil {
		return CurrentUser{}, err
	}

	user, err := s.usersRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return CurrentUser{}, ErrInvalidToken
		}

		return CurrentUser{}, fmt.Errorf("get current user: %w", err)
	}

	if user.Email != claims.Email {
		return CurrentUser{}, ErrInvalidToken
	}

	return CurrentUser{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}, nil
}
