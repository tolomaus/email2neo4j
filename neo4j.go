package email2neo4j

import (
	"github.com/jmcvetta/neoism"
	"log"
	"net/mail"
)

func NewEmailDatabase(server, username, password string) (*EmailDatabase, error) {
	edb := &EmailDatabase{
		server:   server,
		username: username,
		password: password,
	}
	err := edb.initializeNeo4jDatabase()
	return edb, err
}

type EmailImporter interface {
	GetOrCreateAccount(address *mail.Address) (account *neoism.Node, created bool, err error)
	GetOrCreateEmail(msg *mail.Message) (email *neoism.Node, created bool, err error)
	GetOrCreateEmailPlaceHolder(message_id string) (email *neoism.Node, created bool, err error)
}

type EmailDatabase struct {
	*neoism.Database
	server, username, password string
}

func (edb *EmailDatabase) initializeNeo4jDatabase() (err error) {
	log.Println("Connecting to neo4j...")

	var url string
	if edb.username == "" {
		url = "http://" + edb.server + "/db/data"
	} else {
		url = "https://" + edb.username + ":" + edb.password + "@" + edb.server + "/user/" + edb.username
	}
	edb.Database, err = neoism.Connect(url)
	if err != nil {
		return err
	}

	log.Println("Creating constraints and indexes if not yet existing...")
	indexesAccounts, err := edb.Indexes("Account")
	if err != nil {
		return err
	}

	if len(indexesAccounts) == 0 {
		log.Println("Creating the UNIQUE constraint for the Accounts in neo4j...")
		cqAccount := neoism.CypherQuery{
			Statement: `CREATE CONSTRAINT ON (account:Account) ASSERT account.email_address IS UNIQUE`,
		}

		err = edb.Cypher(&cqAccount)
		if err != nil {
			log.Println(err)
		}
	}

	indexesEmails, err := edb.Indexes("Email")
	if err != nil {
		return err
	}

	if len(indexesEmails) == 0 {
		log.Println("Creating the UNIQUE constraint for the Email in neo4j...")
		cqAccount := neoism.CypherQuery{
			Statement: `CREATE CONSTRAINT ON (email:Email) ASSERT email.message_id IS UNIQUE`,
		}

		err = edb.Cypher(&cqAccount)
		if err != nil {
			log.Println(err)
		}
	}

	return nil
}

func (edb *EmailDatabase) GetOrCreateAccount(address *mail.Address) (account *neoism.Node, created bool, err error) {
	account, created, err = edb.GetOrCreateNode("Account", "email_address", neoism.Props{
		"name":          address.Name,
		"email_address": address.Address})

	if err != nil {
		return
	}

	if created {
		err = account.AddLabel("Account")
		if err != nil {
			return
		}
	}

	return
}

func (edb *EmailDatabase) GetOrCreateEmail(msg *mail.Message) (email *neoism.Node, created bool, err error) {
	email, created, err = edb.GetOrCreateNode("Email", "message_id", BuildPropertiesForEmail(msg))

	if err != nil {
		return
	}

	if created {
		err = email.AddLabel("Email")
		if err != nil {
			return
		}
	}

	return
}

func (edb *EmailDatabase) GetOrCreateEmailPlaceHolder(message_id string) (email *neoism.Node, created bool, err error) {
	email, created, err = edb.GetOrCreateNode("Email", "message_id", neoism.Props{
		"message_id":  message_id,
		"placeholder": "true"})

	if err != nil {
		return
	}

	if created {
		err = email.AddLabel("Email")
		if err != nil {
			return
		}
	}

	return
}
