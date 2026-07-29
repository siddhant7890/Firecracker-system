package staff

import "errors"

var (
	ErrInactive       = errors.New("this staff login has been disabled by the shop admin")
	ErrBadCredentials = errors.New("mobile number or PIN is incorrect")
)
