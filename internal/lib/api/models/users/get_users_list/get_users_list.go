package get_users_list

type UserInfoList struct {
	Id        int64  `json:"id"`
	Email     string `json:"email" validate:"required,email"`
	Phone     string `json:"phone" validate:"required,numeric"`
	LastName  string `json:"last_name" validate:"required"`
	FirstName string `json:"first_name" validate:"required"`
	Role      string `json:"role" validate:"required"`
}
