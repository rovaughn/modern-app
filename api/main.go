package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds/rdsutils"
	"github.com/go-sql-driver/mysql"
	"github.com/neelance/graphql-go"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var sess = session.Must(session.NewSession())

var db = func() *sql.DB {
	log.Printf("Loading database")
	dbConfig := mysql.NewConfig()

	switch os.Getenv("environment") {
	case "local":
		dbConfig.Addr = "localhost"
		dbConfig.User = os.Getenv("db_user")
		dbConfig.Passwd = os.Getenv("db_password")
		dbConfig.DBName = os.Getenv("db_name")
	case "lambda":
		rootCertPool := x509.NewCertPool()
		pem, err := ioutil.ReadFile("rds-ca.pem")
		if err != nil {
			panic(err)
		}

		if !rootCertPool.AppendCertsFromPEM(pem) {
			panic("Failed to append PEM")
		}

		mysql.RegisterTLSConfig("custom", &tls.Config{
			RootCAs: rootCertPool,
		})
		dbConfig.TLSConfig = "custom"

		if os.Getenv("db_endpoint") == "" || os.Getenv("db_region") == "" || os.Getenv("db_name") == "" || os.Getenv("db_user") == "" {
			panic("Environment variables required: db_endpoint, db_region, db_name, db_user")
		}

		dbConfig.Addr = os.Getenv("db_endpoint")
		if dbConfig.Addr == "" {
			panic("No db_endpoint provided")
		}

		authToken, err := rdsutils.BuildAuthToken(dbConfig.Addr+":3306", os.Getenv("db_region"), os.Getenv("db_user"), credentials.NewEnvCredentials())
		if err != nil {
			panic(err)
		}
		dbConfig.Passwd = authToken

		dbConfig.Net = "tcp"
		dbConfig.DBName = os.Getenv("db_name")
		dbConfig.User = os.Getenv("db_user")
		dbConfig.AllowCleartextPasswords = true
	default:
		panic("Unknown environment")
	}

	db, err := sql.Open("mysql", dbConfig.FormatDSN())
	if err != nil {
		panic(err)
	}

	return db
}()

var schema = func() *graphql.Schema {
	schema, err := ioutil.ReadFile("schema.graphql")
	if err != nil {
		panic(err)
	}

	return graphql.MustParseSchema(string(schema), new(resolver))
}()

func handler(req events.APIGatewayProxyRequest) (res *events.APIGatewayProxyResponse, err error) {
	log.Printf("handler here")
	defer func() {
		log.Printf("handler exiting %#v %#v", res, err)
	}()
	// TODO is handling panics ourself necessary?
	//defer func() {
	//	if r := recover(); r != nil {
	//		log.Printf("panic: %s: %s", r, debug.Stack())
	//		res = &events.APIGatewayProxyResponse{
	//			StatusCode: 500,
	//			Body:       "internal server error",
	//		}
	//	}
	//}()

	if req.HTTPMethod == "OPTIONS" {
		return &events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Headers: map[string]string{
				"Access-Control-Allow-Headers": "content-type",
				"Access-Control-Allow-Origin":  "*",
			},
		}, nil
	}

	var query struct {
		Query         string                 `json:"query"`
		OperationName string                 `json:"operationName"`
		Variables     map[string]interface{} `json:"variables"`
	}

	if err := json.Unmarshal([]byte(req.Body), &query); err != nil {
		return &events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Headers: map[string]string{
				"Access-Control-Allow-Headers": "content-type",
				"Access-Control-Allow-Origin":  "*",
			},
			Body: fmt.Sprintf("Failed to parse request body as JSON: %s", err),
		}, nil
	}

	log.Printf("Running query %q", query.Query)

	result := schema.Exec(context.TODO(), query.Query, query.OperationName, query.Variables)

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	log.Printf("Returning JSON %q", resultJSON)

	return &events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Headers": "content-type",
			"Content-Type":                 "application/json",
		},
		Body: string(resultJSON),
	}, nil
}

func main() {
	defer db.Close()

	log.Printf("Starting lambda function up")

	switch os.Getenv("environment") {
	case "local":
		var addr string
		flag.StringVar(&addr, "addr", ":80", "Address to listen on")
		flag.Parse()

		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			var req events.APIGatewayProxyRequest
			req.HTTPMethod = r.Method
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				panic(err)
			}
			req.Body = string(body)

			res, err := handler(req)
			if err != nil {
				panic(err)
			}

			for key, value := range res.Headers {
				w.Header().Add(key, value)
			}

			w.WriteHeader(res.StatusCode)

			if _, err := w.Write([]byte(res.Body)); err != nil {
				panic(err)
			}
		})

		log.Println("Serving graphql on", addr)
		log.Fatal(http.ListenAndServe(addr, nil))
	case "lambda":
		lambda.Start(handler)
	default:
		panic("Unknown environment")
	}
}
