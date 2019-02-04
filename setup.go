package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"whapp-irc/whapp"
)

// TODO: check if already set-up
func (conn *Connection) setup(cancel context.CancelFunc) error {
	if _, err := conn.bridge.Start(); err != nil {
		return err
	}

	go func() {
		// this is actually kind rough, but it seems to work better
		// currently...
		<-conn.bridge.ctx.Done()
		cancel()
	}()

	// if we have the current user in the database, try to relogin using the
	// previous localStorage state
	var user User
	found, err := userDb.GetItem(conn.irc.Nick(), &user)
	if err != nil {
		return err
	}

	if found {
		conn.timestampMap.Swap(user.LastReceivedReceipts)
		conn.chats = user.Chats

		conn.irc.Status("logging in using stored session")

		if err := conn.bridge.WI.Navigate(conn.bridge.ctx); err != nil {
			return err
		}
		if err := conn.bridge.WI.SetLocalStorage(
			conn.bridge.ctx,
			user.LocalStorage,
		); err != nil {
			log.Printf("error while setting local storage: %s\n", err.Error())
		}
	}

	// open site
	state, err := conn.bridge.WI.Open(conn.bridge.ctx)
	if err != nil {
		return err
	}

	// if we aren't logged in yet we have to get the QR code and stuff
	if state == whapp.Loggedout {
		for {
			err := func() error {
				qrcodeScanned := make(chan bool)
				go func() {
					defer close(qrcodeScanned)

					log.Println("before getQrCodesUntilLoggedIn()")
					err := conn.getQrCodesUntilLoggedIn(conn.bridge.ctx)
					if err != nil {
						log.Println("error from getQrCodesUntilLoggedIn()", err.Error())
						qrcodeScanned <- false
						return
					}
					qrcodeScanned <- true
				}()

				log.Println("before block for getQrCodesUntilLoggedIn()")
				scan := <-qrcodeScanned

				if scan {
					// we're logged in, break out
					return nil
				}

				return errors.New("exited qr code wait loop with error")
			}()

			// if no error was received
			// it means we're logged in
			// so exit the loop
			if err == nil {
				break
			}

			// otherwise, keep waiting...
		}

	}
	conn.irc.Status("logged in")

	// get localstorage (that contains new login information), and save it to
	// the database
	conn.localStorage, err = conn.bridge.WI.GetLocalStorage(conn.bridge.ctx)
	if err != nil {
		log.Printf("error while getting local storage: %s\n", err.Error())
	} else {
		if err := conn.saveDatabaseEntry(); err != nil {
			return err
		}
	}

	// get information about the user
	conn.me, err = conn.bridge.WI.GetMe(conn.bridge.ctx)
	if err != nil {
		return err
	}

	// get raw chats
	rawChats, err := conn.bridge.WI.GetAllChats(conn.bridge.ctx)
	if err != nil {
		return err
	}

	// convert chats to internal reprenstation, we do this using a second slice
	// and a WaitGroup to preserve the initial order
	chats := make([]*Chat, len(rawChats))
	var wg sync.WaitGroup
	for i, raw := range rawChats {
		wg.Add(1)
		go func(i int, raw whapp.Chat) {
			defer wg.Done()

			chat, err := conn.convertChat(raw)
			if err != nil {
				str := fmt.Sprintf("error while converting chat with ID %s, skipping", raw.ID)
				conn.irc.Status(str)
				log.Printf("%s. error: %s", str, err)
				return
			}

			chats[i] = chat
		}(i, raw)
	}
	wg.Wait()

	// add all chats to connection
	for _, chat := range chats {
		if chat == nil {
			// there was an error converting this chat, skip it.
			continue
		}

		conn.addChat(chat)
	}

	return nil
}
