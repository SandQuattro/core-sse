package repository

import (
	"fmt"
	logdoc "github.com/LogDoc-org/logdoc-go-appender/logrus"
	"github.com/jmoiron/sqlx"
	"sse-demo-core/internal/app/structs"
	"sse-demo-core/internal/errs"
)

type UserRepository struct {
	DB *sqlx.DB
}

func New(db *sqlx.DB) *UserRepository {
	return &UserRepository{db}
}

func (r *UserRepository) FindUserByID(id int) (user *structs.User, err error) {
	defer func() {
		err = errs.WrapWithStackIfErr(">> FindUserById > Ошибка поиска пользователя по id", err)
	}()

	logger := logdoc.GetLogger()

	var u structs.User
	err = r.DB.Get(&u, `SELECT id
									FROM users u
								   WHERE u.id = $1`, id)
	if err != nil || u.ID == 0 {
		logger.Warn(fmt.Sprintf(">> FindUserById > Ошибка поиска пользователя по id: %d", id))
		// A return statement without arguments returns the named return values.
		// This is known as a "naked" return.
		return
	}

	user = &u
	// A return statement without arguments returns the named return values.
	// This is known as a "naked" return.
	return
}

func (r *UserRepository) FindUserBySub(sub string) (user *structs.User, err error) {
	defer func() {
		err = errs.WrapWithStackIfErr(">> FindUserBySub > Ошибка поиска пользователя по sub", err)
	}()

	logger := logdoc.GetLogger()

	var u structs.User
	err = r.DB.Get(&u, `SELECT id
									FROM users u
								   WHERE u.sub = $1`, sub)
	if err != nil || u.ID == 0 {
		logger.Warn(fmt.Sprintf(">> FindUserBySub > Ошибка поиска пользователя по sub: %s", sub))
		// A return statement without arguments returns the named return values.
		// This is known as a "naked" return.
		return
	}

	user = &u
	// A return statement without arguments returns the named return values.
	// This is known as a "naked" return.
	return
}
