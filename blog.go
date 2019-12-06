package main

import (
	"crypto/subtle"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"bf.go/blog"
	"github.com/Ullaakut/nmap"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

type nmapSession struct {
	SessionID   int64  `json:"sessionid"`
	SessionName string `json:"session_name"`
}

type nmapHost struct {
	HostID    int64  `json:"host_id"`
	HostAddr  string `json:"host_addr"`
	SessionID int64  `json:"session_id"`
}

type nmapResult struct {
	ResultID    int64  `json:"result_id"`
	Port        uint16 `json:"port"`
	Status      string `json:"status"`
	Protocol    string `json:"protocol"`
	ServiceName string `json:"service_name"`
	HostID      int64  `json:"host_id"`
}

type jsonErrorMessage struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

func main() {

	r := mux.NewRouter()

	r.HandleFunc("/", blog.HomeHandler)
	r.Handle("/", handlers.LoggingHandler(os.Stdout, http.HandlerFunc(blog.HomeHandler)))
	r.Handle("/about", handlers.LoggingHandler(os.Stdout, blog.AuthHandler(http.HandlerFunc(blog.AboutHandler)))).Methods("GET")
	// "Signin" and "Welcome" are the handlers that we will implement
	r.Handle("/signin", handlers.LoggingHandler(os.Stdout, http.HandlerFunc(blog.Signin))).Methods("GET", "POST")
	r.Handle("/welcome", handlers.LoggingHandler(os.Stdout, http.HandlerFunc(blog.Welcome))).Methods("GET", "POST")
	//r.Handle("/api/v1/putnmap/{id}", handlers.LoggingHandler(os.Stdout, blog.AuthHandler(pool, http.HandlerFunc(putNmapHandler))))
	//r.Handle("/api/v1/putimage/{id}", handlers.LoggingHandler(os.Stdout, http.HandlerFunc(putImage)))
	//r.Handle("/api/v1/get_nmap_sessions", handlers.LoggingHandler(os.Stdout, http.HandlerFunc(getNmapSessions)))
	//r.Handle("/api/v1/get_host_by_session_id/{id}", handlers.LoggingHandler(os.Stdout, http.HandlerFunc(getHostsBySessionID)))
	//r.Handle("/api/v1/get_results_by_host_id/{id}", handlers.LoggingHandler(os.Stdout, http.HandlerFunc(getResultsByHostID)))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	http.Handle("/", r)

	srv := &http.Server{
		Handler: r,
		Addr:    "0.0.0.0:5000",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	//log.Fatal(http.ListenAndServe("0.0.0.0:5000", n))

	log.Fatal(srv.ListenAndServe())
}

func getNmapSessions(w http.ResponseWriter, r *http.Request) {

	db, err := getDatabaseConnection()

	if err != nil {
		jsonError(err, w)
		return
	}
	defer db.Close()

	ns, err := getAllNmapSessions(db)
	if err != nil {
		jsonError(err, w)
		return
	}

	byteResult, err := json.Marshal(ns)

	nrResult := string(byteResult)

	w.WriteHeader(http.StatusOK)

	w.Header().Set("Content-Type", "application/json")

	fmt.Fprintf(w, "{\"status\":\"success\", \"message\": \"nmap sessions success\",\"sessions\":%s}\n", nrResult)
}

func getHostsBySessionID(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	db, err := getDatabaseConnection()

	if err != nil {
		jsonError(err, w)
		return
	}
	defer db.Close()

	i, err := strconv.ParseInt(vars["id"], 10, 64)

	nh, err := getAllHostBySessionID(db, i)
	if err != nil {
		jsonError(err, w)
		return
	}

	byteResult, err := json.Marshal(nh)

	nhResult := string(byteResult)

	w.WriteHeader(http.StatusOK)

	w.Header().Set("Content-Type", "application/json")

	fmt.Fprintf(w, "{\"status\":\"success\", \"message\": \"hosts success\",\"hosts\":%s}\n", nhResult)
}

func getResultsByHostID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	db, err := getDatabaseConnection()

	if err != nil {
		jsonError(err, w)
		return
	}
	defer db.Close()

	i, err := strconv.ParseInt(vars["id"], 10, 64)

	nh, err := getAllResultsByHostID(db, i)
	if err != nil {
		jsonError(err, w)
		return
	}

	byteResult, err := json.Marshal(nh)

	nhResult := string(byteResult)

	w.WriteHeader(http.StatusOK)

	w.Header().Set("Content-Type", "application/json")

	fmt.Fprintf(w, "{\"status\":\"success\", \"message\": \"hosts success\",\"results\":%s}\n", nhResult)
}

func putImage(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	r.ParseMultipartForm(10 << 20)

	file, header, err := r.FormFile("data")
	if err != nil {
		jsonError(err, w)
		return
	}

	fileName := header.Filename

	defer file.Close()

	tempFile, err := ioutil.TempFile("temp", fileName)
	if err != nil {
		fmt.Println(err)
	}
	defer tempFile.Close()

	// read all of the contents of our uploaded file into a
	// byte array
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		jsonError(err, w)
		return
	}

	tempFile.Write(fileBytes)

	w.WriteHeader(http.StatusOK)

	w.Header().Set("Content-Type", "application/json")

	fmt.Fprintf(w, "{\"status\":\"success\", \"message\": \"added image to %s\"}\n", vars["id"])

	return

}

func putNmapHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	r.ParseMultipartForm(10 << 20)

	file, _, err := r.FormFile("data")
	if err != nil {
		jsonError(err, w)
		return
	}

	defer file.Close()

	// read all of the contents of our uploaded file into a
	// byte array
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		jsonError(err, w)
		return
	}

	err = loadNmapSession(vars["id"], fileBytes)
	if err != nil {
		jsonError(err, w)
		return
	}

	w.WriteHeader(http.StatusOK)

	w.Header().Set("Content-Type", "application/json")

	fmt.Fprintf(w, "{\"status\":\"success\", \"message\": \"nmap session %s uploaded\"}\n", vars["id"])

	return
}

func jsonError(err error, w http.ResponseWriter) {

	var jerror jsonErrorMessage

	jerror.Message = err.Error()
	jerror.Status = "error"

	byteError, err := json.Marshal(jerror)
	if err != nil {
		fmt.Printf("Could not marshal error into json string with error %s\n", err)
	}

	errorString := string(byteError)

	w.WriteHeader(http.StatusOK)

	w.Header().Set("Content-Type", "application/json")

	fmt.Fprintf(w, errorString)

	return
}

func loadNmapSession(scanid string, fileBytes []byte) error {

	result, err := nmap.Parse(fileBytes)
	if err != nil {
		return err
	}

	if len(result.Hosts) == 0 {
		return fmt.Errorf("no hosts available in the scan for session: %s", scanid)
	}

	db, err := getDatabaseConnection()

	defer db.Close()

	if err != nil {
		fmt.Printf("Error connecting to the database: %s\n", err)
		return err
	}

	exists, err := nmapSessionExists(db, scanid)
	if err != nil {
		fmt.Println(fmt.Errorf("Error finding nmap session: %s", err))
	}

	if exists == false {
		fmt.Printf("nmap session %s does not exist, creating it\n", scanid)
		_, err := insertNmapSession(db, scanid)
		if err != nil {
			fmt.Printf("Error finding Nmap session: %e\n", err)
			return err
		}
	} else {
		return fmt.Errorf("Nmap session already exists, please use a different session name")
	}

	session, err := getNmapSessionByName(db, scanid)
	if err != nil {
		fmt.Printf("Error finding Nmap session: %e\n", err)
		return err
	}

	fmt.Printf("Session is %d \n", session.SessionID)

	for _, host := range result.Hosts {
		if len(host.Ports) == 0 || len(host.Addresses) == 0 {
			continue
		}

		var nh nmapHost
		nh.HostAddr = host.Addresses[0].Addr
		nh.SessionID = session.SessionID

		insertHost(db, &nh)

		for _, port := range host.Ports {

			var nr nmapResult
			nr.Port = port.ID
			nr.Status = port.State.State
			nr.Protocol = port.Protocol
			nr.ServiceName = port.Service.Name
			nr.HostID = nh.HostID

			insertHostResult(db, &nr)

		}
	}
	return nil
}

func basicAuth(w http.ResponseWriter, r *http.Request, username, password, realm string) bool {

	user, pass, ok := r.BasicAuth()

	if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(username)) != 1 || subtle.ConstantTimeCompare([]byte(pass), []byte(password)) != 1 {
		w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
		w.WriteHeader(401)
		w.Write([]byte("Unauthorised.\n"))
		return false
	}

	return true
}

func getDatabaseConnection() (*sql.DB, error) {

	if _, err := os.Stat("./nmap.db"); os.IsNotExist(err) {
		fmt.Println("error accessing sqlite database")
	}

	database, err := sql.Open("sqlite3", "./nmap.db")

	if err != nil {
		fmt.Println(fmt.Errorf("error accessing sqlite database: %s", err))
	}

	return database, nil
}

