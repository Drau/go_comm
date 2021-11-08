---
id: MXoSu
name: Server manager
file_version: 1.0.2
app_version: 0.6.5-2
file_blobs:
  conn_manager.go: 3ac47654c323500dc190ea7359cc5e755f42c9db
---

We have a main go routine in charge of incoming connections to the app.

basically its an infinite loop waiting for new connections then starting a new go routine per connection.

<br/>

<!-- NOTE-swimm-snippet: the lines below link your snippet to Swimm -->
### 📄 conn_manager.go
```go
⬜ 89     	}
⬜ 90     	defer l.Close()
⬜ 91     
🟩 92     	for {
🟩 93     		s, err := l.Accept()
🟩 94     		if err != nil {
🟩 95     			fmt.Fprintf(logs, "Error connecting to %s:%s - %s\n", connHost, connPort, err.Error())
🟩 96     			return
🟩 97     		}
🟩 98     		remoteUser := user{name: RandomString(5), above18: true, conn: s, active: true}
🟩 99     		fmt.Fprintf(logs, "Client "+s.RemoteAddr().String()+" connected.\n")
🟩 100    		conns[remoteUser.name] = &remoteUser
🟩 101    		loggedUsersUpdate <- true
🟩 102    		sendDataToNewClient(conns["local"], &remoteUser, logs)
🟩 103    		go handleConnection(conns, &remoteUser, chat, logs, loggedUsersUpdate)
🟩 104    	}
⬜ 105    }
⬜ 106    
⬜ 107    func connectClient(conns connections, host string, chat *tview.TextView, logs *tview.TextView, loggedUsersUpdate chan bool, propogateConnect bool) bool {
```

<br/>

This file was generated by Swimm. [Click here to view it in the app](https://app.swimm.io/repos/Z2l0aHViJTNBJTNBZ29fY29tbSUzQSUzQURyYXU=/docs/MXoSu).