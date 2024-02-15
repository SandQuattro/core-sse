package services

import "sse-demo-core/internal/app/structs"

type UserService interface {
	FindUserById(id int) (user *structs.User, err error)
	FindUserBySub(sub string) (user *structs.User, err error)
}
