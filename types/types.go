package types

type DBSTATUS struct {
	Success bool
	Message string
}

type USER struct {
	Password string
	Name     string
	Email    string
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
}

type USERDATA struct {
	Name  string
	Email string
}

type REQUEST struct {
	Password string `json:"password"`
	OTP      int `json:"otp"`
}
