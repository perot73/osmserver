package repo

import (
	"encoding/xml"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"time"
)

type Osm struct {
	XMLName   xml.Name `xml:"osm"`
	Version   float32  `xml:"version,attr"`
	Generator string   `xml:"generator,attr"`
}

type OSMChangeset struct {
	Osm
	Changeset Changeset `xml:"changeset"`
}

type Tag struct {
	K string `xml:"k,attr"`
	V string `xml:"v,attr"`
}

type Changeset struct {
	XMLName   xml.Name  `xml:"changeset"`
	Id        int64     `xml:"id,attr"`
	Uid       int64     `xml:"uid,attr"`
	User      int64     `xml:"user,attr"`
	Minlon    float32   `xml:"min_lon,attr"`
	Minlat    float32   `xml:"min_lat,attr"`
	Maxlon    float32   `xml:"max_lon,attr"`
	Maxlat    float32   `xml:"max_lat,attr"`
	CreatedAt time.Time `xml:"created_at,attr"`
	Open      bool      `xml:"open,attr"`
	Tags      TagMap
}

type Entity struct {
	Id        int64     `xml:"id,attr"`
	User      string    `xml:"user,attr"`
	Uid       int32     `xml:"uid,attr"`
	Visible   bool      `xml:"visible,attr"`
	Version   int32     `xml:"version,attr"`
	Changeset int64     `xml:"changeset,attr"`
	Timestamp time.Time `xml:"timestamp,attr"`
	Tags      TagMap
}

type Node struct {
	Entity
	Lat float32 `xml:"lat,attr"`
	Lon float32 `xml:"lon,attr"`
}

type Way struct {
	Entity
	WayNodes WayNodes
}

type Nd struct {
	Ref int64 `xml:"Ref,attr"`
}

type Relation struct {
	Entity
	Members []Member `xml:"member"`
}

type Member struct {
	Type string `xml:"type,attr"`
	Ref  int64  `xml:"ref,attr"`
	Role string `xml:"role,attr"`
}

type User struct {
	XMLName     xml.Name `xml:"user"`
	Id          int64    `xml:"id,attr"`
	Email       string   `xml:"email,attr"`
	DisplayName string   `xml:"display_name,attr"`
}

/*
	<?xml version="1.0" encoding="UTF-8"?>
	 <osm version="0.6" generator="OpenStreetMap server">
	   <api>
	     <version minimum="0.6" maximum="0.6"/>
	     <area maximum="0.25"/>
	     <tracepoints per_page="5000"/>
	     <waynodes maximum="2000"/>
	     <changesets maximum_elements="50000"/>
	     <timeout seconds="300"/>
	     <status database="online" api="online" gpx="online"/>
	   </api>
 </osm>
*/

type Capabilities struct {
	Osm
	Api Api `xml:"api"`
}

type Api struct {
	Version Version `xml:"version"`
	Area    Area    `xml:"area"`
	Status  Status  `xml:"status"`
}

type Status struct {
	Database string `xml:"database,attr"`
	Api      string `xml:"api,attr"`
	Gpx      string `xml:"gpx,attr"`
}

type Area struct {
	Maximum float32 `xml:"maximum,attr"`
}

type Version struct {
	Min float32 `xml:"minimum,attr"`
	Max float32 `xml:"maximum,attr"`
}

type Repository struct {
	db *sqlx.DB
}

// NewRepository creates a new Repository
// using the supplied sqlx connection
func NewRepository(db *sqlx.DB) *Repository {
	repo := &Repository{db}
	return repo
}

func (repo *Repository) GetUser(email string) (User, error) {
	var user User
	sql := `SELECT 	id, email, display_name as displayname
			FROM	users
			WHERE	email = $1
	`
	err := repo.db.Get(&user, sql, email)
	if err != nil {
		return user, err
	}
	return user, nil
}

func (repo *Repository) GetNode(id int64) (Node, error) {
	var node Node
	sql := `SELECT id, version, user_id, tstamp, changeset_id, tags
			FROM nodes
			WHERE	id = $1
	`
	err := repo.db.Get(&node, sql, id)
	if err != nil {
		return node, err
	}

	return node, nil
}

func (repo *Repository) CreateChangeset(user User, changeset Changeset) (Changeset, error) {
	var cs Changeset
	sql := `
		insert into changeset (user_id , created_at, tags) 
		VALUES (?, ?, ?)
		RETURNING id;			
	`

	err := repo.db.Get(&cs.Id, sql, user.Id, changeset.Tags)

	if err != nil {
		return cs, err
	}

	return cs, nil

}
