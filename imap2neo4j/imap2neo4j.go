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

	if len(os.Args) < 6 {
		log.Println("Usage : imap2neo4j imapServer imapUsername imapPassword imapMailbox neo4jServer [neo4jUsername (use - for no authorization)] [neo4jPassword (use - for no authorization)] [paging size (use * or omit for all at once)] [messages (first:last; omit for all; * for latest)]")
		log.Println("Eg:")
		log.Println("        imap2neo4j imap.googlemail.com user@gmail.com password Inbox localhost:7474")
		log.Println("        imap2neo4j imap.googlemail.com user@gmail.com password Inbox localhost:7473 neo4j password")
		log.Println("        imap2neo4j imap.googlemail.com user@gmail.com password Inbox localhost:7474 - - 10 0:100")
		log.Println("        imap2neo4j imap.googlemail.com user@gmail.com password Inbox localhost:7473 neo4j password * 500:*")
		os.Exit(1)
	}

	imapServer := os.Args[1]
	imapUsername := os.Args[2]
	imapPassword := os.Args[3]
	imapMailbox := os.Args[4]
	neo4jServer := os.Args[5]

	var neo4jUsername, neo4jPassword string

	var pagingSize int = -1
	var firstMessage int = 1
	var lastMessage int = -1

	if len(os.Args) >= 8 {
		if os.Args[6] != "-" {
			neo4jUsername = os.Args[6]
		}

		if os.Args[7] != "-" {
			neo4jPassword = os.Args[7]
		}

		if len(os.Args) >= 9 {
			pagingSize, err = strconv.Atoi(os.Args[8])
			if err != nil {
				log.Fatalln("Unable to parse paging size input argument:", err)
			}

			if len(os.Args) >= 10 {
				messages := strings.Split(os.Args[9], ":")
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
	}

	importedMessages, messagesToProcess, failedMessagesErrors, err := email2neo4j.NewImapImporter(imapServer, imapUsername, imapPassword, neo4jServer, neo4jUsername, neo4jPassword).Import(imapMailbox, pagingSize, firstMessage, lastMessage)
	if err != nil {
		log.Println("ImportMailbox - FATAL ERROR:\n", err)
	}
	log.Println("Imported", strconv.Itoa(importedMessages), "of", strconv.Itoa(messagesToProcess), "messages.")
	log.Println(strconv.Itoa(len(failedMessagesErrors)), "failed emails:")
	for _, failedMessage := range failedMessagesErrors {
		log.Println("	", failedMessage)
	}
}
