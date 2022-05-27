package config

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/rds/auth"
	"log"

	_ "github.com/lib/pq"
)

//var DB *sql.DB

// ConnectRDS returns a DB instance.
// The Lambda function leverages IAM roles to gain access to the DB Proxy.
// The function does NOT set the search_path to the organization schema as multiple
// concurrent upload session can be handled across multiple organizations.
func ConnectRDS(organizationId int) (*sql.DB, error) {

	var dbName string = "pennsieve_postgres"
	var dbUser string = "dev_rds_proxy_user"
	var dbHost string = "dev-pennsieve-postgres-use1-proxy.proxy-ctkakwd4msv8.us-east-1.rds.amazonaws.com"
	var dbPort int = 5432
	var dbEndpoint string = fmt.Sprintf("%s:%d", dbHost, dbPort)
	var region string = "us-east-1"

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("configuration error: " + err.Error())
	}

	authenticationToken, err := auth.BuildAuthToken(
		context.TODO(), dbEndpoint, region, dbUser, cfg.Credentials)
	if err != nil {
		panic("failed to create authentication token: " + err.Error())
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s",
		dbHost, dbPort, dbUser, authenticationToken, dbName,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	// Set Search Path to organization
	_, err = db.Exec(fmt.Sprintf("SET search_path = \"%d\";", organizationId))
	if err != nil {
		log.Println(fmt.Sprintf("Unable to set search_path to %d.", organizationId))
		err := db.Close()
		if err != nil {
			return nil, err
		}
		return nil, err
	}

	return db, err
}
