package main

import (
	"github.com/tolomaus/email2neo4j"
	"log"
	"os"
	"strconv"
	"strings"
)

func main() {
	var err error

	if len(os.Args) < 5 {
		log.Println("Usage : imap2neo4j server mailbox email passwd [paging size (use * or omit for all at once)] [messages (first:last; omit for all; * for latest)]")
		os.Exit(1)
	}

	server := os.Args[1]
	mailbox := os.Args[2]
	email := os.Args[3]
	passwd := os.Args[4]

	var pagingSize int = -1
	var firstMessage int = 1
	var lastMessage int = -1
	if len(os.Args) >= 6 {
		pagingSize, err = strconv.Atoi(os.Args[5])
		if err != nil {
			log.Fatalln("Unable to parse paging size input argument:", err)
		}

		if len(os.Args) >= 7 {
			messages := strings.Split(os.Args[6], ":")
			if len(messages) != 2 {
				log.Fatalln("Unable to parse messages input argument:", err)
			}

			firstMessage, err = strconv.Atoi(messages[0])
			if err != nil {
				log.Fatalln("Unable to parse messages input argument:", err)
			}

			if messages[1] != "*" {
				lastMessage, err = strconv.Atoi(messages[1])
				if err != nil {
					log.Fatalln("Unable to parse messages input argument:", err)
				}

				if firstMessage > lastMessage {
					log.Fatalln("Unable to parse messages input argument: first is bigger than last")
				}
			}
		}
	}

	importedMessages, messagesToProcess, failedMessagesErrors, err := email2neo4j.ImportImapMailbox(server, mailbox, email, passwd, pagingSize, firstMessage, lastMessage)
	if err != nil {
		log.Println("ImportMailbox - FATAL ERROR:\n", err)
	}
	log.Println("Imported", strconv.Itoa(importedMessages), "of", strconv.Itoa(messagesToProcess), "messages.")
	log.Println(strconv.Itoa(len(failedMessagesErrors)), "failed emails:")
	for _, failedMessage := range failedMessagesErrors {
		log.Println("	", failedMessage)
	}
}
