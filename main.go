package main

import (
	"log"

	"vacancies/config"
	"vacancies/grpc"
	"vacancies/parse"
)

func main() {
	conf, err := config.NewConfig()
	if err != nil {
		log.Fatalln("config.NewConfig: ", err)
	}
	log.Printf("Establishing connection to: %s\n "+
		"Sending request to receive authorization data...\n", conf.AuthorizeTarget)
	authorizationData, err := grpc.GetAuthorizationData(conf.AuthorizeTarget)
	if err != nil {
		log.Fatalf("grpc.GetAuthorizationData: %v", err)
	}
	log.Println("Authorization data received.")

	err = parse.Do(conf, authorizationData)
	if err != nil {
		log.Fatalf("parse.Do: %v", err)
	}
	log.Println("----------------Done")
}
