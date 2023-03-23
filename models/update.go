package models

import (
	"fmt"
	"strconv"
)

type Update struct {
	id int64
}

func NewUpdate(userID int64, body string) (*Update, error) {
	// making an auto increment field "id" which is similar to primary key you can say, so here if this function doesn't see "user:next-id"(which is just a filler) in list then it will start from 0, which we want
	id, err := client.Incr("update:next-id").Result()
	if err != nil {
		return nil, err
	}

	key := fmt.Sprintf("update:%d", id)

	// making a pipeline which sends all the requests to the redis server and brings back the responses in one go rather than making requests separately
	pipe := client.Pipeline()
	pipe.HSet(key, "id", id)
	pipe.HSet(key, "user_id", userID)
	pipe.HSet(key, "body", body)
	pipe.LPush("updates", id)
	pipe.LPush(fmt.Sprintf("user:%d:updates", userID), id)
	_, err = pipe.Exec()
	if err != nil {
		return nil, err
	}
	return &Update{id}, nil
}

func (update *Update) GetBody() (string, error) {
	key := fmt.Sprintf("update:%d", update.id)
	return client.HGet(key, "body").Result()
}

func (update *Update) GetUser() (*User, error) {
	key := fmt.Sprintf("update:%d", update.id)
	userID, err := client.HGet(key, "user_id").Int64()
	if err != nil {
		return nil, err
	}

	return GetUserByID(userID)
}

func QueryUpdates(key string) ([]*Update, error) {
	// get update IDs to make a list of updates
	updateIDs, err := client.LRange(key, 0, 10).Result()
	if err != nil {
		return nil, err
	}
	// make list of updates with update IDs
	updates := make([]*Update, len(updateIDs))
	for i, strID := range updateIDs {
		id, err := strconv.Atoi(strID) //converting string type IDs to int type because LRange.Result() outputs list of string
		if err != nil {
			return nil, err
		}
		updates[i] = &Update{int64(id)} //converting int to int64 since id was in int
	}

	return updates, nil
}

func GetAllUpdates() ([]*Update, error) {
	return QueryUpdates("updates")
}

func GetUpdates(userID int64) ([]*Update, error) {
	key := fmt.Sprintf("user:%d:updates", userID)
	return QueryUpdates(key)
}

func PostUpdate(userID int64, body string) error {
	_, err := NewUpdate(userID, body)
	return err
}
