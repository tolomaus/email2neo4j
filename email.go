package email2neo4j

import (
	"errors"
	"fmt"
	"github.com/jmcvetta/neoism"
	"log"
	"net/mail"
	"strconv"
	"strings"
)

func ProcessMessage(msg *mail.Message, edb *EmailDatabase) (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		}
	}()

	if msg == nil {
		return errors.New("msg is nil")
	}

	//body := imap.Asstrconv.Itoa(rsp.MessageInfo().Attrs["RFC822.TEXT"])

	date, err := msg.Header.Date()
	if err != nil {
		return err
	}

	from, err := mail.ParseAddress(msg.Header.Get("From"))
	if err != nil {
		return err
	}
	from_label := fmt.Sprint(from.Name, " (", from.Address, ")")

	to, _ := msg.Header.AddressList("To")
	var to_label string
	for _, to_address := range to {
		to_label += fmt.Sprint(to_address.Name, " (", to_address.Address, "), ")
	}

	cc, _ := msg.Header.AddressList("Cc")
	var cc_label string
	for _, cc_address := range cc {
		cc_label += fmt.Sprint(cc_address.Name, " (", cc_address.Address, "), ")
	}

	log.Println()
	log.Println("============================================")
	log.Println("Message-ID:", msg.Header.Get("Message-ID"))
	log.Println("Date:", date.Format("2006-01-02 15:04:05 -01:00"))
	log.Println("From:", from_label)
	log.Println("To:", to_label)
	log.Println("Cc:", cc_label)
	log.Println("Subject:", msg.Header.Get("Subject"))
	log.Println("In-Reply-To:", msg.Header.Get("In-Reply-To"))
	log.Println("References:", msg.Header.Get("References"))
	log.Println("")
	//log.Println("Body:")
	//log.Println(body)

	log.Println("Creating node for 'from' account...")
	from_account, created, err := edb.GetOrCreateAccount(from)
	if err != nil {
		return err
	}
	if !created {
		log.Println("The node already exists.")
	}

	log.Println("Creating node for", strconv.Itoa(len(to)), "'to' account(s)...")
	to_accounts := []*neoism.Node{}
	for _, address := range to {
		to_account, created, err := edb.GetOrCreateAccount(address)
		if err != nil {
			return err
		}
		if !created {
			log.Println("The node already exists.")
		}
		to_accounts = append(to_accounts, to_account)
	}

	cc_accounts := []*neoism.Node{}
	if len(cc) > 0 {
		log.Println("Creating node for", strconv.Itoa(len(cc)), "'cc' account(s)...")
		for _, address := range cc {
			cc_account, created, err := edb.GetOrCreateAccount(address)
			if err != nil {
				return err
			}
			if !created {
				log.Println("The node already exists.")
			}
			cc_accounts = append(cc_accounts, cc_account)
		}
	}

	log.Println("Creating node for email...")
	email, created, err := edb.GetOrCreateEmail(msg)
	if err != nil {
		return err
	}

	if !created {
		log.Println("The email node already exists.")

		log.Println("Verifying if the email node is a placeholder...")
		placeholder, err := email.Property("placeholder")
		if err != nil && !strings.EqualFold(err.Error(), "Cannot find in database.") {
			return err
		}

		if placeholder == "true" {
			log.Println("The email node is a placeholder, setting the fully properties now...")
			email.SetProperties(BuildPropertiesForEmail(msg))
			created = true
		}
	}

	if created {
		log.Println("Creating relationship from -SENT-> email...")
		from_account.Relate("SENT", email.Id(), nil)

		if len(to_accounts) > 0 {
			log.Println("Creating relationship(s) email -TO-> to...")
			for _, to_account := range to_accounts {
				email.Relate("TO", to_account.Id(), nil)
			}
		}

		if len(cc_accounts) > 0 {
			log.Println("Creating relationship(s) email -CC-> cc...")
			for _, cc_account := range cc_accounts {
				email.Relate("CC", cc_account.Id(), nil)
			}
		}

		if msg.Header.Get("In-Reply-To") != "" {
			log.Println("This email is a reply.")

			in_reply_tos := strings.Fields(msg.Header.Get("In-Reply-To"))

			for _, in_reply_to := range in_reply_tos {
				log.Println("Creating a placeholder for the original email if it doesn't exist yet...")
				orig_email, _, err := edb.GetOrCreateEmailPlaceHolder(removeSpecialChars(in_reply_to))
				if err != nil {
					return err
				}

				log.Println("Creating relationship email -REPLY-> original email...")
				email.Relate("REPLY", orig_email.Id(), nil)
			}

		} else if msg.Header.Get("References") != "" {
			log.Println("This email has references to other emails.")

			references := strings.Fields(msg.Header.Get("References"))

			for _, reference := range references {
				log.Println("Creating a placeholder for the original email if it doesn't exist yet...")
				orig_email, _, err := edb.GetOrCreateEmailPlaceHolder(removeSpecialChars(reference))
				if err != nil {
					return err
				}

				log.Println("Creating relationship email -REFERENCE-> original email...")
				email.Relate("REFERENCE", orig_email.Id(), nil)
			}
		}
	} else {
		log.Println("Email already existed as a node, so not adding the relationships.")
	}

	log.Println("============================================")
	log.Println()

	return nil
}

func BuildPropertiesForEmail(msg *mail.Message) (p neoism.Props) {
	date, _ := msg.Header.Date()
	return neoism.Props{
		"message_id": removeSpecialChars(msg.Header.Get("Message-ID")),
		"subject":    msg.Header.Get("Subject"),
		"date":       date}
}

func removeSpecialChars(string string) string {
	string = strings.Replace(string, "<", "", -1)
	string = strings.Replace(string, ">", "", -1)
	string = strings.Replace(string, "\"", "", -1)
	return string
}
