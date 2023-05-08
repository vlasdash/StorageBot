package credential

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

type MySQLRepo struct {
	DB *sql.DB
}

var _ CredentialRepo = (*MySQLRepo)(nil)

func NewMySQLRepo(db *sql.DB) *MySQLRepo {
	return &MySQLRepo{
		DB: db,
	}
}

func (r *MySQLRepo) Add(userID int64, serverName string, login string, password string) error {
	row := r.DB.QueryRow(
		"SELECT user_id, login, password FROM credentials WHERE server_name = ? AND user_id = ?",
		serverName,
		userID,
	)
	credit := Credential{}
	err := row.Scan(&credit.UserID, &credit.Login, &credit.Password)
	if err != sql.ErrNoRows {
		return ErrAlreadyExist
	}

	_, err = r.DB.Exec(
		"INSERT INTO credentials (`user_id`, `server_name`, `login`, `password`) VALUES (?, ?, ?, ?)",
		userID,
		serverName,
		login,
		password,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *MySQLRepo) GetByServerName(userID int64, serverName string) (Credential, error) {
	row := r.DB.QueryRow(
		"SELECT user_id, login, password FROM credentials WHERE server_name = ?",
		serverName,
	)

	credit := Credential{}
	err := row.Scan(&credit.UserID, &credit.Login, &credit.Password)
	if err == sql.ErrNoRows {
		return credit, ErrNoExist
	}
	if err != nil {
		return credit, err
	}
	if credit.UserID != userID {
		return Credential{}, ErrNoAccess
	}

	return credit, nil

}

func (r *MySQLRepo) Delete(userID int64, serverName string) (Credential, error) {
	row := r.DB.QueryRow(
		"SELECT user_id, login, password FROM credentials WHERE server_name = ?",
		serverName,
	)

	credit := Credential{}
	err := row.Scan(&credit.UserID, &credit.Login, &credit.Password)
	if err == sql.ErrNoRows {
		return credit, ErrNoExist
	}
	if err != nil {
		return credit, err
	}
	if credit.UserID != userID {
		return Credential{}, ErrNoAccess
	}

	_, err = r.DB.Exec(
		"DELETE FROM credentials WHERE server_name = ?",
		serverName,
	)

	return credit, err
}
