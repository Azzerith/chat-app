package controllers

import (
    "errors"
    "net/http"
    "os"
    "strings"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
    "github.com/google/uuid"
    "golang.org/x/crypto/bcrypt"
    "gorm.io/gorm"

    "chat-app/models"
)


type UserController struct {
    DB        *gorm.DB
    SecretKey []byte
}

// NewUserController returns a controller with DB and secret key initialised.
func NewUserController(db *gorm.DB) *UserController {
    secret := []byte(os.Getenv("JWT_SECRET"))
    if len(secret) == 0 {
        // fallback for development; make sure JWT_SECRET env var is set in production
        secret = []byte("change-me-please")
    }
    return &UserController{DB: db, SecretKey: secret}
}

// ======== Request structs ========

type registerInput struct {
    Username string `json:"username" binding:"required,min=3,max=50"`
    Email    string `json:"email"    binding:"required,email,max=100"`
    Password string `json:"password" binding:"required,min=6,max=100"`
}

type loginInput struct {
    Email    string `json:"email"    binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

type updateInput struct {
    Username *string `json:"username" binding:"omitempty,min=3,max=50"`
    Email    *string `json:"email"    binding:"omitempty,email,max=100"`
    Password *string `json:"password" binding:"omitempty,min=6,max=100"`
    IsOnline *bool   `json:"is_online"`
}

// ======== JWT helpers & middleware ========

type jwtCustomClaims struct {
    UserID   string `json:"uid"`
    Username string `json:"uname"`
    jwt.RegisteredClaims
}

func (uc *UserController) generateToken(user *models.User, ttl time.Duration) (string, error) {
    claims := jwtCustomClaims{
        UserID:   user.ID,
        Username: user.Username,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Subject:   user.ID,
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(uc.SecretKey)
}

func (uc *UserController) parseToken(tokenStr string) (*jwtCustomClaims, error) {
    tkn, err := jwt.ParseWithClaims(tokenStr, &jwtCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
        return uc.SecretKey, nil
    })
    if err != nil {
        return nil, err
    }
    claims, ok := tkn.Claims.(*jwtCustomClaims)
    if !ok || !tkn.Valid {
        return nil, errors.New("invalid token")
    }
    return claims, nil
}

// JWTAuthMiddleware validates token and stores claims in the context
func (uc *UserController) JWTAuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing Authorization header"})
            return
        }
        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "bad Authorization header format"})
            return
        }
        claims, err := uc.parseToken(parts[1])
        if err != nil {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
            return
        }
        // attach claims to context for downstream handlers
        c.Set("userID", claims.UserID)
        c.Set("username", claims.Username)
        c.Next()
    }
}

// ======== Handler methods ========

// Register (POST /api/register)
func (uc *UserController) Register(c *gin.Context) {
    var input registerInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // uniqueness check
    var count int64
    uc.DB.Model(&models.User{}).Where("email = ? OR username = ?", input.Email, input.Username).Count(&count)
    if count > 0 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "username or email already registered"})
        return
    }

    // hash password
    hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
        return
    }

    user := models.User{
        ID:       uuid.NewString(),
        Username: input.Username,
        Email:    input.Email,
        Password: string(hashed),
        IsOnline: false,
    }

    if err := uc.DB.Create(&user).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"message": "registration successful"})
}

// Login (POST /api/login) – returns JWT token
func (uc *UserController) Login(c *gin.Context) {
    var input loginInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    var user models.User
    if err := uc.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
        return
    }

    if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
        return
    }

    // mark online and reset LastSeen
    uc.DB.Model(&user).Updates(models.User{IsOnline: true, LastSeen: nil})

    token, err := uc.generateToken(&user, 24*time.Hour)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"access_token": token, "token_type": "bearer", "expires_in": 86400})
}

// Logout (POST /api/logout)
func (uc *UserController) Logout(c *gin.Context) {
    uid, exists := c.Get("userID")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
        return
    }
    now := time.Now()
    if err := uc.DB.Model(&models.User{}).Where("id = ?", uid).Updates(models.User{IsOnline: false, LastSeen: &now}).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "logout successful"})
}

// GetUsers (GET /api/users)
func (uc *UserController) GetUsers(c *gin.Context) {
    var users []models.User
    if err := uc.DB.Find(&users).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    // hide password before sending
    for i := range users {
        users[i].Password = ""
    }
    c.JSON(http.StatusOK, users)
}

// GetUser (GET /api/users/:id)
func (uc *UserController) GetUser(c *gin.Context) {
    id := c.Param("id")
    var user models.User
    if err := uc.DB.First(&user, "id = ?", id).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    user.Password = ""
    c.JSON(http.StatusOK, user)
}

// UpdateUser (PUT /api/users/:id)
func (uc *UserController) UpdateUser(c *gin.Context) {
    id := c.Param("id")
    var input updateInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    var user models.User
    if err := uc.DB.First(&user, "id = ?", id).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // handle updates
    if input.Username != nil {
        user.Username = *input.Username
    }
    if input.Email != nil {
        user.Email = *input.Email
    }
    if input.Password != nil {
        hashed, err := bcrypt.GenerateFromPassword([]byte(*input.Password), bcrypt.DefaultCost)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
            return
        }
        user.Password = string(hashed)
    }
    if input.IsOnline != nil {
        user.IsOnline = *input.IsOnline
    }

    if err := uc.DB.Save(&user).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    user.Password = ""
    c.JSON(http.StatusOK, user)
}

// DeleteUser (DELETE /api/users/:id)
func (uc *UserController) DeleteUser(c *gin.Context) {
    id := c.Param("id")
    if err := uc.DB.Delete(&models.User{}, "id = ?", id).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "user deleted"})
}