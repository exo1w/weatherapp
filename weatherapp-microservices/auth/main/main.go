package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"time"
	"weatherapp.com/auth/authdb"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	defaultDBHost     = "127.0.0.1"
	defaultDBUser     = "authuser"
	defaultDBPassword = "authpassword"
	defaultDBName     = "weatherapp"
	defaultDBPort     = "3306"
	defaultSecretKey  = "xco0sr0fh4e52x03g9mv"
	defaultAuthPort   = "8080"
)

var (
	dbHost     = getEnv("DB_HOST", defaultDBHost)
	dbUser     = getEnv("DB_USER", defaultDBUser)
	dbPassword = getEnv("DB_PASSWORD", defaultDBPassword)
	dbName     = getEnv("DB_NAME", defaultDBName)
	dbPort     = getEnv("DB_PORT", defaultDBPort)
	secretKey  = getEnv("SECRET_KEY", defaultSecretKey)
	authPort   = getEnv("AUTH_PORT", defaultAuthPort)
)

// Prometheus metrics
var requestCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "auth_service_requests_total",
		Help: "Number of requests processed by endpoint",
	},
	[]string{"endpoint"},
)

func init() {
	prometheus.MustRegister(requestCounter)
}

type UserCreds struct {
	Username string `json:"user_name"`
	Password string `json:"user_password"`
}

func main() {
	// اتصال بقاعدة البيانات
	db, err := authdb.Connect(dbUser, dbPassword, dbHost, dbPort)
	if err != nil {
		fmt.Println("Database connection error:", err)
		panic(err)
	}

	// إنشاء قاعدة البيانات والجداول إذا لم تكن موجودة
	if err := authdb.CreateDB(db, dbName); err != nil {
		fmt.Println("Error creating database:", err)
		panic(err)
	}
	if err := authdb.CreateTables(db, dbName); err != nil {
		fmt.Println("Error creating tables:", err)
		panic(err)
	}

	// إعداد Gin router و CORS
	router := gin.Default()
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"*"}
	corsConfig.AllowCredentials = true
	router.Use(cors.New(corsConfig))

	// Endpoints
	router.GET("/", health)
	router.POST("/users", createUser)
	router.POST("/users/:id", loginUser)

	// Prometheus metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	fmt.Println("Auth service running on port", authPort)
	router.Run(":" + authPort)
}

// دالة لاسترجاع المتغيرات البيئية
func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}

// Health endpoint
func health(c *gin.Context) {
	requestCounter.WithLabelValues("/").Inc() // تحديث العداد

	db, err := authdb.Connect(dbUser, dbPassword, dbHost, dbPort)
	if err != nil || db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database not reachable"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": "Auth service is running"})
}

// تسجيل الدخول
func loginUser(c *gin.Context) {
	requestCounter.WithLabelValues("/users/:id").Inc() // تحديث العداد

	var uc UserCreds
	if err := c.BindJSON(&uc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	db, err := authdb.Connect(dbUser, dbPassword, dbHost, dbPort)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection failed"})
		return
	}

	user, err := authdb.GetUserByName(uc.Username, db, dbName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// تشفير الباسوورد للتحقق
	passwordHash := md5.Sum([]byte(uc.Password))
	if user != (authdb.User{}) && user.Password == hex.EncodeToString(passwordHash[:]) {
		token, err := GenerateJWT(user.Name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"JWT": token})
	} else {
		c.JSON(http.StatusForbidden, gin.H{"error": "Bad credentials"})
	}
}

// إنشاء مستخدم جديد
func createUser(c *gin.Context) {
	requestCounter.WithLabelValues("/users").Inc() // تحديث العداد

	var u authdb.User
	if err := c.BindJSON(&u); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	db, err := authdb.Connect(dbUser, dbPassword, dbHost, dbPort)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection failed"})
		return
	}

	created, err := authdb.CreateUser(db, u, dbName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error creating user: %v", err)})
		return
	}
	if !created {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "User already exists"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": "User added successfully"})
}

// إنشاء JWT
func GenerateJWT(userName string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["authorized"] = true
	claims["username"] = userName
	claims["exp"] = time.Now().Add(time.Minute * 30).Unix()
	return token.SignedString([]byte(secretKey))
}