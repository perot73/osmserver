package repo

import (
	"encoding/xml"
	//"github.com/jmoiron/sqlx"
	"github.com/lib/pq/hstore"
	//"time"
	//"fmt"
	"errors"
	"log"
	"strconv"
	"strings"
)

type OSMMap struct {
	Osm
	Bounds    Bounds     `xml:"bounds"`
	Nodes     []Node     `xml:"node"`
	Ways      []Way      `xml:"way"`
	Relations []Relation `xml:"relation"`
}

type TagMap struct {
	hstore.Hstore
}

func (t *TagMap) MarshalXML(e *xml.Encoder, start xml.StartElement) error {

	for k, v := range t.Map {
		e.EncodeToken(
			xml.StartElement{xml.Name{Local: "tag"}, []xml.Attr{
				xml.Attr{Name: xml.Name{Local: "k"}, Value: k},
				xml.Attr{Name: xml.Name{Local: "v"}, Value: v.String},
			},
			},
		)
		e.EncodeToken(
			xml.EndElement{xml.Name{Local: "tag"}},
		)
	}
	return nil
}

type WayNodes []int64

func (wn *WayNodes) Scan(src interface{}) error {
	asBytes, ok := src.([]byte)
	if !ok {
		return error(errors.New("Scan source was not []byte"))
	}
	asString := string(asBytes)
	(*wn) = strToIntSlice(asString)
	return nil
}

func strToIntSlice(s string) []int64 {
	r := strings.Trim(s, "{}")
	a := make([]int64, 0, 10)
	for _, t := range strings.Split(r, ",") {
		i, _ := strconv.ParseInt(t, 10, 64)
		a = append(a, i)
	}
	return a
}

func (wn *WayNodes) MarshalXML(e *xml.Encoder, start xml.StartElement) error {

	for _, v := range *wn {
		e.EncodeToken(
			xml.StartElement{xml.Name{Local: "nd"}, []xml.Attr{
				xml.Attr{Name: xml.Name{Local: "ref"}, Value: strconv.FormatInt(v, 10)},
			},
			},
		)
		e.EncodeToken(
			xml.EndElement{xml.Name{Local: "nd"}},
		)
	}
	return nil
}

func (repo *Repository) GetMap(bounds Bounds) (OSMMap, error) {
	var result OSMMap

	boundsSQL := `
	with
		boxnodes as (
			select id from nodes where st_intersects(geom, st_makeenvelope($1, $2, $3, $4, 4326)) 
		),
    	boxways as (
    		select distinct way_id from way_nodes where node_id in (select id from boxnodes)
    	),
    	waynodes as (
    		select node_id from way_nodes where way_id in (select way_id from boxways)
    	),
		extents as (
			select st_envelope(st_extent(geom)) as geom
			from 	nodes 
    		where	id in ( select node_id from waynodes)
		)
		select 	st_xmin(geom) as minlon,
				st_ymin(geom) as minlat,
				st_xmax(geom) as maxlon,
				st_ymax(geom) as maxlat
		from extents    	
	`
	err := repo.db.Get(&result.Bounds, boundsSQL, bounds.Minlon, bounds.Minlat, bounds.Maxlon, bounds.Maxlat)

	if err != nil {
		return result, err
	}
	log.Println("GetMap: Bounds fetched...")

	nodeSQL := `
		with
		boxnodes as (
			select id from nodes where st_intersects(geom, st_makeenvelope($1, $2, $3, $4, 4326)) 
		),
    	boxways as (
    		select distinct way_id from way_nodes where node_id in (select id from boxnodes)
    	),
    	waynodes as (
    		select node_id from way_nodes where way_id in (select way_id from boxways)
    	)
    	select 	id,
	           	version,
	           	changeset_id as changeset,
	           	tstamp as timestamp,
	           	user_id as uid,
	           	'' as user,
	           	1 as visible,
	           	tags, 
	           	st_x(geom) as lon, 
	           	st_y(geom) as lat
    	from 	nodes 
    	where	id in ( 
    				select node_id from waynodes 
    				union select id from boxnodes
    			)
		`

	err = repo.db.Select(&result.Nodes, nodeSQL, bounds.Minlon, bounds.Minlat, bounds.Maxlon, bounds.Maxlat)

	if err != nil {
		return result, err
	}
	log.Println("GetMap: Nodes fetched...")

	waySQL := `
   	 	with
           boxnodes as (
           	select id from nodes where st_intersects(geom, st_makeenvelope($1, $2, $3, $4, 4326))
           ),
           boxways as (
           	select distinct way_id from way_nodes where node_id in (select id from boxnodes)
           )
           select 	id,
           		version,
           		changeset_id as changeset,
           		tstamp as timestamp,
           		user_id as uid,
           		user_id as user,
           		1 as visible,
           		tags,
           	   	(
           	   		select 	array_agg(node_id::bigint)
           	   	  	from (
           	   	  		select 	node_id
           	   	  		from 	way_nodes 
           	   	  		where 	way_id = id 
           	   	  		order by sequence_id
           	   	  	) as foo

           	   	) as waynodes
           from 	ways
           where 	id in ( select way_id from boxways)
		`
	err = repo.db.Select(&result.Ways, waySQL, bounds.Minlon, bounds.Minlat, bounds.Maxlon, bounds.Maxlat)
	if err != nil {
		return result, err
	}

	log.Println("GetMap: Ways fetched...")
	result.Generator = "golang OSM Server"
	//fmt.Printf("nodes: %v", nodes)
	//result.Nodes = nodes
	result.Version = 0.6
	return result, nil
}
