// Receive session change notifications from Windows

// Receive session change notifications from Windows.
//
// Example
//     package main
//
//     import (
//     	"github.com/brunoqc/go-windows-session-notifications"
//     	"log"
//     )
//
//     func main() {
//     	quit := make(chan int)
//
//     	changes := make(chan int, 100)
//     	closeChan := make(chan int)
//
//     	go func() {
//     		for {
//     			select {
//     			case c := <-changes:
//     				switch c {
//     				case session_notifications.WTS_SESSION_LOCK:
//     					log.Println("session locked")
//     				case session_notifications.WTS_SESSION_UNLOCK:
//     					log.Println("session unlocked")
//     				}
//     			}
//     		}
//     	}()
//
//     	session_notifications.Subscribe(changes, closeChan)
//
//     	// ctrl+c to quit
//     	<-quit
//     }
//
package session_notifications

/*
#include <windows.h>
extern HANDLE Start();
extern void Stop(HANDLE);
*/
import "C"

// http://msdn.microsoft.com/en-us/library/aa383828(v=vs.85).aspx
const (
	WTS_CONSOLE_CONNECT        = 0x1
	WTS_CONSOLE_DISCONNECT     = 0x2
	WTS_REMOTE_CONNECT         = 0x3
	WTS_REMOTE_DISCONNECT      = 0x4
	WTS_SESSION_LOGON          = 0x5
	WTS_SESSION_LOGOFF         = 0x6
	WTS_SESSION_LOCK           = 0x7
	WTS_SESSION_UNLOCK         = 0x8
	WTS_SESSION_REMOTE_CONTROL = 0x9
	WTS_SESSION_CREATE         = 0xA
	WTS_SESSION_TERMINATE      = 0xB
)

var (
	changes = make(chan int, 1000)
)

//export sessionChange
func sessionChange(value int) {
	changes <- value
}

// Subscribe will make it so that subChan will receive the session events.
// To unsubscribe, close closeChan
func Subscribe(subChan chan int, closeChan chan int) {
	var threadHandle C.HANDLE
	go func() {
		for {
			select {
			case <-closeChan:
				C.Stop(threadHandle)
				return
			case c := <-changes:
				subChan <- c
			}
		}
	}()
	threadHandle = C.Start()
}
