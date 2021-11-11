package main

import (
	"context"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"restaurant/server"
)

const (
	host = "localhost"
	port = 5432
	user = "postgres"
	password = "1488qwerdf"
	dbname = "restaurant"
)

func main(){
	baseUrl:= fmt.Sprintf("host= %s port= %d user= %s password= %s dbname= %s sslmode=disable", host,port,user,password,dbname)

	srv:= server.NewServer(context.Background(), ":8080", baseUrl)

	er:=srv.Run()
	if er != nil{
		log.Println("Error while running the server", er)
	}
}
