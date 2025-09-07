package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/viper"
	"github.com/jordanknott/taskcafe/internal/config"
)

func main() {
	// Set environment variables the same way as CI
	os.Setenv("TASKCAFE_DATABASE_HOST", "localhost")
	os.Setenv("TASKCAFE_DATABASE_USER", "test_user")
	os.Setenv("TASKCAFE_DATABASE_PASSWORD", "test_password")
	os.Setenv("TASKCAFE_DATABASE_NAME", "taskcafe_test")
	os.Setenv("TASKCAFE_DATABASE_PORT", "5432")
	os.Setenv("TASKCAFE_DATABASE_SSLMODE", "disable")

	// Initialize config exactly the same way as the application
	viper.SetEnvPrefix("TASKCAFE")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	config.InitDefaults()

	// Get database config
	dbConfig := config.GetDatabaseConfig()
	
	fmt.Printf("Database Configuration:\n")
	fmt.Printf("Host: %s\n", dbConfig.Host)
	fmt.Printf("User: %s\n", dbConfig.Username)
	fmt.Printf("Password: %s\n", dbConfig.Password)
	fmt.Printf("Name: %s\n", dbConfig.Name)
	fmt.Printf("Port: %s\n", dbConfig.Port)
	fmt.Printf("SSL Mode: %s\n", dbConfig.SslMode)
	
	connectionString := dbConfig.GetDatabaseConnectionUri()
	fmt.Printf("\nConnection String: %s\n", connectionString)
	
	if dbConfig.Username == "test_user" {
		fmt.Println("\n✅ SUCCESS: Configuration correctly uses 'test_user' instead of 'root'")
	} else {
		log.Fatalf("❌ FAILURE: Configuration uses '%s' instead of 'test_user'", dbConfig.Username)
	}
}
