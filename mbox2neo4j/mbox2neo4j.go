package main

import (
	"github.com/tolomaus/email2neo4j"
	"log"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 3 {
		log.Println("Usage : mbox2neo4j filename neo4jServer [neo4jUsername] [neo4jPassword]")
		log.Println("Eg:")
		log.Println("        mbox2neo4j /path/to/mbox/file localhost:7474")
		log.Println("        mbox2neo4j /path/to/mbox/file localhost:7473 neo4j password")
		os.Exit(1)
	}

	filename := os.Args[1]
	neo4jServer := os.Args[2]

	var neo4jUsername, neo4jPassword string

	if len(os.Args) >= 5 {
		if os.Args[3] != "-" {
			neo4jUsername = os.Args[3]
		}

		if os.Args[4] != "-" {
			neo4jPassword = os.Args[4]
		}
	}

	importedMessages, messagesToProcess, failedMessagesErrors, err := email2neo4j.ImportMboxFile(filename, neo4jServer, neo4jUsername, neo4jPassword)
	if err != nil {
		log.Println("ImportMailbox - FATAL ERROR:\n", err)
	}
	log.Println("Imported", strconv.Itoa(importedMessages), "of", strconv.Itoa(messagesToProcess), "messages.")
	log.Println(strconv.Itoa(len(failedMessagesErrors)), "failed emails:")
	for _, failedMessage := range failedMessagesErrors {
		log.Println("	", failedMessage)
	}
}
