package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -lobjc -framework Foundation -framework CoreFoundation -framework AppKit -framework Cocoa

#import <DetectorAppDelegate.h>
*/
import "C"
import (
	"fmt"
	"github.com/bramvdbogaerde/go-scp"
	"github.com/bramvdbogaerde/go-scp/auth"
	"github.com/progrium/macdriver/cocoa"
	"github.com/progrium/macdriver/core"
	"github.com/progrium/macdriver/objc"
	"golang.design/x/clipboard"
	"golang.org/x/crypto/ssh"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

type Screenshot struct {
	path      string
	createdAt string
}

const BundleId = "com.revilon1991.screenshot"

var screenshotChan = make(chan Screenshot)
var screenshotPull = make(map[string]Screenshot)

var application cocoa.NSApplication

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	runtime.LockOSThread()

	makeDbIfNotExist()

	go listener()

	application = cocoa.NSApplication{Object: objc.Get("NSApplication").Send("sharedApplication")}
	applicationDelegate := objc.Get("DetectorAppDelegate").Alloc().Init()

	application.Send("setDelegate:", applicationDelegate)
	application.SetActivationPolicy(cocoa.NSApplicationActivationPolicyAccessory)

	application.Run()
}

//export openPreferences
func openPreferences() {
	userConfig := getUserConfig()

	hostField := objc.Get("NSTextField").Alloc().Send("initWithFrame:", core.Rect(50, 250, 300, 20.0))
	hostField.Set("placeholderString:", core.NSString_FromString("Host"))

	portField := objc.Get("NSTextField").Alloc().Send("initWithFrame:", core.Rect(50, 220, 300, 20.0))
	portField.Set("placeholderString:", core.NSString_FromString("Port"))

	userField := objc.Get("NSTextField").Alloc().Send("initWithFrame:", core.Rect(50, 190, 300, 20.0))
	userField.Set("placeholderString:", core.NSString_FromString("User"))

	passField := objc.Get("NSTextField").Alloc().Send("initWithFrame:", core.Rect(50, 160, 300, 20.0))
	passField.Set("placeholderString:", core.NSString_FromString("Pass"))

	pathField := objc.Get("NSTextField").Alloc().Send("initWithFrame:", core.Rect(50, 130, 300, 20.0))
	pathField.Set("placeholderString:", core.NSString_FromString("Path"))

	linkField := objc.Get("NSTextField").Alloc().Send("initWithFrame:", core.Rect(50, 100, 300, 20.0))
	linkField.Set("placeholderString:", core.NSString_FromString("Link"))

	if userConfig != nil {
		hostField.Set("stringValue:", core.NSString_FromString(userConfig.host))
		portField.Set("stringValue:", core.NSString_FromString(strconv.Itoa(userConfig.port)))
		userField.Set("stringValue:", core.NSString_FromString(userConfig.user))
		passField.Set("stringValue:", core.NSString_FromString(userConfig.pass))
		pathField.Set("stringValue:", core.NSString_FromString(userConfig.path))
		linkField.Set("stringValue:", core.NSString_FromString(userConfig.link))
	}

	saveButton := objc.Get("NSButton").Alloc().Send("initWithFrame:", core.Rect(50, 30, 50, 20.0))
	saveButton.Set("title:", core.NSString_FromString("Save"))
	saveButton.Set("action:", objc.Sel("savePreferencesSel"))

	window := cocoa.NSWindow_Init(
		core.Rect(0, 0, 400, 300),
		cocoa.NSClosableWindowMask|cocoa.NSTitledWindowMask|cocoa.NSMiniaturizableWindowMask|cocoa.NSResizableWindowMask,
		cocoa.NSBackingStoreBuffered,
		false,
	)
	window.MakeKeyAndOrderFront(window)
	window.SetTitle("Preferences (SFTP)")
	window.Center()

	window.ContentView().Send("addSubview:", hostField)
	window.ContentView().Send("addSubview:", portField)
	window.ContentView().Send("addSubview:", userField)
	window.ContentView().Send("addSubview:", passField)
	window.ContentView().Send("addSubview:", pathField)
	window.ContentView().Send("addSubview:", linkField)
	window.ContentView().Send("addSubview:", saveButton)
}

//export openHelp
func openHelp() {
	helpView := cocoa.NSTextView_Init(core.Rect(50, 250, 300, 20.0))
	helpView.SetEditable(false)
	helpView.SetString(`
Screenshot.

This application provides delivering screenshots to your sftp server.
Just fill fields in the preferences menu and take screenshot.
Generally, you can do it by combination shift+cmd+4.
After you can put through the link from your buffer by cmd+v.

This is an open-source application.
You can become a contributor developing with this repo https://github.com/revilon1991/screenshot

RevilOn <revil-on@mail.ru>
`)

	window := cocoa.NSWindow_Init(
		core.Rect(0, 0, 400, 300),
		cocoa.NSClosableWindowMask|cocoa.NSTitledWindowMask|cocoa.NSMiniaturizableWindowMask|cocoa.NSResizableWindowMask,
		cocoa.NSBackingStoreBuffered,
		false,
	)
	window.MakeKeyAndOrderFront(window)
	window.SetTitle("Help")
	window.Center()

	window.ContentView().Send("addSubview:", helpView)
}

