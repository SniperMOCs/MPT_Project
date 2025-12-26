package follow

// FollowInfoDTO представляет информацию о подписках пользователя
// @Description Количество подписчиков и подписок пользователя
type FollowInfoDTO struct {
	FollowersCount int `json:"followers_count" example:"100"` // Количество подписчиков
	FollowingCount int `json:"following_count" example:"50"`  // Количество подписок
}
