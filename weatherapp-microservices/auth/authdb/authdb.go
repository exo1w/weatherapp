package authdb

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

type User struct {
	ID       int    `json:"user_id"`
	Name     string `json:"user_name"`
	Password string `json:"user_password"`
}

// Connect to MySQL server
func Connect(dbUser, dbPassword, dbHost, dbPort string) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/", dbUser, dbPassword, dbHost, dbPort)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	// تحقق من الاتصال فعليًا
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

// CreateDB creates the database if not exists
func CreateDB(db *sql.DB, dbName string) error {
	query := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", dbName)
	_, err := db.Exec(query)
	return err
}

// CreateTables creates the users table if not exists
func CreateTables(db *sql.DB, dbName string) error {
	query := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s.users (
		user_id INT AUTO_INCREMENT,
		user_name CHAR(50) NOT NULL,
		user_password CHAR(128),
		PRIMARY KEY(user_id)
	)`, dbName)
	_, err := db.Exec(query)
	return err
}

// InsertUser inserts a new user into the database
func InsertUser(db *sql.DB, user User, dbName string) error {
	passwordHash := md5.Sum([]byte(user.Password))
	query := fmt.Sprintf("INSERT INTO %s.users (user_name, user_password) VALUES (?, ?)", dbName)
	_, err := db.Exec(query, user.Name, hex.EncodeToString(passwordHash[:]))
	return err
}

// GetUserByName fetches a user by username
func GetUserByName(userName string, db *sql.DB, dbName string) (User, error) {
	var user User
	query := fmt.Sprintf("SELECT user_id, user_name, user_password FROM %s.users WHERE user_name = ?", dbName)
	row := db.QueryRow(query, userName)
	err := row.Scan(&user.ID, &user.Name, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			// المستخدم غير موجود
			return User{}, nil
		}
		// أي خطأ آخر
		return User{}, err
	}
	return user, nil
}

// CreateUser creates a new user if not exists
func CreateUser(db *sql.DB, u User, dbName string) (bool, error) {
	user, err := GetUserByName(u.Name, db, dbName)
	if err != nil {
		return false, err
	}
	if user != (User{}) {
		// المستخدم موجود بالفعل
		return false, nil
	}
	// المستخدم غير موجود → إنشاءه
	err = InsertUser(db, u, dbName)
	if err != nil {
		return false, err
	}
	return true, nil
}