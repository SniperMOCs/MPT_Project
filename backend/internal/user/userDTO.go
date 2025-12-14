package user

type UserDTO struct {
	Username    *string `json:"username"`
	FullName    *string `json:"full_name"`
	AvatarURL   *string `json:"avatar_url"`
	Description *string `json:"description"`
}
