package main

import (
	//"bufio"
	"encoding/xml"
	//"errors"
	//"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/pat"
	"github.com/jmoiron/sqlx"
	"github.com/perot73/osmserver/repo"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
)

var (
	repository *repo.Repository
)

func GetNode(w http.ResponseWriter, req *http.Request) {
	id := req.URL.Query().Get(":id")
	log.Print("id: ", id)

	node, err := repository.GetNode(12)
	xml, err := xml.Marshal(node)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/xml")
	w.Write(xml)
}

func GetMap(w http.ResponseWriter, req *http.Request) {
	log.Println("getMap called")
	bbox := req.URL.Query().Get("bbox")
	log.Print("bbox: ", bbox)
	bounds, err := repo.NewBounds(bbox)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	osmmap, err := repository.GetMap(*bounds)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	xml, err := xml.Marshal(osmmap)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/xml")
	w.Write(xml)
}

func GetCapabilities(w http.ResponseWriter, req *http.Request) {
	data := `<?xml version="1.0" encoding="UTF-8"?>
	<osm version="0.6" generator="OpenStreetMap server" copyright="OpenStreetMap and contributors" attribution="http://www.openstreetmap.org/copyright" license="http://opendatacommons.org/licenses/odbl/1-0/">
	  <api>
	    <version minimum="0.6" maximum="0.6"/>
	    <area maximum="0.25"/>
	    <tracepoints per_page="5000"/>
	    <waynodes maximum="2000"/>
	    <changesets maximum_elements="50000"/>
	    <timeout seconds="300"/>
	    <status database="online" api="online" gpx="online"/>
	  </api>
	  <policy>
	    <imagery>
	      <blacklist regex=".*\.googleapis\.com/.*"/>
	      <blacklist regex=".*\.google\.com/.*"/>
	      <blacklist regex=".*\.google\.ru/.*"/>
	    </imagery>
	  </policy>
	</osm>`
	w.Header().Set("Content-Type", "text/xml")
	w.Write([]byte(data))
}

func GetUserDetails(w http.ResponseWriter, req *http.Request) {
	log.Println("GetUserDetails called")

	/*
		user, err := repository.GetUser("perot73@gmail.com")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		xml, err := xml.Marshal(user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("%v", xml)
	*/
	data := []byte(`<?xml version="1.0" encoding="UTF-8"?>
		<osm version="0.6" generator="OpenStreetMap server">
		  <user id="69054" display_name="Per-Olof Norén" account_created="2008-09-21T18:21:16Z">
		    <description></description>
		    <contributor-terms agreed="true" pd="false"/>
		    <img href="http://www.gravatar.com/avatar/6071b99b60343313beb3a89719bf5959.jpg"/>
		    <roles>
		    </roles>
		    <changesets count="0"/>
		    <traces count="0"/>
		    <blocks>
		      <received count="0" active="0"/>
		    </blocks>
		    <home lat="59.331915290443" lon="17.987415701444" zoom="3"/>
		    <languages>
		      <lang>sv-SE</lang>
		      <lang>sv</lang>
		      <lang>en-US</lang>
		      <lang>en</lang>
		    </languages>
		    <messages>
		      <received count="0" unread="0"/>
		      <sent count="0"/>
		    </messages>
		  </user>
		</osm>
		`)

	w.Header().Set("Content-Type", "text/xml")
	w.Write(data)
}

func CreateChangeset(w http.ResponseWriter, req *http.Request) {
	log.Println("CreateChangeset called")

	dump, err := httputil.DumpRequest(req, true)
	log.Printf("%s", dump)

	var cs repo.OSMChangeset

	err = xml.NewDecoder(req.Body).Decode(&cs)
	if err != nil {
		log.Panic("Unable to unmarshal api response", err)
	}
	log.Println(cs)
	//
	//xml, err := xml.Marshal(user)
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusInternalServerError)
	//	return
	//}
	//w.Header().Set("Content-Type", "text/xml")
	//w.Write(xml)
}

func CloseChangeset(w http.ResponseWriter, req *http.Request) {
	log.Println("CloseChangeset called")

	dump, err := httputil.DumpRequest(req, true)
	log.Printf("%s", dump)

	var cs repo.OSMChangeset

	err = xml.NewDecoder(req.Body).Decode(&cs)
	if err != nil {
		log.Panic("Unable to unmarshal api response", err)
	}
	log.Println(cs)
	//
	//xml, err := xml.Marshal(user)
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusInternalServerError)
	//	return
	//}
	//w.Header().Set("Content-Type", "text/xml")
	//w.Write(xml)
}

/*
func BasicAuth(pass handler) {

	return func(w http.ResponseWriter, r *http.Request) {

		auth := strings.SplitN(r.Header["Authorization"][0], " ", 2)

		if len(auth) != 2 || auth[0] != "Basic" {
			http.Error(w, "bad syntax", http.StatusBadRequest)
			return
		}

		payload, _ := base64.StdEncoding.DecodeString(auth[1])
		pair := strings.SplitN(string(payload), ":", 2)

		if len(pair) != 2 || !Validate(pair[0], pair[1]) {
			http.Error(w, "authorization failed", http.StatusUnauthorized)
			return
		}

		pass(w, r)
	}
}
*/
func main() {
	log.Println("Starting Server...")
	// Assign the repository to global variable
	sqlx, err := sqlx.Open("postgres", "host=127.0.0.1 dbname=osm sslmode=disable")
	if err != nil {
		log.Fatal("Unable to connect to database: ", err)
	}
	sqlx.Ping()
	repository = repo.NewRepository(sqlx)

	log.Println("Registering api endpoints...")

	m := pat.New()

	m.Get("/api/capabilities", GetCapabilities)
	m.Get("/api/0.6/user/details", GetUserDetails)

	m.Get("/api/0.6/map", GetMap)
	m.Get("/api/0.6/node/{id}", GetNode)
	m.Put("/api/0.6/changeset/create", CreateChangeset)
	m.Put("/api/0.6/changeset/close", CloseChangeset)

	/*
	   PUT /api/0.6/changeset/create
	   GET /api/0.6/changeset/#id?include_discussion=true
	   PUT /api/0.6/changeset/#id
	   PUT /api/0.6/changeset/#id/close
	   GET /api/0.6/changeset/#id/download
	   POST /api/0.6/changeset/#id/expand_bbox
	   GET /api/0.6/changesets
	   POST /api/0.6/changeset/#id/upload
	   PUT /api/0.6/[node|way|relation]/create
	   GET /api/0.6/[node|way|relation]/#id
	   PUT /api/0.6/[node|way|relation]/#id
	   DELETE /api/0.6/[node|way|relation]/#id
	   GET /api/0.6/[way|relation]/#id/full
	*/
	http.Handle("/", m)
	err = http.ListenAndServe(":8888", handlers.LoggingHandler(os.Stdout, m))
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
