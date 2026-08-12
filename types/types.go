package types

type DBSTATUS struct {
	Success bool
	Message string
}

type USER struct {
	Password string
	Name     string
	Email    string
	Role     string
}

type USERLOGIN struct {
	Email    string
	Password string
}

type USEROBJECT struct {
	Id       int
	Password string
	Name     string
	Email    string
	Role     string
}

type USERDATA struct {
	Name  string
	Email string
	Role  string
}

type REQUEST struct {
	Password string `json:"password"`
	OTP      int    `json:"otp"`
}

type AdminCreateRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}
