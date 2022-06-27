package core

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/rds/auth"
	_ "github.com/lib/pq"
	"log"
)

// ConnectRDS returns a DB instance.
// The Lambda function leverages IAM roles to gain access to the DB Proxy.
// The function does NOT set the search_path to the organization schema.
func ConnectRDS() (*sql.DB, error) {
	var dbName string = "pennsieve_postgres"
	var dbUser string = "dev_rds_proxy_user"

	// TODO: Replace postgres location dynamically
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

	return db, err
}

// ConnectRDSWithOrg returns a DB instance.
// The Lambda function leverages IAM roles to gain access to the DB Proxy.
// The function DOES set the search_path to the organization schema.
func ConnectRDSWithOrg(orgId int) (*sql.DB, error) {
	db, err := ConnectRDS()

	// Set Search Path to organization
	_, err = db.Exec(fmt.Sprintf("SET search_path = \"%d\";", orgId))
	if err != nil {
		log.Println(fmt.Sprintf("Unable to set search_path to %d.", orgId))
		err := db.Close()
		if err != nil {
			return nil, err
		}
		return nil, err
	}

	return db, err
}
