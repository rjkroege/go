package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net/mail"
	"os"
	"strconv"
	"strings"
	"time"

	"9fans.net/go/acme"
	"github.com/tcnksm/go-gitconfig"
)

func main() {
	// We need an email address for the sender. Try $EMAIL, then git config.
	mailaddr := os.Getenv("EMAIL")
	if mailaddr == "" {
		var err error
		mailaddr, err = gitconfig.Email()
		if err != nil {
			log.Fatalf("can't read email from .gitconfig %v", err)
		}
	}
	me, err := mail.ParseAddress(mailaddr)
	if err != nil {
		log.Fatalf("me %q is an invalid email address: %v", mailaddr, err)
	}

	// Get the environment variable values
	//   samfile := os.Getenv("samfile")
	winids := os.Getenv("winid")
	winid, err := strconv.Atoi(winids)
	if err != nil {
		log.Fatalf("Error converting winid to int: %v", err)
	}

	win, err := acme.Open(winid, nil)
	if err != nil {
		log.Fatalf("failed opening window %d: %v:", winid, err)
	}
	defer win.CloseFiles()

	reader := win.NewReader("body")

	msg, err := mail.ReadMessage(reader)
	if err != nil {
		// TODO(rjk): Make these into log messages.
		log.Fatal("Error parsing email:", err)
		return
	}

	// Extract the To,From
	ofromaddr, err := msg.Header.AddressList("From")
	if err != nil {
		// It's possible here
		log.Fatalf("Error getting From address: %v", err)
	}
	otoaddr, err := msg.Header.AddressList("To")
	if err != nil {
		fmt.Println("Error getting To address:", err)
		return
	}
	occaddr, err := msg.Header.AddressList("CC")
	if err != nil {
		// Missing CC list is not unusual. Simply ignore it.
		occaddr = []*mail.Address{}
	}
	ntoaddr := invertedFrom(ofromaddr, otoaddr, me)
	ntoaddrstring := stringifyAddressList(ntoaddr)

	nccaddr := invertedFrom(nil, occaddr, me)
	nccaddrstring := stringifyAddressList(nccaddr)

	when, err := msg.Header.Date()
	if err != nil {
		log.Println("Don't have a date?", err)
	}

	subject := msg.Header.Get("Subject")
	body := new(bytes.Buffer)
	body.ReadFrom(msg.Body)

	// Select everything in the body so that I can replace it with the replaced email.
	if err := win.Addr("1,$"); err != nil {
		log.Fatalf("oops! Can't write Addr %v", err)
	}
	bodywriter := bufio.NewWriter(win.NewWriter("data"))
	fmt.Fprintf(bodywriter, "From: %s\n", me.String())
	fmt.Fprintf(bodywriter, "To: %s\n", ntoaddrstring)
	if nccaddrstring != "" {
		fmt.Fprintf(bodywriter, "CC: %s\n", nccaddrstring)
	}
	fmt.Fprintf(bodywriter, "Date: %s\n", time.Now().Format(time.RFC1123))
	bodywriter.WriteString(rewriteSubject(subject))
	bodywriter.WriteRune('\n')
	fmt.Fprintf(bodywriter, "On %s,  %s said:\n", when.Format(time.RFC1123Z), ofromaddr[0].String())

	hadnl := true
	for _, r := range body.String() {
		if hadnl {
			hadnl = false
			bodywriter.WriteRune('>')
			bodywriter.WriteRune(' ')
		}
		bodywriter.WriteRune(r)
		if r == '\n' {
			hadnl = true
		}
	}
	bodywriter.Flush()

	// Update the tag.
	if err := win.Fprintf("ctl", "cleartag"); err != nil {
		log.Fatalf("Can't clear tag: %v", err)
	}
	if err := win.Fprintf("tag", " sendit"); err != nil {
		log.Fatalf("Can't send sendit to tag: %v", err)
	}

}

func rewriteSubject(subject string) string {
	if strings.HasPrefix(subject, "Re:") {
		return fmt.Sprintf("Subject: %s\n", subject)
	}
	return fmt.Sprintf("Subject: Re:  %s\n", subject)
}

func invertedFrom(from, to []*mail.Address, me *mail.Address) []*mail.Address {
	unionifying := make(map[string]*mail.Address)

	participants := make([]*mail.Address, 0, len(from)+len(to))
	participants = append(participants, from...)
	participants = append(participants, to...)

	for _, a := range participants {
		if a.Address == me.Address {
			continue
		}
		unionifying[a.Address] = a
	}

	toparticipants := make([]*mail.Address, 0, len(unionifying))
	for _, v := range unionifying {
		toparticipants = append(toparticipants, v)
	}
	return toparticipants
}

func stringifyAddressList(adr []*mail.Address) string {
	builder := strings.Builder{}
	for i, a := range adr {
		if i != 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(a.String())
	}
	return builder.String()
}
