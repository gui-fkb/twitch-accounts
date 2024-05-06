package main

import (
	"fmt"

	"github.com/ox-y/GoGmailnator"
)

func fastEmailTest() {
	var sess GoGmailnator.Session

	err := sess.Init(nil)
	if err != nil {
		panic(err)
	}

	isAlive, err := sess.IsAlive()
	if err != nil {
		panic(err)
	}

	if isAlive {
		fmt.Println("Session is alive.")
	} else {
		fmt.Println("Session is dead.")
		return
	}

	emailAddress, err := sess.GenerateEmailAddress()
	if err != nil {
		panic(err)
	}

	fmt.Println("Email address is " + emailAddress + ".")

	emails, err := sess.RetrieveMail(emailAddress)
	if err != nil {
		panic(err)
	}

	for _, email := range emails {
		fmt.Printf("From: %s, Subject: %s, Time: %s\n", email.From, email.Subject, email.Time)
	}
}