func insertNmapSession(db *sql.DB, name string) (*nmapSession, error) {

	stmt, err := db.Prepare("INSERT INTO nmap_sessions (session_name) values (?)")
	if err != nil {
		return nil, err
	}

	res, err := stmt.Exec(name)
	if err != nil {
		return nil, err
	}

	var session nmapSession

	session.SessionName = name
	session.SessionID, err = res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &session, nil
}

func nmapSessionExists(db *sql.DB, name string) (bool, error) {

	ns, err := getNmapSessionByName(db, name)
	if err != nil {
		return false, nil
	}

	if ns != nil {
		return true, nil
	}

	if err != nil {
		return false, err
	}

	return false, nil

}

func getNmapSessionByName(db *sql.DB, name string) (*nmapSession, error) {

	stmt, err := db.Prepare("SELECT id, session_name FROM nmap_sessions where session_name=?")
	if err != nil {
		return nil, err
	}

	var ns nmapSession
	err = stmt.QueryRow(name).Scan(&ns.SessionID, &ns.SessionName)
	if err != nil {
		return nil, err
	}

	return &ns, nil
}

func getAllNmapSessions(db *sql.DB) ([]nmapSession, error) {

	rows, err := db.Query("SELECT id, session_name FROM nmap_sessions")
	if err != nil {
		return nil, err
	}
	var nmapSessions []nmapSession
	for rows.Next() {
		var ns nmapSession
		rows.Scan(&ns.SessionID, &ns.SessionName)
		nmapSessions = append(nmapSessions, ns)
	}

	return nmapSessions, nil
}

func insertHost(db *sql.DB, host *nmapHost) error {

	stmt, err := db.Prepare("INSERT INTO hosts (session_id, host_addr) values (?,?)")
	if err != nil {
		return err
	}

	res, err := stmt.Exec(host.SessionID, host.HostAddr)
	if err != nil {
		return err
	}

	host.HostID, err = res.LastInsertId()
	if err != nil {
		return err
	}

	return nil
}

func getAllHostBySessionID(db *sql.DB, sessionID int64) ([]nmapHost, error) {

	rows, err := db.Query("SELECT id, session_id, host_addr FROM hosts WHERE session_id=?", sessionID)
	if err != nil {
		return nil, err
	}

	var nmapHosts []nmapHost
	for rows.Next() {
		var nh nmapHost
		rows.Scan(&nh.HostID, &nh.SessionID, &nh.HostAddr)
		nmapHosts = append(nmapHosts, nh)
	}

	return nmapHosts, nil
}

func getAllHosts(db *sql.DB) ([]nmapHost, error) {

	rows, err := db.Query("SELECT id, session_id, host_addr FROM hosts")
	if err != nil {
		return nil, err
	}
	var nmapHosts []nmapHost
	for rows.Next() {
		var nh nmapHost
		rows.Scan(&nh.HostID, &nh.SessionID, &nh.HostAddr)
		nmapHosts = append(nmapHosts, nh)
	}

	return nmapHosts, nil

}

func getAllResultsByHostID(db *sql.DB, id int64) ([]nmapResult, error) {

	rows, err := db.Query("SELECT id, port, status, protocol, service_name, host_id FROM results WHERE host_id=?", id)
	if err != nil {
		return nil, err
	}

	var nmapResults []nmapResult
	for rows.Next() {
		var nr nmapResult
		rows.Scan(&nr.ResultID, &nr.Port, &nr.Status, &nr.Protocol, &nr.ServiceName, &nr.HostID)
		nmapResults = append(nmapResults, nr)
	}

	return nmapResults, nil
}

func getHostAllResults(db *sql.DB) ([]nmapResult, error) {

	rows, err := db.Query("SELECT id, port, status, protocol, service_name, host_id FROM results")
	if err != nil {
		return nil, err
	}
	var nmapResults []nmapResult
	for rows.Next() {
		var nr nmapResult
		rows.Scan(&nr.ResultID, &nr.Port, &nr.Status, &nr.Protocol, &nr.ServiceName, &nr.HostID)
		nmapResults = append(nmapResults, nr)
	}

	return nmapResults, nil
}

func insertHostResult(db *sql.DB, result *nmapResult) error {

	stmt, err := db.Prepare("INSERT INTO results (host_id, port, status, protocol, service_name) values (?,?,?,?,?)")
	if err != nil {
		return err
	}

	res, err := stmt.Exec(result.HostID, result.Port, result.Status, result.Protocol, result.ServiceName)
	if err != nil {
		return err
	}

	result.ResultID, err = res.LastInsertId()
	if err != nil {
		return err
	}

	return nil
}
