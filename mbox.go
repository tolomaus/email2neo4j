package email2neo4j

import (
	"github.com/bthomson/mbox"
	"log"
	"strconv"
)

func ImportMboxFile(filename, neo4jServer, neo4jUsername, neo4jPassword string) (importedMessages int, messagesToProcess int, failedMessagesErrors []string, err error) {
	log.Println("Starting...")

	log.Println("Reading mbox file...")
	msgs, err := mbox.ReadFile(filename, true)

	log.Println("Initializing neo4j...")
	edb, err := NewEmailDatabase(neo4jServer, neo4jUsername, neo4jPassword)
	if err != nil {
		return 0, 0, nil, err
	}

	var currentMessage int = 0
	messagesToProcess = len(msgs)

	for _, msg := range msgs {
		currentMessage++
		log.Println("Processing message", strconv.Itoa(currentMessage), "of", strconv.Itoa(messagesToProcess), "...")

		err = ProcessMessage(msg, edb)
		if err != nil {
			failedMessagesErrors = append(failedMessagesErrors, "Not able to process message "+strconv.Itoa(currentMessage)+": "+err.Error())
			log.Println()
		}
	}

	return currentMessage, messagesToProcess, failedMessagesErrors, nil
}
