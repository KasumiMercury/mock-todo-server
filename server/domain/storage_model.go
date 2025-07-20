package domain

import "time"

// UserStorage is the internal model used for JSON file storage
// It includes the hashed password field for persistence
type UserStorage struct {
	ID             int       `json:"id"`
	Username       string    `json:"username"`
	HashedPassword string    `json:"hashed_password"`
	CreatedAt      time.Time `json:"created_at"`
}

// ToUser converts UserStorage to User (for API responses)
func (us *UserStorage) ToUser() *User {
	return &User{
		ID:             us.ID,
		Username:       us.Username,
		HashedPassword: us.HashedPassword,
		CreatedAt:      us.CreatedAt,
	}
}

// ToStorage converts User to UserStorage with the provided hashed password
func (u *User) ToStorage(hashedPassword string) *UserStorage {
	return &UserStorage{
		ID:             u.ID,
		Username:       u.Username,
		HashedPassword: hashedPassword,
		CreatedAt:      u.CreatedAt,
	}
}

// CreateUserStorageFromCredentials creates a UserStorage instance from registration data
func CreateUserStorageFromCredentials(id int, username, hashedPassword string, createdAt time.Time) *UserStorage {
	return &UserStorage{
		ID:             id,
		Username:       username,
		HashedPassword: hashedPassword,
		CreatedAt:      createdAt,
	}
}
