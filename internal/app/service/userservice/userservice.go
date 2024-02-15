package userservice

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	urepository "sse-demo-core/internal/app/repository/users"
	"sse-demo-core/internal/app/structs"
)

type UserServiceImpl struct {
	users urepository.UserRepository
}

func New(db *sqlx.DB) *UserServiceImpl {
	urepo := urepository.New(db)
	return &UserServiceImpl{*urepo}
}

func (s *UserServiceImpl) FindUserById(id int) (user *structs.User, err error) {
	u, err := s.users.FindUserByID(id)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, fmt.Errorf("пользователь отсутсвует в БД")
	}
	return u, nil
}

func (s *UserServiceImpl) FindUserBySub(sub string) (user *structs.User, err error) {
	u, err := s.users.FindUserBySub(sub)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, fmt.Errorf("пользователь отсутсвует в БД")
	}
	return u, nil
}
