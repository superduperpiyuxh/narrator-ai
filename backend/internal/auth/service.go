package auth

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	APIKey       string `json:"api_key,omitempty"`
	OpenRouterKey string `json:"-"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

type Service struct {
	db        *sql.DB
	jwtSecret []byte
}

func NewService(db *sql.DB, jwtSecret string) *Service {
	return &Service{
		db:        db,
		jwtSecret: []byte(jwtSecret),
	}
}

func (s *Service) Signup(email, password string) (*User, error) {
	var exists int
	err := s.db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", email).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("check existing user: %w", err)
	}
	if exists > 0 {
		return nil, fmt.Errorf("email already registered")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	id := uuid.New().String()
	apiKey := "nai_" + uuid.New().String()[:16]
	now := time.Now().UTC().Format(time.RFC3339)

	_, err = s.db.Exec(`
		INSERT INTO users (id, email, password_hash, api_key, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		id, email, string(hash), apiKey, now, now)
	if err != nil {
		return nil, fmt.Errorf("insert user: %w", err)
	}

	return &User{
		ID:        id,
		Email:     email,
		APIKey:    apiKey,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (s *Service) Login(email, password string) (string, *User, error) {
	var user User
	var passwordHash string
	err := s.db.QueryRow(`
		SELECT id, email, password_hash, api_key, openrouter_key, created_at, updated_at
		FROM users WHERE email = ?`, email).Scan(
		&user.ID, &user.Email, &passwordHash, &user.APIKey, &user.OpenRouterKey,
		&user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return "", nil, fmt.Errorf("invalid credentials")
	}
	if err != nil {
		return "", nil, fmt.Errorf("query user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return "", nil, fmt.Errorf("invalid credentials")
	}

	token, err := s.generateToken(user.ID, user.Email)
	if err != nil {
		return "", nil, fmt.Errorf("generate token: %w", err)
	}

	user.PasswordHash = ""
	return token, &user, nil
}

func (s *Service) GetUserByID(id string) (*User, error) {
	var user User
	err := s.db.QueryRow(`
		SELECT id, email, api_key, openrouter_key, created_at, updated_at
		FROM users WHERE id = ?`, id).Scan(
		&user.ID, &user.Email, &user.APIKey, &user.OpenRouterKey,
		&user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query user: %w", err)
	}
	return &user, nil
}

func (s *Service) GetUserByAPIKey(apiKey string) (*User, error) {
	var user User
	err := s.db.QueryRow(`
		SELECT id, email, api_key, openrouter_key, created_at, updated_at
		FROM users WHERE api_key = ?`, apiKey).Scan(
		&user.ID, &user.Email, &user.APIKey, &user.OpenRouterKey,
		&user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query user by api key: %w", err)
	}
	return &user, nil
}

func (s *Service) UpdateOpenRouterKey(userID, apiKey string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.Exec(`
		UPDATE users SET openrouter_key = ?, updated_at = ? WHERE id = ?`,
		apiKey, now, userID)
	return err
}

func (s *Service) generateToken(userID, email string) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "narrator-ai",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

type contextKey string

const UserIDKey contextKey = "user_id"

func ContextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}
