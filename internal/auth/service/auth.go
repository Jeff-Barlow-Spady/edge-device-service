package service

import (
    "crypto/rand"
    "encoding/hex"
    "encoding/json"
    "os"
    "sync"
    "time"

    "github.com/dgrijalva/jwt-go"
    "golang.org/x/crypto/bcrypt"
)

type AuthService struct {
    secretKey   []byte
    tokenExpiry time.Duration
    usersFile   string
    users      map[string]User
    mu         sync.RWMutex
}

func NewAuthService() *AuthService {
    secretKey := os.Getenv("JWT_SECRET_KEY")
    if secretKey == "" {
        key := make([]byte, 32)
        rand.Read(key)
        secretKey = hex.EncodeToString(key)
    }

    service := &AuthService{
        secretKey:   []byte(secretKey),
        tokenExpiry: 24 * time.Hour,
        usersFile:   "auth/users.json",
        users:       make(map[string]User),
    }

    service.loadUsers()
    return service
}

func (s *AuthService) loadUsers() {
    s.mu.Lock()
    defer s.mu.Unlock()

    data, err := os.ReadFile(s.usersFile)
    if err != nil {
        return
    }

    json.Unmarshal(data, &s.users)
}

func (s *AuthService) saveUsers() error {
    s.mu.RLock()
    defer s.mu.RUnlock()

    data, err := json.Marshal(s.users)
    if err != nil {
        return err
    }

    return os.WriteFile(s.usersFile, data, 0644)
}

func (s *AuthService) CreateUser(username, password string) bool {
    s.mu.Lock()
    defer s.mu.Unlock()

    if _, exists := s.users[username]; exists {
        return false
    }

    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return false
    }

    s.users[username] = User{
        Hash:      string(hash),
        CreatedAt: time.Now(),
    }

    s.saveUsers()
    return true
}

func (s *AuthService) VerifyUser(username, password string) bool {
    s.mu.RLock()
    defer s.mu.RUnlock()

    user, exists := s.users[username]
    if !exists {
        return false
    }

    return bcrypt.CompareHashAndPassword([]byte(user.Hash), []byte(password)) == nil
}

func (s *AuthService) CreateToken(username string) string {
    token := jwt.New(jwt.SigningMethodHS256)
    claims := token.Claims.(jwt.MapClaims)
    claims["sub"] = username
    claims["exp"] = time.Now().Add(s.tokenExpiry).Unix()

    tokenString, _ := token.SignedString(s.secretKey)
    return tokenString
}

func (s *AuthService) VerifyToken(tokenString string) (string, bool) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        return s.secretKey, nil
    })

    if err != nil || !token.Valid {
        return "", false
    }

    claims := token.Claims.(jwt.MapClaims)
    username := claims["sub"].(string)

    s.mu.RLock()
    _, exists := s.users[username]
    s.mu.RUnlock()

    return username, exists
}