//export savePreferences
func savePreferences() {
	keyWindow := cocoa.NSWindow{Object: application.Send("keyWindow")}
	view := keyWindow.ContentView()

	subviews := core.NSArray{Object: view.Send("subviews")}

	var sftpConfig = make(map[string]string)
	for i := uint64(0); i < subviews.Count(); i++ {
		if subviews.ObjectAtIndex(i).Class().String() != "NSTextField" {
			continue
		}

		key := subviews.ObjectAtIndex(i).Get("placeholderString").String()
		value := subviews.ObjectAtIndex(i).Send("stringValue").String()

		sftpConfig[key] = value
	}

	if len(sftpConfig["Host"]) == 0 ||
		len(sftpConfig["Port"]) == 0 ||
		len(sftpConfig["User"]) == 0 ||
		len(sftpConfig["Pass"]) == 0 ||
		len(sftpConfig["Path"]) == 0 ||
		len(sftpConfig["Link"]) == 0 {
		errorAlert := objc.Get("NSAlert").Alloc().Init()
		errorAlert.Send("setMessageText:", core.NSString_FromString("Invalid parameters"))
		errorAlert.Send("setInformativeText:", core.NSString_FromString("All fields must be filled"))
		errorAlert.Send("addButtonWithTitle:", core.NSString_FromString("OK"))
		errorAlert.Send("runModal")

		return
	}

	port, _ := strconv.Atoi(sftpConfig["Port"])

	setUserConfig(
		sftpConfig["Host"],
		port,
		sftpConfig["User"],
		sftpConfig["Pass"],
		sftpConfig["Path"],
		sftpConfig["Link"],
	)

	keyWindow.Close()
}

//export ui
func ui() {
	obj := cocoa.NSStatusBar_System().StatusItemWithLength(cocoa.NSVariableStatusItemLength)
	obj.Retain()
	obj.Button().SetTitle("📸")
	menu := cocoa.NSMenu_New()
	itemQuit := cocoa.NSMenuItem_Init("Quit", objc.Sel("terminate:"), "q")
	itemPreferences := cocoa.NSMenuItem_Init("Preferences", objc.Sel("openPreferencesSel"), "s")
	itemHelp := cocoa.NSMenuItem_Init("Help", objc.Sel("openHelpSel"), "h")
	menu.AddItem(itemHelp)
	menu.AddItem(itemPreferences)
	menu.AddItem(itemQuit)
	obj.SetMenu(menu)

	NSBundle := cocoa.NSBundle_Main().Class()
	NSBundle.AddMethod("__bundleIdentifier", func(_ objc.Object) objc.Object {
		return core.String(BundleId)
	})
	NSBundle.Swizzle("bundleIdentifier", "__bundleIdentifier")
}

//export queryWire
func queryWire(pathPointer unsafe.Pointer, createdAtPointer unsafe.Pointer) {
	if pathPointer == nil || createdAtPointer == nil {
		return
	}

	path := objc.ObjectPtr(uintptr(pathPointer))
	path.Retain()
	pathString := fmt.Sprint(path)

	createdAt := objc.ObjectPtr(uintptr(createdAtPointer))
	createdAt.Retain()
	createdAtString := fmt.Sprint(createdAt)

	screenshotChan <- Screenshot{
		path:      pathString,
		createdAt: createdAtString,
	}
}

func listener() {
	now := time.Now().Unix()

	for {
		select {
		case screenshot := <-screenshotChan:
			if _, found := screenshotPull[screenshot.path]; found {
				continue
			}

			createdAt, _ := time.Parse("2006-01-02 15:04:05 -0700", screenshot.createdAt)

			if now > createdAt.Unix() {
				continue
			}

			screenshotPull[screenshot.path] = screenshot

			userConfig := getUserConfig()

			fileLink, err := sendFile(userConfig, screenshot.path)

			if err != nil {
				fmt.Printf("Error while copying file %s", err.Error())
				sendNotify("Error while copying file", err.Error())

				continue
			}

			clipboard.Write(clipboard.FmtText, []byte(fileLink))

			sendNotify("Screenshot delivered", fileLink)
		}
	}
}

func sendFile(userConfig *UserConfig, screenshotPath string) (string, error) {
	clientConfig, _ := auth.PasswordKey(
		userConfig.user,
		userConfig.pass,
		ssh.InsecureIgnoreHostKey(),
	)

	client := scp.NewClient(
		fmt.Sprintf("%s:%s", userConfig.host, strconv.Itoa(userConfig.port)),
		&clientConfig,
	)

	err := client.Connect()

	if err != nil {
		fmt.Printf("Couldn't establish a connection to the remote server %s", err.Error())
		sendNotify("Couldn't establish a connection to the remote server", err.Error())

		return "", err
	}

	file, _ := os.Open(screenshotPath)
	filename := filepath.Base(file.Name())

	defer client.Close()
	defer func() {
		err = file.Close()

		if err != nil {
			panic(err)
		}
	}()

	filename = strings.ReplaceAll(filename, " ", "_")
	filePath := fmt.Sprintf("%s%s", userConfig.path, filename)

	err = client.CopyFile(
		file,
		filePath,
		"0655",
	)

	if err != nil {
		return "", err
	}

	fileLink := fmt.Sprintf("%s%s", userConfig.link, filename)

	return fileLink, nil
}

func sendNotify(title string, text string) {
	notification := objc.Get("NSUserNotification").Alloc().Init()
	notification.Set("title:", core.String(title))
	notification.Set("informativeText:", core.String(text))
	center := objc.Get("NSUserNotificationCenter").Send("defaultUserNotificationCenter")
	center.Send("deliverNotification:", notification)
	notification.Release()
}
