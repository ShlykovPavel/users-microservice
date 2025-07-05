package get_user

type AuthUser struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=3,max=64"`
	Phone     string `json:"phone" validate:"required,numeric"`
	LastName  string `json:"last_name"`
	FirstName string `json:"first_name"`
}
