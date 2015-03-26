package email2neo4j

import (
	"github.com/jmcvetta/neoism"
	"log"
	"net/mail"
)

func InitializeNeo4jDatabase() (db *neoism.Database, err error) {
	log.Println("Connecting to neo4j...")
	db, err = neoism.Connect("http://localhost:7474/db/data")
	if err != nil {
		return nil, err
	}

	log.Println("Creating constraints and indexes if not yet existing...")
	indexesAccounts, err := db.Indexes("Account")
	if err != nil {
		return nil, err
	}

	if len(indexesAccounts) == 0 {
		log.Println("Creating the UNIQUE constraint for the Accounts in neo4j...")
		cqAccount := neoism.CypherQuery{
			Statement: `CREATE CONSTRAINT ON (account:Account) ASSERT account.email_address IS UNIQUE`,
		}

		err = db.Cypher(&cqAccount)
		if err != nil {
			log.Println(err)
		}

		/*
			log.Println("Creating the index for the Accounts in neo4j...")
			_, err = db.CreateIndex("Account", "email_address")
			if err != nil {
				log.Println(err)
			}
		*/
	}

	indexesEmails, err := db.Indexes("Email")
	if err != nil {
		return nil, err
	}

	if len(indexesEmails) == 0 {
		log.Println("Creating the UNIQUE constraint for the Email in neo4j...")
		cqAccount := neoism.CypherQuery{
			Statement: `CREATE CONSTRAINT ON (email:Email) ASSERT email.message_id IS UNIQUE`,
		}

		err = db.Cypher(&cqAccount)
		if err != nil {
			log.Println(err)
		}

		/*
			log.Println("Creating the index for the Emails in neo4j...")
			_, err = db.CreateIndex("Email", "message_id")
			if err != nil {
				log.Println(err)
			}
		*/
	}

	return db, nil
}

func GetOrCreateAccount(db *neoism.Database, address *mail.Address) (account *neoism.Node, created bool, err error) {
	account, created, err = db.GetOrCreateNode("Account", "email_address", neoism.Props{
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

func GetOrCreateEmail(db *neoism.Database, msg *mail.Message) (email *neoism.Node, created bool, err error) {
	email, created, err = db.GetOrCreateNode("Email", "message_id", BuildPropertiesForEmail(msg))

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

func GetOrCreateEmailPlaceHolder(db *neoism.Database, message_id string) (email *neoism.Node, created bool, err error) {
	email, created, err = db.GetOrCreateNode("Email", "message_id", neoism.Props{
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
