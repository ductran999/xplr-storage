package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Account struct {
	ID      int `gorm:"primaryKey"`
	Name    string
	Balance int
}

func withdraw(db *gorm.DB, workerName string, accountID int, amount int, wg *sync.WaitGroup) {
	defer wg.Done()

	log.Printf("[%s] Start transaction, try to lock...", workerName)
	err := db.Transaction(func(tx *gorm.DB) error {
		var account Account

		err := tx.Clauses(clause.Locking{
			Strength: "UPDATE",
			Options:  "NOWAIT",
		}).First(&account, accountID).Error

		if err != nil {
			log.Printf("[%s]: row is locked! err: %v", workerName, err)
			return err
		}

		log.Printf("[%s] Locked! Current Balance: %d. Processing...", workerName, account.Balance)
		time.Sleep(3 * time.Second)

		if account.Balance < amount {
			return fmt.Errorf("Not enough balance")
		}
		account.Balance -= amount

		return tx.Save(&account).Error
	})

	if err == nil {
		log.Printf("[%s] Transaction Completed!", workerName)
	}
}

func main() {
	dsn := "host=localhost user=test password=test dbname=dbtest port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("connect db failed:", err)
	}

	db.Migrator().DropTable(&Account{})
	db.AutoMigrate(&Account{})
	db.Create(&Account{ID: 1, Name: "Alice", Balance: 1000})

	var wg sync.WaitGroup
	wg.Add(2)

	go withdraw(db, "WORKER_1", 1, 300, &wg)

	time.Sleep(500 * time.Millisecond)

	go withdraw(db, "WORKER_2", 1, 400, &wg)

	wg.Wait()

	var finalAcc Account
	db.First(&finalAcc, 1)
	fmt.Printf("balance: %d VNĐ\n", finalAcc.Balance)
}
