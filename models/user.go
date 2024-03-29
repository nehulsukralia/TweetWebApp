package models

import (
	"errors"
	"fmt"

	"github.com/go-redis/redis"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound  = errors.New("user not found")
	ErrInvalidLogin  = errors.New("invalid login")
	ErrUsernameTaken = errors.New("username taken")
)

type User struct {
	id int64
}

func NewUser(username string, hash []byte) (*User, error) {
	// check if username already exists
	exists, err := client.HExists("user:by-username", username).Result()
	if exists {
		return nil, ErrUsernameTaken
	}
	
	// making an auto increment field "id" which is similar to primary key you can say, so here if this function doesn't see "user:next-id"(which is just a filler) in list then it will start from 0, which we want
	id, err := client.Incr("user:next-id").Result()
	if err != nil {
		return nil, err
	}

	key := fmt.Sprintf("user:%d", id)

	// making a pipeline which sends all the requests to the redis server and brings back the responses in one go rather than making requests separately
	pipe := client.Pipeline()
	pipe.HSet(key, "id", id)
	pipe.HSet(key, "username", username)
	pipe.HSet(key, "hash", hash)
	pipe.HSet("user:by-username", username, id)
	_, err = pipe.Exec()
	if err != nil {
		return nil, err
	}
	return &User{id}, nil
}

func (user *User) GetID() (int64, error) {
	return user.id, nil
}

func (user *User) GetUsername() (string, error) {
	key := fmt.Sprintf("user:%d", user.id)
	return client.HGet(key, "username").Result()
}

func (user *User) GetHash() ([]byte, error) {
	key := fmt.Sprintf("user:%d", user.id)
	return client.HGet(key, "hash").Bytes()
}

func (user *User) Authenticate(password string) error {
	hash, err := user.GetHash()
	if err != nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword(hash, []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return ErrInvalidLogin
	}

	return err
}

func GetUserByID(id int64) (*User, error) {
	return &User{id}, nil
}

func GetUserByUsername(username string) (*User, error) {
	id, err := client.HGet("user:by-username", username).Int64()
	if err == redis.Nil {
		return nil, ErrUserNotFound
	} else if err != nil {
		return nil, err
	}

	return GetUserByID(id)
}

func RegisterUser(username, password string) error {
	// generate password hash and store to storage
	cost := bcrypt.DefaultCost
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil { // (password hashing generally fails due to cost, but here we have used the default cost, but still let's handle this error)
		return err
	}

	_, err = NewUser(username, hash)

	return err
}

func AuthenticateUser(username, password string) (*User, error) {
	user, err := GetUserByUsername(username)
	if err != nil {
		return nil, err
	}
	return user, user.Authenticate(password)
}
