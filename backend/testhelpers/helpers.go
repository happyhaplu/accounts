// Package testhelpers provides shared setup utilities for integration tests.
package testhelpers

import (
"os"
"time"

"outcraftly/accounts/config"
"outcraftly/accounts/database"
"outcraftly/accounts/middleware"
"outcraftly/accounts/models"
"outcraftly/accounts/routes"

	"github.com/glebarez/sqlite"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
"gorm.io/gorm"
"gorm.io/gorm/logger"
)

const TestJWTSecret = "test-integration-secret"

// SetupTestDB opens a fresh in-memory SQLite database and runs AutoMigrate.
func SetupTestDB() *gorm.DB {
os.Setenv("JWT_SECRET", TestJWTSecret)
os.Setenv("APP_URL", "http://localhost:5173")
os.Setenv("SMTP_HOST", "")
os.Setenv("STRIPE_SECRET_KEY", "")
os.Setenv("ADMIN_SECRET", "test-admin-secret")

db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared&_fk=1"), &gorm.Config{
Logger: logger.Default.LogMode(logger.Silent),
})
if err != nil {
panic("testhelpers: could not open test db: " + err.Error())
}

if err := db.AutoMigrate(
&models.User{},
&models.Workspace{},
&models.WorkspaceMember{},
&models.WorkspaceInvite{},
&models.Product{},
&models.Subscription{},
&models.BillingCustomer{},
); err != nil {
panic("testhelpers: AutoMigrate failed: " + err.Error())
}

database.DB = db

// Populate config.Cfg so handlers/middleware can reference it.
config.Load()

return db
}

// NewApp builds a Fiber app with all routes registered.
func NewApp() *fiber.App {
app := fiber.New(fiber.Config{DisableStartupMessage: true})
routes.Setup(app)
return app
}

// MakeJWT issues a signed JWT for testing.
func MakeJWT(userID, email string) string {
claims := jwt.MapClaims{
"sub":   userID,
"email": email,
"role":  "user",
"iss":   "accounts.outcraftly.com",
"aud":   "reach",
"exp":   time.Now().Add(24 * time.Hour).Unix(),
"iat":   time.Now().Unix(),
}
tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
s, _ := tok.SignedString([]byte(TestJWTSecret))
return s
}

// CreateVerifiedUser inserts a verified user directly and returns user + JWT.
func CreateVerifiedUser(db *gorm.DB, email, password string) (models.User, string) {
hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
user := models.User{
ID:              uuid.New(),
Email:           email,
PasswordHash:    string(hash),
EmailVerified:   true,
ProfileComplete: true,
Name:            "Test User",
CompanyName:     "Test Corp",
}
db.Create(&user)
return user, MakeJWT(user.ID.String(), email)
}

// AuthBearer returns the full Authorization header value.
func AuthBearer(token string) string { return "Bearer " + token }

// UseProtected re-exports middleware.Protected for test packages.
var UseProtected = middleware.Protected
