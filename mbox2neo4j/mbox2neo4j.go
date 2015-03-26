package main

import (
	"github.com/tolomaus/email2neo4j"
	"log"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) != 2 {
		log.Println("Usage : mbox2neo4j filename")
		os.Exit(1)
	}

	filename := os.Args[1]

	importedMessages, messagesToProcess, failedMessagesErrors, err := email2neo4j.ImportMboxFile(filename)
	if err != nil {
		log.Println("ImportMailbox - FATAL ERROR:\n", err)
	}
	log.Println("Imported", strconv.Itoa(importedMessages), "of", strconv.Itoa(messagesToProcess), "messages.")
	log.Println(strconv.Itoa(len(failedMessagesErrors)), "failed emails:")
	for _, failedMessage := range failedMessagesErrors {
		log.Println("	", failedMessage)
	}
}
