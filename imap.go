package email2neo4j

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/mxk/go-imap/imap"
	"log"
	"net/mail"
	"strconv"
	"strings"
	"time"
)

func NewImapImporter(imapServer, imapUsername, imapPassword, neo4jServer, neo4jUsername, neo4jPassword string) *ImapMailboxImporter {
	return &ImapMailboxImporter{
		imapServer:    imapServer,
		imapUsername:  imapUsername,
		imapPassword:  imapPassword,
		neo4jServer:   neo4jServer,
		neo4jUsername: neo4jUsername,
		neo4jPassword: neo4jPassword,
	}
}

type ImapMailboxImporter struct {
	imapServer, imapUsername, imapPassword, neo4jServer, neo4jUsername, neo4jPassword string
}

func (imi *ImapMailboxImporter) Import(imapMailbox string, pagingSize, firstMessage, lastMessage int) (importedMessages int, messagesToProcess int, failedMessagesErrors []string, err error) {
	log.Println("Starting...")

	var response *imap.Response
	var cmd *imap.Command

	log.Println("Opening IMAP mailbox " + imapMailbox + " on server " + imi.imapServer + " with user " + imi.imapUsername + "...")
	client, err := imi.openImapMailbox(imapMailbox)
	if err != nil {
		return 0, 0, nil, err
	}
	defer client.Logout(30 * time.Second)

	log.Println("Initializing neo4j...")
	edb, err := NewEmailDatabase(imi.neo4jServer, imi.neo4jUsername, imi.neo4jPassword)
	if err != nil {
		return 0, 0, nil, err
	}

	if lastMessage == -1 {
		lastMessage = int(client.Mailbox.Messages)
	}

	messagesToProcess = lastMessage - firstMessage + 1

	if pagingSize == -1 {
		pagingSize = messagesToProcess //all at once
	}

	var totalPages = messagesToProcess / pagingSize //rounded to the lowest int

	if messagesToProcess%pagingSize > 0 { //if there is a remainder then add a page for it
		totalPages++
	}

	var currentMessage int = 0

	for page := 1; page <= totalPages; page++ {
		firstMessageOfPage := firstMessage + ((page - 1) * pagingSize)
		var lastMessageOfPage int
		if firstMessage+(page*pagingSize)-1 < lastMessage {
			lastMessageOfPage = firstMessage + (page * pagingSize) - 1
		} else {
			lastMessageOfPage = lastMessage
		}

		log.Println("Processing page", strconv.Itoa(page), "of", strconv.Itoa(totalPages), "(messages", strconv.Itoa(firstMessageOfPage), "till", strconv.Itoa(lastMessageOfPage), ")...")

		set, _ := imap.NewSeqSet("")
		set.AddRange(uint32(firstMessageOfPage), uint32(lastMessageOfPage))

		log.Println("Fetching messages from mailbox...")
		cmd, err = client.Fetch(set, "RFC822.HEADER") //, "RFC822.TEXT")
		if err != nil {
			return currentMessage, messagesToProcess, nil, err
		}

		for cmd.InProgress() {
			err = client.Recv(30 * time.Second)
			if err != nil {
				return 0, messagesToProcess, nil, err
			}

			for _, response = range cmd.Data {
				currentMessage++
				log.Println("Processing message", strconv.Itoa(currentMessage), "of", strconv.Itoa(messagesToProcess), "...")

				err = imi.processResponse(response, edb)
				if err != nil {
					failedMessagesErrors = append(failedMessagesErrors, "Not able to process message "+strconv.Itoa(currentMessage)+": "+err.Error())
					log.Println()
				}
			}
			cmd.Data = nil
		}

		if page < totalPages {
			if imi.askToContinue() == false {
				return currentMessage, messagesToProcess, failedMessagesErrors, nil
			}
		}
	}

	// Process unilateral server data
	for _, response = range client.Data {
		log.Println("Server data:", response)
	}
	client.Data = nil

	return currentMessage, messagesToProcess, failedMessagesErrors, nil
}

func (imi *ImapMailboxImporter) openImapMailbox(mailbox string) (client *imap.Client, err error) {
	var cmd *imap.Command

	conf := &tls.Config{
		Rand: rand.Reader,
	}

	log.Println("Connecting to the server...")
	client, err = imap.DialTLS(imi.imapServer, conf)
	if err != nil {
		return nil, err
	}

	log.Println("Server says hello:", client.Data[0].Info)
	client.Data = nil

	if client.Caps["STARTTLS"] {
		log.Println("Starting TLS...")
		client.StartTLS(nil)
	}

	if client.State() == imap.Login {
		log.Println("Logging in...")
		client.Login(imi.imapUsername, imi.imapPassword)
	}

	if client.State() != imap.Auth {
		return nil, errors.New("Cannot authenticate " + imi.imapUsername)
	}

	log.Println("Retrieving top mailbox:")
	cmd, err = imap.Wait(client.List("", "%"))
	if err != nil {
		return nil, err
	}

	top_mailbox := cmd.Data[0].MailboxInfo().Name
	delim := cmd.Data[0].MailboxInfo().Delim
	log.Println("Found", top_mailbox)

	log.Println("Listing sub mailboxes of", top_mailbox, ":")
	cmd, err = imap.Wait(client.List("", cmd.Data[0].MailboxInfo().Name+cmd.Data[0].MailboxInfo().Delim+"%"))
	if err != nil {
		return nil, err
	}

	for _, response := range cmd.Data {
		log.Println("|--", response.MailboxInfo().Name)
	}

	var mailbox_path string
	if strings.EqualFold(top_mailbox, mailbox) {
		mailbox_path = top_mailbox
	} else {
		mailbox_path = top_mailbox + delim + mailbox
	}

	log.Println("Opening mailbox", mailbox_path, "...")
	_, err = client.Select(mailbox_path, true)
	if err != nil {
		return nil, err
	}
	log.Println("Total messages:", client.Mailbox.Messages)
	log.Println("Recent messages:", client.Mailbox.Recent)
	log.Println("Unseen messages:", client.Mailbox.Unseen)

	return client, nil
}

func (imi *ImapMailboxImporter) processResponse(response *imap.Response, edb *EmailDatabase) (err error) {
	header := imap.AsBytes(response.MessageInfo().Attrs["RFC822.HEADER"])

	msg, err := mail.ReadMessage(bytes.NewReader(header))
	if err != nil {
		return err
	}

	return ProcessMessage(msg, edb)
}

func (imi *ImapMailboxImporter) askToContinue() bool {
	fmt.Println("Continue? [yn]")

	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		log.Fatal(err)
	}

	if response == "y" {
		return true
	} else if response == "n" {
		return false
	} else {
		fmt.Println("Only y or n are valid responses")
		return imi.askToContinue()
	}
}
