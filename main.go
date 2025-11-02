package main

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Account struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	RefreshToken string `json:"refresh_token"`
	ClientID     string `gorm:"uniqueIndex" json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type Metric struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	CurrentCount int       `json:"current_count"`
	TotalCount   int       `json:"total_count"`
	UsedCount    int       `json:"used_count"`
	APICallCount int       `json:"api_call_count"`
	Timestamp    time.Time `gorm:"index" json:"timestamp"`
	Hour         string    `gorm:"index" json:"hour"`
}

var db *gorm.DB

func initDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("accounts.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	err = db.AutoMigrate(&Account{}, &Metric{})
	if err != nil {
		panic("failed to migrate database")
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		println("No .env file found, proceeding with environment variables")
	}
	initDB()

	mode := os.Getenv("GIN_MODE")
	if mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.POST("/api/auth", authenticate)
	r.GET("/api/accounts", getAccounts)
	r.POST("/api/accounts", addAccounts)
	r.GET("/api/metrics", getMetrics)

	r.Static("/assets", "./frontend/dist/assets")
	r.StaticFile("/", "./frontend/dist/index.html")
	r.NoRoute(func(c *gin.Context) {
		c.File("./frontend/dist/index.html")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	_, err = gin.DefaultWriter.Write([]byte("Starting server on port " + port + "\n"))
	if err != nil {
		panic(err)
	}
	err = r.Run(":" + port)
	if err != nil {
		panic(err)
	}
}

func authenticate(c *gin.Context) {
	var input struct {
		Passkey string `json:"passkey"`
	}
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if input.Passkey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "passkey required"})
		return
	}

	envPasskey := os.Getenv("PASSKEY")
	if envPasskey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server configuration error"})
		return
	}

	if input.Passkey == envPasskey {
		c.JSON(http.StatusOK, gin.H{"token": input.Passkey})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid passkey"})
	}
}

func getAccounts(c *gin.Context) {
	auth := c.GetHeader("Authorization")
	if auth == "" || len(auth) < 7 || auth[:7] != "Bearer " {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	countStr := c.DefaultQuery("count", "0")
	count, err := strconv.Atoi(countStr)
	if err != nil || count < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid count parameter"})
		return
	}

	var accounts []Account
	if count > 0 {
		var total int64
		db.Model(&Account{}).Count(&total)
		if int(total) < count {
			c.JSON(http.StatusBadRequest, gin.H{"error": "insufficient accounts", "available": total, "requested": count})
			return
		}

		db.Limit(count).Find(&accounts)
		for _, account := range accounts {
			db.Delete(&account)
		}

		var lastMetric Metric
		db.Order("id desc").First(&lastMetric)

		var current int64
		db.Model(&Account{}).Count(&current)

		now := time.Now()
		hour := now.Format("2006-01-02 15:00:00")

		var hourMetric Metric
		if db.Where("hour = ?", hour).First(&hourMetric).Error == nil {
			hourMetric.CurrentCount = int(current)
			hourMetric.TotalCount = lastMetric.TotalCount
			hourMetric.UsedCount = lastMetric.UsedCount + count
			hourMetric.APICallCount = lastMetric.APICallCount + 1
			hourMetric.Timestamp = now
			db.Save(&hourMetric)
		} else {
			db.Create(&Metric{
				CurrentCount: int(current),
				TotalCount:   lastMetric.TotalCount,
				UsedCount:    lastMetric.UsedCount + count,
				APICallCount: lastMetric.APICallCount + 1,
				Timestamp:    now,
				Hour:         hour,
			})
		}
	} else {
		db.Find(&accounts)
	}
	c.JSON(http.StatusOK, accounts)
}

func addAccounts(c *gin.Context) {
	var input struct {
		Accounts []Account `json:"accounts"`
	}
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var lastMetric Metric
	db.Order("id desc").First(&lastMetric)

	added := 0
	for _, account := range input.Accounts {
		if err := db.Create(&account).Error; err == nil {
			added++
		}
	}

	var current int64
	db.Model(&Account{}).Count(&current)

	now := time.Now()
	hour := now.Format("2006-01-02 15:00:00")

	var hourMetric Metric
	if db.Where("hour = ?", hour).First(&hourMetric).Error == nil {
		hourMetric.CurrentCount = int(current)
		hourMetric.TotalCount = lastMetric.TotalCount + added
		hourMetric.Timestamp = now
		db.Save(&hourMetric)
	} else {
		db.Create(&Metric{
			CurrentCount: int(current),
			TotalCount:   lastMetric.TotalCount + added,
			UsedCount:    lastMetric.UsedCount,
			APICallCount: lastMetric.APICallCount,
			Timestamp:    now,
			Hour:         hour,
		})
	}

	c.JSON(http.StatusOK, gin.H{"message": "accounts added"})
}

func getMetrics(c *gin.Context) {
	last24Hours := time.Now().Add(-24 * time.Hour)
	var metrics []Metric
	db.Where("timestamp >= ?", last24Hours).Order("timestamp asc").Find(&metrics)
	c.JSON(http.StatusOK, metrics)
}
