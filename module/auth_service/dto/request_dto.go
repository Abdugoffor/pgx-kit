package auth_dto

type Register struct {
	FullName string `json:"full_name" validate:"required,min=2,max=255"`
	Phone    string `json:"phone"     validate:"required,min=7,max=50"`
	Password string `json:"password"  validate:"required,min=6"`
}

type Login struct {
	Phone    string `json:"phone"    validate:"required"`
	Password string `json:"password" validate:"required"`
}
