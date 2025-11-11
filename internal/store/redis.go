package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"task-dashboard/internal/models"

	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	client *redis.Client
}

func NewRedisStore(addr string) *RedisStore {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   0,
	})
	return &RedisStore{client: client}
}

// User operations
func (s *RedisStore) CreateUser(ctx context.Context, user *models.User) error {
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, fmt.Sprintf("user:%s", user.ID), data, 0).Err()
}

func (s *RedisStore) GetUser(ctx context.Context, id string) (*models.User, error) {
	data, err := s.client.Get(ctx, fmt.Sprintf("user:%s", id)).Result()
	if err != nil {
		return nil, err
	}
	var user models.User
	if err := json.Unmarshal([]byte(data), &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *RedisStore) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	userID, err := s.client.Get(ctx, fmt.Sprintf("user:email:%s", email)).Result()
	if err != nil {
		return nil, err
	}
	return s.GetUser(ctx, userID)
}

func (s *RedisStore) SetEmailIndex(ctx context.Context, email, userID string) error {
	return s.client.Set(ctx, fmt.Sprintf("user:email:%s", email), userID, 0).Err()
}

// Session operations
func (s *RedisStore) CreateSession(ctx context.Context, session *models.Session) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}
	ttl := time.Until(session.ExpiresAt)
	return s.client.Set(ctx, fmt.Sprintf("session:%s", session.ID), data, ttl).Err()
}

func (s *RedisStore) GetSession(ctx context.Context, id string) (*models.Session, error) {
	data, err := s.client.Get(ctx, fmt.Sprintf("session:%s", id)).Result()
	if err != nil {
		return nil, err
	}
	var session models.Session
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return nil, err
	}
	return &session, nil
}

func (s *RedisStore) DeleteSession(ctx context.Context, id string) error {
	return s.client.Del(ctx, fmt.Sprintf("session:%s", id)).Err()
}

// Task operations
func (s *RedisStore) CreateTask(ctx context.Context, task *models.Task) error {
	data, err := json.Marshal(task)
	if err != nil {
		return err
	}

	pipe := s.client.Pipeline()
	pipe.Set(ctx, fmt.Sprintf("task:%s", task.ID), data, 0)
	pipe.ZAdd(ctx, fmt.Sprintf("user:%s:tasks", task.UserID), redis.Z{
		Score:  float64(task.CreatedAt.Unix()),
		Member: task.ID,
	})
	_, err = pipe.Exec(ctx)
	return err
}

func (s *RedisStore) GetTask(ctx context.Context, id string) (*models.Task, error) {
	data, err := s.client.Get(ctx, fmt.Sprintf("task:%s", id)).Result()
	if err != nil {
		return nil, err
	}
	var task models.Task
	if err := json.Unmarshal([]byte(data), &task); err != nil {
		return nil, err
	}
	return &task, nil
}

func (s *RedisStore) UpdateTask(ctx context.Context, task *models.Task) error {
	task.UpdatedAt = time.Now()
	data, err := json.Marshal(task)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, fmt.Sprintf("task:%s", task.ID), data, 0).Err()
}

func (s *RedisStore) DeleteTask(ctx context.Context, userID, taskID string) error {
	pipe := s.client.Pipeline()
	pipe.Del(ctx, fmt.Sprintf("task:%s", taskID))
	pipe.ZRem(ctx, fmt.Sprintf("user:%s:tasks", userID), taskID)
	_, err := pipe.Exec(ctx)
	return err
}

func (s *RedisStore) GetUserTasks(ctx context.Context, userID, status string) ([]*models.Task, error) {

}
