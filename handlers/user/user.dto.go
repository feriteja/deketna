package user

type CreateUserRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required,min=6" example:"password123"`
}

type SignInRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user1@example.com"`
	Password string `json:"password" binding:"required,min=6" example:"password123"`
}

type SignInResponse struct {
	Token string `json:"token" example:"your_jwt_token"`
}

type UserResponse struct {
	ID        uint   `json:"id"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	DeletedAt string `json:"deleted_at"`
}

type ProfileResponse struct {
	ID        uint         `json:"id"`
	Address   string       `json:"address"`
	Name      string       `json:"name"`
	UserID    uint         `json:"user_id"`
	ImageURL  string       `json:"image_url"`
	CreatedAt string       `json:"created_at"`
	UpdatedAt string       `json:"updated_at"`
	User      UserResponse `json:"user,omitempty"`
}

type EditProfileRequest struct {
	Name     string `json:"name" binding:"omitempty,max=255"`
	Address  string `json:"address" binding:"omitempty,max=255"`
	ImageURL string `json:"image_url" binding:"omitempty,url"`
}

type EditProfileResponse struct {
	ID        uint   `json:"id"`
	Address   string `json:"address"`
	Name      string `json:"name"`
	UserID    uint   `json:"user_id"`
	ImageURL  string `json:"image_url"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
