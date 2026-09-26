package staff

import "errors"

var (
	ErrInactive       = errors.New("this staff login has been disabled by the shop admin")
	ErrLoginDisabled  = errors.New("contact admin for login")
	ErrBadCredentials = errors.New("mobile number or PIN is incorrect")
)
