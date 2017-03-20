package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

type FileHandler struct{}

var (
	usrManager      *UsrManager
	mux             map[string]func(http.ResponseWriter, *http.Request)
	adminAccountMap map[string]struct{}
	localHostAddr   = ""
)

func init() {
	usrManager = new(UsrManager)
	usrManager.init()

	// Bind address and function
	mux = make(map[string]func(http.ResponseWriter, *http.Request))
	mux["/upload"] = upload
	mux["/setaccess"] = setAccess
	mux["/setadmin"] = setAdmin

	// Initialize the map
	adminAccountMap = make(map[string]struct{})

	// Find the local host address
	host, _ := os.Hostname()
	addrs, _ := net.LookupIP(host)
	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			localHostAddr = ipv4.String()
		}
	}
}

func main() {
	adminAccountMap[SuperAdminAccount] = struct{}{}

	// Create a new http server
	server := http.Server{
		Addr:        Address,
		Handler:     &FileHandler{},
		ReadTimeout: 1 * time.Minute,
	}
	fmt.Println("Listening on :4000")
	// Start serving
	server.ListenAndServe()
}

func (*FileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Return when the same user access frequently
	usrAddr := strings.Split(r.RemoteAddr, ":")[0]
	if !usrManager.canAccess(usrAddr) {
		fmt.Fprintf(w, "Please try again after %d second\n", TimeBoforeRestoreInSecond)
		log.Printf("[%s] attempted to access server, but was denied.\n", r.RemoteAddr)
		return
	}
	// Handle other request except looking over file request
	if h, ok := mux[r.URL.EscapedPath()]; ok {
		h(w, r)
		return
	}
	// Handler looking over file request
	http.FileServer(http.Dir(FileMainDir)).ServeHTTP(w, r)
	log.Printf("[%s] access file server\n", r.RemoteAddr)
}

// Handle the upload file request
func upload(w http.ResponseWriter, r *http.Request) {
	// Record the access logging.
	log.Printf("[%s] %s->%s\n", r.RemoteAddr, "Access upload", r.Method)

	if r.Method == "GET" {
		t, _ := template.ParseFiles(TemplateDir + "/upload.html")
		t.Execute(w, "Upload file")

	} else {
		// r.Method == "POST"

		// Parse file from request
		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile("uploadFile")
		if err != nil {
			fmt.Fprintf(w, "%s", "Upload failed, please try again after refresh this page.")
			log.Printf("[%s] %s --> %s\n", r.RemoteAddr, "r.FormFile() failed", err.Error())
			return
		}

		// Create a file directory if not exist
		createFileDir()
		// Format a new file name, avoid to have duplicated name.
		// Or the new file will replace the old file if their name are same.
		// ex. 123.txt --> 123#hhmmss.txt    123 --> 123#hhmmss
		fileName := handler.Filename
		if strings.Contains(fileName, ".") {
			dotIndex := strings.LastIndex(fileName, ".")
			fileName = fileName[:dotIndex] + time.Now().Format("#150405") + fileName[dotIndex:]
		} else {
			fileName += time.Now().Format("#150405")
		}
		fileName = FileMainDir + "/" + FileSubDir + "/" + fileName
		// Open a new file to get ready for storing
		f, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			fmt.Fprintf(w, "%s", "Upload failed, please try again after refresh this page.")
			log.Printf("[%s] %s --> %s\n", r.RemoteAddr, "Create a new file error", err.Error())
			return
		}
		// Copy file to disk
		fileSize, err := io.Copy(f, file)
		if err != nil {
			fmt.Fprintf(w, "%s", "Upload failed, please try again after refresh this page.")
			log.Printf("[%s] %s --> %s\n", r.RemoteAddr, "Store file error", err.Error())
			return
		}

		// Upload file successfully
		briefName := strings.TrimLeft(fileName, FileMainDir)
		fmt.Fprintf(w, "%s\n", "Upload finished, file's address: "+briefName)
		log.Printf("[%s] %s%d\n", r.RemoteAddr, "Upload finished, file's address: "+briefName+"  file's size: ", fileSize)
	}
}

// Create a file directory if not exist
func createFileDir() {
	if err := os.MkdirAll(FileMainDir+"/"+FileSubDir, 0666); err != nil {
		panic(err.Error())
	}
}

// Used for set access permission for user, operated by administrator.
// acc : The account of administrator
// pwd : The administrator account's password
// usr : The user who you want to operate
// ok  : Whether the user can access server,
//	     true mean can, false mean can't
func setAccess(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	account := r.FormValue("acc")  // The account of administrator
	password := r.FormValue("pwd") // The administrator account's password
	usrAddr := r.FormValue("usr")  // The user who you want to operate
	canAccess := r.FormValue("ok") // Whether the user can access server

	// Return when the account or password is wrong
	_, ok := adminAccountMap[account]
	if !ok || password != AdminPwd {
		fmt.Fprintf(w, "Incorrect account or password.\n")
		log.Printf("[%s](%s) Set access permission for user(%s) failed\n", r.RemoteAddr, account, usrAddr)
		return
	}
	// Can not change permission for local host
	if usrAddr == localHostAddr {
		fmt.Fprintf(w, "You have not enough permission for this operation.\n")
		log.Printf("[%s](%s) Can not change permission for local host\n", r.RemoteAddr, account)
		return
	}
	// No one can not change permission for administrator,
	// except the super administrator
	if _, ok := adminAccountMap[usrAddr]; ok && account != SuperAdminAccount {
		fmt.Fprintf(w, "You have not enough permission for this operation.\n")
		log.Printf("[%s](%s) Can not change permission administrator\n", r.RemoteAddr, account)
		return
	}
	// Set access permission for user
	usrManager.setAccess(usrAddr, canAccess == "true")
	fmt.Fprintf(w, "Change access permission for %s successfully.\n", usrAddr)
	log.Printf("[%s](%s) Change access permission for user(%s), access : %s\n", r.RemoteAddr, account, usrAddr, canAccess)
}

// Set a administrator with specific name, operated by super administrator
// acc : The account of super administrator
// pwd : The super administrator account's password
// name : The name of administrator that you want to set
// ok   : Whether should be a administrator, true mean is, false mean is not
func setAdmin(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	account := r.FormValue("acc")  // The account of super administrator
	password := r.FormValue("pwd") // The super administrator account's password
	name := r.FormValue("name")    // The name of administrator that you want to set
	isAdmin := r.FormValue("ok")   // Whether should be a administrator

	// Return when the account or password is wrong
	if account != SuperAdminAccount || password != SuperAdminAccount {
		fmt.Fprintf(w, "Incorrect account or password.\n")
		log.Printf("[%s](%s) Set administrator permission for user(%s) failed\n", r.RemoteAddr, account, name)
		return
	}
	// Can not change permission for super administrator
	if name == SuperAdminAccount {
		fmt.Fprintf(w, "You have not enough permission for this operation.\n")
		log.Printf("[%s](%s) Can not change permission for super administrator\n", r.RemoteAddr, account)
		return
	}

	if isAdmin == "true" {
		adminAccountMap[name] = struct{}{}
	} else {
		delete(adminAccountMap, name)
	}

	fmt.Fprintf(w, "Set administrator for %s successfully.\n", name)
	log.Printf("[%s](%s) Set administrator fo user(%s), admin : %s\n", r.RemoteAddr, account, name, isAdmin)
}
