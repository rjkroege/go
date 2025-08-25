package main

import (
	"log"
	"net/mail"
	"os"
	"os/exec"
	"strconv"

	"9fans.net/go/acme"
)

func main() {
	log.Println("hi")

	// Get the environment variable values
	samfile := os.Getenv("samfile")
	winids := os.Getenv("winid")
	winid, err := strconv.Atoi(winids)
	if err != nil {
		log.Fatalf("Error converting winid to int:", err)
	}

	log.Println("samfile", samfile, "winid", winid)

	win, err := acme.Open(winid, nil)
	if err != nil {
		log.Fatalf("failed opening window %d: %v:", winid, err)
	}
	defer win.CloseFiles()

	wininfo, err := win.Info()
	if err != nil {
		log.Fatalf("failed getting info for window %d: %v:", winid, err)
	}

	log.Println(wininfo)

	if wininfo.IsModified {
		log.Fatalf("Save %s before trying to send", samfile)
	}

	// It's possible that reading it from Acme/Edwood is slower than the
	// buffer cache? I claim that this doesn't matter. Also: I should make
	// the filesystem fast.
	reader := win.NewReader("body")
	msg, err := mail.ReadMessage(reader)
	if err != nil {
		log.Fatal("Error parsing email:", err)
		return
	}

	// Extract the To, From
	fromaddr, err := msg.Header.AddressList("From")
	if err != nil {
		log.Fatalf("Error getting From address: %v", err)
	}
	toaddr, err := msg.Header.AddressList("To")
	if err != nil {
		log.Fatalf("Error getting To address:", err)
		return
	}

	// rewind
	if _, err := win.Seek("body", 0, 0); err != nil {
		log.Fatalf("Can't rewind the body: %v", err)
	}

	// Fork the sender.
	args := make([]string, 0)
	args = append(args, "-sender", fromaddr[0].Address)
	for _, a := range toaddr {
		args = append(args, a.Address)
	}
	// Perhaps sendgmail should have an absolute path?
	// "sendgmail" is small enough that one might consider inlining it?
	cmd := exec.Command("sendgmail", args...)

	log.Println("args", args)

	cmd.Stdin = reader
	if err := cmd.Run(); err != nil {
		log.Fatalf("didn't send email: %v", err)
	}

	// Close the window.
	if err := win.Ctl("del"); err != nil {
	}

	// Remove the file if we sent it.
	if err := os.Remove(samfile); err != nil {
		log.Fatalf("can't remove the file %s: %v", samfile, err)
	}
}
