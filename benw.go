package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/gorilla/mux"
	//"github.com/julienschmidt/httprouter"
)

// Model of stuff to render a page
type Model struct {
	Title string
	Name  string
}

// “export” means “public” => with upper case first letter
// the reason is simple: the renderer package use the reflect package in order to get/set fields values
// the reflect package can only access public/exported struct fields.
// So try defining
type pageData struct {
	Title  string
	Navbar string
}

type messData struct {
	Rt      string
	DT      string
	Message string
}

type ouData struct {
	OU       string
	Name     string
	Ad       bool
	Desc     string
	User     string
	Pass     string
	Mc       int
	Mess     [6]messData
	ChkMess  bool
	ChkAudit bool
	ShMess   string
}

var (
	pagedata pageData
	ou       ouData
)

var (
	server   string
	port     = 1433
	user     string
	password string
	defuser  string
	defpass  string
	database = "ben"
)

func testHandler(w http.ResponseWriter, r *http.Request) {

	data := pageData{Title: "Project", Navbar: "hello"}
	tmpl := template.Must(template.ParseFiles("test.html"))
	err := tmpl.Execute(w, data)
	//err := tmpl.Execute(os.Stdout, data)
	if err != nil {
		panic(err)
	}
}

func messcount(db *sql.DB) (int, error) {
	var count int
	tsql := fmt.Sprintf("SELECT count(*) as mycount FROM [BEN].[dbo].[Messages] " +
		"WHERE receivetime > DATEADD(HOUR, -1, GETDATE())")

	row := db.QueryRow(tsql)
	err := row.Scan(&count)
	if err != nil {
		log.Fatal(err)
	}
	//defer row.Close()

	return count, err
}

func summary(db *sql.DB) (int, error) {
	tsql := fmt.Sprintf("SELECT top 6 ScadaSystemDate,convert(varchar, ReceiveTime, 21) as RT,Message," +
		"DATEDIFF(SECOND ,ScadaSystemDate, ReceiveTime) as DT FROM BEN.dbo.Messages" +
		" order by ReceiveTime desc;")
	rows, err := db.Query(tsql)
	if err != nil {
		fmt.Println("Error reading rows: " + err.Error())
		return -1, err
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		var ScadaSystemDate, RT, Message, DT string

		err := rows.Scan(&ScadaSystemDate, &RT, &Message, &DT)
		if err != nil {
			fmt.Println("Error reading rows: " + err.Error())
			return -1, err
		}
		//layout := "2006-01-02"
		//t, _ := time.Parse(layout, ReceiveTime)
		//ou.Mess[count].Rt = t
		ou.Mess[count].Rt = RT
		ou.Mess[count].Message = Message
		ou.Mess[count].DT = DT

		// fmt.Println("Last Message received   : " + ReceiveTime)
		// fmt.Println("Last Message Scada Time : " + ScadaSystemDate)
		// fmt.Println("Last Message            : " + Message)
		// fmt.Println("Last Message DT         : " + DT)

		//fmt.Printf("%s \t %s \t %s \n", ReceiveTime, Message, DT)
		count++
	}
	return count, nil
}

