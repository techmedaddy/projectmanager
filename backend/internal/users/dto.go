package users

import "time"

// Response is the public JSON representation of a user.
type Response struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// AssigneesResponse is the response body for assignee-options endpoints.
type AssigneesResponse struct {
	Users []Response `json:"users"`
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

// ToResponses converts repository user models into API DTOs.
func ToResponses(users []User) []Response {
	items := make([]Response, 0, len(users))
	for _, user := range users {
		items = append(items, ToResponse(user))
	}

	return items
}
