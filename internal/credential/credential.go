package credential

import "errors"

var (
	ErrNoExist      = errors.New("user doesn`t exist")
	ErrNoAccess     = errors.New("no access to credentials")
	ErrAlreadyExist = errors.New("the data already exists for server")
)

type Credential struct {
	UserID   int64
	Login    string
	Password string
}

type CredentialRepo interface {
	Add(userID int64, serverName string, login string, password string) error
	GetByServerName(userID int64, serverName string) (Credential, error)
	Delete(userID int64, serverName string) (Credential, error)
}

type PasswordCoder interface {
	Encrypt(password string) (string, error)
	Decrypt(passwordCrypt string) (string, error)
}