func regiontHandler(w http.ResponseWriter, r *http.Request) {

	//Need to do something with this !!!
	defuser = "elec\\klopperd"
	defpass = "detany1910"

	switch r.Method {
	case "GET":

		switch ou.OU {
		case "ec":
			ou.Name = "ECOU"
			ou.Desc = "Eastern Cape OU"
			server = "elnvmsa006\\scadaapps"
			user = defuser
			password = defpass
		case "kz":
			ou.Name = "KZNOU"
			ou.Desc = "KZN OU"
			server = "mkdvmsa006\\scadaapps"
			user = defuser
			password = defpass
		case "gt":
			ou.Name = "GOU"
			ou.Desc = "Gauteng OU"
			server = "spnvmsa010\\scadaapps"
			user = defuser
			password = defpass
		case "mp":
			ou.Name = "MPOU"
			ou.Desc = "Mpumalanga OU"
			server = "wtkvmsa008\\scadaapps"
			user = defuser
			password = defpass
		case "wc":
			ou.Name = "WCOU"
			ou.Desc = "Western Cape OU"
			server = "blvvmsa014\\scadaapps"
			user = defuser
			password = defpass
		default:
			ou.Name = "TEST"
			ou.Desc = "Test Environment "
			server = "hwhvmsa004"
			user = "qaelec\\klopperd"
			password = "Eskom899#"
		}

		connString := fmt.Sprintf("server=%s;user id=%s;password=%s;database=%s;",
			server, user, password, database)

		conn, err := sql.Open("mssql", connString)
		if err != nil {
			log.Fatal("Open connection failed:", err.Error())
		}
		fmt.Printf("Connected!\n")
		//fmt.Println(reg)
		defer conn.Close()

		// Get Last hour Count
		mscount, mserr := messcount(conn)
		if mserr != nil {
			log.Fatal("Getting count failed:", mserr.Error())
		}
		fmt.Printf("Count done\n")

		// Get Last 6 Messages
		scount, serr := summary(conn)
		if serr != nil {
			log.Fatal("Getting Summary failed:", mserr.Error())
		}
		fmt.Printf("Summary done %d \n", scount)

		ou.Mc = mscount
		// ou.Mess[1].Message = "Hello"
		// ou.Mess[2].Message = "twak"

		tmpl := template.Must(template.ParseFiles("region.html", "header.html", "headernav.html"))
		err = tmpl.Execute(w, ou)
		//err := tmpl.Execute(os.Stdout, data)
		if err != nil {
			panic(err)
		}
	case "POST":
		// Call ParseForm() to parse the raw query and update r.PostForm and r.Form.
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		//fmt.Fprintf(w, "Post from website! r.PostFrom = %v\n", r.PostForm)
		ou.OU = r.FormValue("sou")
		ou.ShMess = r.FormValue("shMess")

		MyShow := r.Form["shOptions"]
		fmt.Println("SA Selected : ", contains(MyShow, "SA"))

		//fmt.Println("shMess = ", ou.ShMess)
		// if ChkMess {
		// 	fmt.Println("ChkMess = true")
		// } else {
		// 	fmt.Println("ChkMess = false")
		// }

		fmt.Println("Name = ", ou.OU)
		fmt.Println("data posted")
		http.Redirect(w, r, "/region", http.StatusSeeOther)
	default:
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}

	fmt.Println("project executed")

}

func contains(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}
	_, ok := set[item]
	return ok
}

func cst(w http.ResponseWriter, r *http.Request) {

	tmpl := template.Must(template.ParseFiles("problemsl.html", "header.html", "headernav.html"))

	err := tmpl.Execute(w, nil)
	if err != nil {
		panic(err)
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	// data := struct {
	// 	Title  string
	// 	Header string
	// 	Uname  string
	// }{
	// 	Title:  "Index Page",
	// 	Header: "Hello, World!",
	// 	Uname:  "Deon",
	// }

	//Note use of ExecuteTemplate as apposed to just execute
	//err := tmpls.ExecuteTemplate(w, "index.html", data)
	//var tmpls = template.Must(template.ParseFiles("index.html", "headernav.html"))

	pagedata.Title = "Home"
	tmpl := template.Must(template.ParseFiles("index.html", "header.html", "headernav.html"))
	err := tmpl.Execute(w, pagedata)
	if err != nil {
		panic(err)
	}

	fmt.Println("index executed")
}

func about(w http.ResponseWriter, r *http.Request) {
	s2 := template.Must(template.ParseFiles("about.html", "header.html", "headernav.html"))
	s2.Execute(w, nil)
	fmt.Println("about executed")
}

func main() {
	ou.OU = "tt"

	router := mux.NewRouter()
	router.HandleFunc("/", index)
	router.HandleFunc("/about", about)
	router.HandleFunc("/test", testHandler)
	router.HandleFunc("/region", regiontHandler)
	router.HandleFunc("/cst", cst)

	router.PathPrefix("/css/").Handler(http.StripPrefix("/css/", http.FileServer(http.Dir("./css"))))
	router.PathPrefix("/js/").Handler(http.StripPrefix("/js/", http.FileServer(http.Dir("./js"))))

	http.Handle("/", router)

	//cssHandler := http.FileServer(http.Dir("./css/"))
	//imagesHandler := http.FileServer(http.Dir("./images/"))

	//http.Handle("/css/", http.StripPrefix("/css/", cssHandler))
	//http.Handle("/images/", http.StripPrefix("/images/", imagesHandler))

	port := ":3000"
	fmt.Println("Listening on localhost" + port)

	log.Fatal(http.ListenAndServe(port, router))
}
