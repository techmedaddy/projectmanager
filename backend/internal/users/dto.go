package users

import "time"

// Response is the public JSON representation of a user.
type Response struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// ToResponse converts a repository user model into its API DTO.
func ToResponse(user User) Response {
	return Response{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}
}
