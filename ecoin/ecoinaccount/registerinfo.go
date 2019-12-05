package eaccount

// RegisterInfo 注册信息
type RegisterInfo struct {
	NameField string	`json:"name"`
	PhoneField string 	`json:"phone"`
	InstitutionField string 	`json:"institution"`
}
