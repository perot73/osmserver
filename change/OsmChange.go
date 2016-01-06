package main

import (
	"encoding/xml"
	//"github.com/jmoiron/sqlx"
	"fmt"
	_ "github.com/lib/pq"
	"time"
)

type OsmChange struct {
	XMLName   xml.Name `xml:"osmChange"`
	Version   string   `xml:"version,attr"`
	Generator string   `xml:"generator,attr"`
	Create    Creates  `xml:"create"`
	Modify    Modifies `xml:"modify"`
	Delete    Deletes  `xml:"delete"`
}

type Creates struct {
	Nodes     []Node     `xml:"node"`
	Ways      []Way      `xml:"way"`
	Relations []Relation `xml:"relation"`
}

type Modifies struct {
	Nodes     []Node     `xml:"node"`
	Ways      []Way      `xml:"way"`
	Relations []Relation `xml:"relation"`
}

type Deletes struct {
	Nodes     []Node     `xml:"node"`
	Ways      []Way      `xml:"way"`
	Relations []Relation `xml:"relation"`
}

type Node struct {
	Id        int64     `xml:"id,attr"`
	Version   int32     `xml:"version,attr"`
	Changeset int64     `xml:"changeset,attr"`
	Timestamp time.Time `xml:"timestamp,attr"`
	Lat       float32   `xml:"lat,attr"`
	Lon       float32   `xml:"lon,attr"`
	Tags      []Tag     `xml:"tag"`
}

type Way struct {
	Id        int64     `xml:"id,attr"`
	Version   int32     `xml:"version,attr"`
	Changeset int64     `xml:"changeset,attr"`
	Timestamp time.Time `xml:"timestamp,attr"`
	Nodes     []Nd      `xml:"nd"`
	Tags      []Tag     `xml:"tag"`
}

type Nd struct {
	Ref int64 `xml:"ref,attr"`
}

type Relation struct {
	Id        int64     `xml:"id,attr"`
	Version   int32     `xml:"version,attr"`
	Changeset int64     `xml:"changeset,attr"`
	Timestamp time.Time `xml:"timestamp,attr"`
	Members   []Member  `xml:"member"`
	Tags      []Tag     `xml:"tag"`
}

type Tag struct {
	Key int64  `xml:"k,attr"`
	Val string `xml:"v,attr"`
}

type Member struct {
	Type string `xml:"type,attr"`
	Ref  int64  `xml:"ref,attr"`
	Role string `xml:"role,attr"`
}

type IdCache struct {
	Node     map[int64]int64
	Way      map[int64]int64
	Relation map[int64]int64
}

func CreateNode(n Node, ids *IdCache) (int64, error) {
	var id int64
	fmt.Printf("\n insert into nodes (lat,lon, timestamp, version, changeset) values ($1,$2,$3,$4) returning id")
	for _, t := range n.Tags {
		fmt.Printf("\n insert into node_tags (node_id, k,v ) values (%v,%v,%v) returning id", n.Id, t.Key, t.Val)
	}
	//TODO: insert id to cache
	id = 1
	return id, nil
}

func CreateWay(w Way, ids *IdCache) (int64, error) {
	var id int64
	fmt.Printf("\n insert into way (timestamp, version, changeset) values ($1,$2,$3,$4,$5) returning id")
	for _, t := range w.Tags {
		fmt.Printf("\n insert into way_tags (way_id, k,v ) values (%v,%v,%v) returning id", w.Id, t.Key, t.Val)
	}
	for i, nd := range w.Nodes {
		//TODO: lookup id from cache
		fmt.Printf("\n insert into way_nodes (way_id, node_sequence_id) values (%v,%v,%v)", w.Id, nd.Ref, i)
	}

	return id, nil
}

func CreateRelation(r Relation, ids *IdCache) (int64, error) {
	var id int64
	fmt.Printf("\n insert into relation (timestamp, version, changeset) values ($1,$2,$3,$4,$5) returning id")
	for _, t := range r.Tags {
		fmt.Printf("\n insert into relation_tags (relation_id, k,v ) values (%v,%v,%v) returning id", r.Id, t.Key, t.Val)
	}
	for _, m := range r.Members {
		fmt.Printf("\n insert into relation_members (relation_id, member_type, member_id, member_role, version, sequence_id) values (%v,%v,%v)", r.Id, m.Ref, m.Role)
	}
	return id, nil
}

func ProcessCreates(c Creates, ids *IdCache) {

	for _, n := range c.Nodes {
		id, _ := CreateNode(n, ids)
		ids.Node[n.Id] = id
	}

	for _, w := range c.Ways {
		id, _ := CreateWay(w, ids)
		ids.Way[w.Id] = id
	}

	for _, r := range c.Relations {
		id, _ := CreateRelation(r, ids)
		ids.Relation[r.Id] = id
	}
	fmt.Printf("\nid:s %v", ids)
}

func ProcessModifies(m Modifies, ids *IdCache) {

}

func ProcessDeletes(d Deletes, ids *IdCache) {

}

func ProcessChange(oc OsmChange) {

	// open database

	// create node id cache
	ids := IdCache{make(map[int64]int64), make(map[int64]int64), make(map[int64]int64)}

	ProcessCreates(oc.Create, &ids)
	ProcessModifies(oc.Modify, &ids)
	ProcessDeletes(oc.Delete, &ids)

}

func main() {
	data := `<osmChange version="0.3" generator="Osmosis">
				<create>
				    <node id="-1" timestamp="2007-01-02T00:00:00.0+11:00" lat="-33.9133118622908" lon="151.117335519304">
				        <tag k="created_by" v="JOSM"/>
				    </node>
				    <node id="-2" timestamp="2007-01-02T00:00:00.0+11:00" lat="-33.9233118622908" lon="151.117335519304">
				        <tag k="created_by" v="JOSM"/>
				    </node>
				    <way id="-3" timestamp="2007-01-02T00:00:00.0+11:00">
				        <nd ref="-1"/>
				        <nd ref="-2"/>
				        <tag k="created_by" v="JOSM"/>
				    </way>
					<relation id="56688" user="kmvar" uid="56190" visible="true" version="28" changeset="6947637" timestamp="2011-01-12T14:23:49Z">
  						<member type="node" ref="294942404" role=""/>
  						<member type="node" ref="364933006" role=""/>
  						<tag k="name" v="Küstenbus Linie 123"/>
  					</relation>
				</create>
			</osmChange>
	`
	var cs OsmChange
	err := xml.Unmarshal([]byte(data), &cs)
	if err != nil {
		fmt.Printf("error: %v", err)
		return
	}
	//fmt.Printf("%v", cs)
	ProcessChange(cs)

}
