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
//     	chanMessages := make(chan session_notifications.Message, 100)
//     	chanClose := make(chan int)
//
//     	go func() {
//     		for {
//     			select {
//     			case m := <-chanMessages:
//     				switch m.UMsg {
//     				case session_notifications.WM_WTSSESSION_CHANGE:
//     					switch m.Param {
//     					case session_notifications.WTS_SESSION_LOCK:
//     						log.Println("session locked")
//     					case session_notifications.WTS_SESSION_UNLOCK:
//     						log.Println("session unlocked")
//     					}
//     				case session_notifications.WM_QUERYENDSESSION:
//     					log.Println("log off or shutdown")
//     				}
//     				close(m.ChanOk)
//     			}
//     		}
//     	}()
//
//     	session_notifications.Subscribe(chanMessages, chanClose)
//
//     	// ctrl+c to quit
//     	<-quit
//     }
//
package session_notifications

// #cgo LDFLAGS: -lwtsapi32
/*
#include <windows.h>
extern HANDLE Start();
extern void Stop(HANDLE);
*/
import "C"

import (
	"syscall"
)

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

	WM_QUERYENDSESSION   = 0x11
	WM_WTSSESSION_CHANGE = 0x2B1

	ENDSESSION_CLOSEAPP = 0x00000001
	ENDSESSION_CRITICAL = 0x40000000
	ENDSESSION_LOGOFF   = 0x80000000
)

type Message struct {
	UMsg   int
	Param  int
	ChanOk chan int
}

var (
	chanMessages = make(chan Message, 1000)

	kernel32    = syscall.MustLoadDLL("kernel32.dll")
	CloseHandle = kernel32.MustFindProc("CloseHandle")
)

//export relayMessage
func relayMessage(message C.uint, wParam C.uint) {
	msg := Message{
		UMsg:  int(message),
		Param: int(wParam),
	}
	msg.ChanOk = make(chan int)

	chanMessages <- msg

	// wait for the app to do it's thing
	// it's usefull for WM_QUERYENDSESSION if we need time to save before Windows shutdown
	<-msg.ChanOk
}

// Subscribe will make it so that subChan will receive the session events.
// chanSessionEnd will receive a '1' when the session ends (when Windows shut down)
// To unsubscribe, close closeChan
// You must close 'ChanOk' after processing the event. This channel is to give you time to save if the event is WM_QUERYENDSESSION
func Subscribe(subchanMessages chan Message, closeChan chan int) {
	var threadHandle C.HANDLE
	go func() {
		for {
			select {
			case <-closeChan:
				C.Stop(threadHandle)
				r, _, err := CloseHandle.Call(uintptr(threadHandle))
				if r == 0 {
					panic(err)
				}

				return
			case c := <-chanMessages:
				subchanMessages <- c
			}
		}
	}()
	threadHandle = C.Start()
}
