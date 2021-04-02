/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"

	"github.com/spf13/viper"

	"github.com/mailbadger/app/entities"
	"github.com/mailbadger/app/storage"
	"github.com/mailbadger/app/utils"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "fixtures",
	Short: "fixtures is a cli for generating test data for mailbadger",
	Long: `Fixtures can generate testing data a user with a few campaigns alongside with a few templates. Also about 
hundreds of subscribers in a few segments`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

var (
	// db keeps the connection to the database
	db storage.Storage
)

func init() {
	initConfig()

	// Connecting to database
	driver := viper.GetString("DATABASE_DRIVER")
	conf := makeConfigFromEnv(driver)
	db = storage.From(openDbConn(driver, conf))
}

func initConfig() {
	viper.AddConfigPath(".")
	viper.SetConfigType("env")
	viper.SetConfigName("app")

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err.Error())
	}
}

func makeConfigFromEnv(driver string) string {
	switch driver {
	case "sqlite3":
		return viper.GetString("SQLITE3_FILE")
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true",
			viper.GetString("MYSQL_USER"),
			viper.GetString("MYSQL_PASS"),
			viper.GetString("MYSQL_HOST"),
			viper.GetString("MYSQL_PORT"),
			viper.GetString("MYSQL_DATABASE"),
		)
	default:
		return ""
	}
}

// openDbConn creates a database connection using the driver and config string
func openDbConn(driver, config string) *gorm.DB {
	db, err := gorm.Open(driver, config)
	if err != nil {
		log.WithError(err).Fatalln("db connection failed")
	}

	if driver == "mysql" {
		db.DB().SetMaxIdleConns(0)
	}

	fresh := false
	switch driver {
	case "sqlite3":
		if _, err := os.Stat(config); err != nil || config == ":memory:" {
			fresh = true
		}
	case "mysql":
		err := db.First(&entities.User{}).Error
		if err != nil {
			fresh = true
		}
	}

	if fresh {
		err = initDb(config, db)
		if err != nil {
			log.WithError(err).Fatalln("migrations failed")
		}
	}

	db.LogMode(utils.IsDebugMode())

	return db
}

// initDb seeds the database with the admin user, if the database has not been
// initialized before
func initDb(config string, db *gorm.DB) error {
	log.Info("Generating new credentials...")

	// Hashing the password with the default cost of 10
	secret, err := utils.GenerateRandomString(12)
	if err != nil {
		return fmt.Errorf("init db: gen rand string: %w", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("init db: hash password: %w", err)
	}

	uuid := uuid.New()

	// Create the default user
	nolimit := &entities.Boundaries{}
	err = db.Where("type = ?", "nolimit").First(nolimit).Error
	if err != nil {
		return fmt.Errorf("init db: fetch nolimit boundaries: %w", err)
	}

	admin := entities.User{
		Username: "admin",
		UUID:     uuid.String(),
		Password: sql.NullString{
			String: string(hashedPassword),
			Valid:  true,
		},
		Active:     true,
		Verified:   true,
		Boundaries: *nolimit,
		Source:     "mailbadger.io",
	}

	err = db.Save(&admin).Error
	if err != nil {
		return fmt.Errorf("init db: save user: %w", err)
	}

	log.WithFields(log.Fields{"user": "admin", "password": secret}).Info("Admin user credentials..make sure to change that password!")

	return nil
}