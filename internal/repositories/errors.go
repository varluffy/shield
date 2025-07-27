package repositories

import "errors"

var (
	// ErrUserNotFound 用户未找到错误
	ErrUserNotFound = errors.New("user not found")

	// ErrUserAlreadyExists 用户已存在错误
	ErrUserAlreadyExists = errors.New("user already exists")

	// ErrInvalidInput 无效输入错误
	ErrInvalidInput = errors.New("invalid input")

	// ErrDatabaseConnection 数据库连接错误
	ErrDatabaseConnection = errors.New("database connection error")
)
