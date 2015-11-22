package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"uber"

	"github.com/julienschmidt/httprouter"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type LocationCont struct {
	session *mgo.Session
}

type Input struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	City    string `json:"city"`
	State   string `json:"state"`
	Zip     string `json:"zip"`
}

type Output struct {
	Id      bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	Name    string        `json:"name"`
	Address string        `json:"address"`
	City    string        `json:"city" `
	State   string        `json:"state"`
	Zip     string        `json:"zip"`

	Coordinate struct {
		Lat  string `json:"lat"`
		Lang string `json:"lang"`
	}
}

type mapsResponse struct {
	Results []mapsResult
}

type mapsResult struct {
	Address      string        `json:"formatted_address"`
	AddressParts []AddressPart `json:"address_components"`
	Geometry     Geometry
	Types        []string
}

type AddressPart struct {
	Name      string `json:"long_name"`
	ShortName string `json:"short_name"`
	Types     []string
}

func NewLocationCont(s *mgo.Session) *LocationCont {
	return &LocationCont{s}
}

func getGoogLocation(address string) Output {
	client := &http.Client{}

	reqURL := "http://maps.google.com/maps/api/geocode/json?address="
	reqURL += url.QueryEscape(address)
	reqURL += "&sensor=false"
	fmt.Println("URL: " + reqURL)
	req, err := http.NewRequest("GET", reqURL, nil)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("error", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error: ", err)
	}

	var res mapsResponse
	err = json.Unmarshal(body, &res)
	if err != nil {
		fmt.Println("error in unmashalling: ", err)
	}

	var ret Output
	ret.Coordinate.Lat = strconv.FormatFloat(res.Results[0].Geometry.Location.Lat, 'f', 7, 64)
	ret.Coordinate.Lang = strconv.FormatFloat(res.Results[0].Geometry.Location.Lng, 'f', 7, 64)

	return ret
}

type Geometry struct {
	Bounds   Bounds
	Location Point
	Type     string
	Viewport Bounds
}
type Bounds struct {
	NorthEast, SouthWest Point
}

type Point struct {
	Lat float64
	Lng float64
}

type TripInput struct {
	Startinglocationid string `json:"starting_from_location_id"`
	Locationids        []string
}

type TripOutput struct {
	Id                 bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	Status             string        `json:"status"`
	Startinglocationid string        `json:"startinglocation_id"`
	Bestlocation_ids   []string
	Totalubercosts     int     `json:"totalcosts"`
	Totaluberduration  int     `json:"totalduration"`
	Totaldistance      float64 `json:"totaldistance"`
}

type UberOutput struct {
	Cost     int
	Duration int
	Distance float64
}

type TripPutOutput struct {
	Id                 bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	Status             string        `json:"status"`
	Startinglocationid string        `json:"startinglocation_id"`
	Nextlocationid     string        `json:"nextlocationid"`
	Bestlocationids    []string
	Totalcosts         int     `json:"totalcosts"`
	Totalduration      int     `json:"totalduration"`
	Totaldistance      float64 `json:"totaldistance"`
	Uberwaittimeeta    int     `json:"uberwaiteta"`
}

type putStruct struct {
	trip_route  []string
	trip_visits map[string]int
}

type Final_struct struct {
	theMap map[string]Struct_for_put
}

// CreateLocation
func (uc LocationCont) CreateLoc(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var u Input
	var oA Output

	json.NewDecoder(r.Body).Decode(&u)
	googResCoor := getGoogLocation(u.Address + "+" + u.City + "+" + u.State + "+" + u.Zip)
	fmt.Println("resp is: ", googResCoor.Coordinate.Lat, googResCoor.Coordinate.Lang)

	oA.Id = bson.NewObjectId()
	oA.Name = u.Name
	oA.Address = u.Address
	oA.City = u.City
	oA.State = u.State
	oA.Zip = u.Zip
	oA.Coordinate.Lat = googResCoor.Coordinate.Lat
	oA.Coordinate.Lang = googResCoor.Coordinate.Lang

	// Write the user to mongo
	uc.session.DB("locations").C("locationA").Insert(oA)

	uj, _ := json.Marshal(oA)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	fmt.Fprintf(w, "%s", uj)
}

// GetLocation
func (uc LocationCont) GetLoc(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("location_id")
	if !bson.IsObjectIdHex(id) {
		w.WriteHeader(404)
		return
	}

	oid := bson.ObjectIdHex(id)
	var o Output
	if err := uc.session.DB("locations").C("locationA").FindId(oid).One(&o); err != nil {
		w.WriteHeader(404)
		return
	}
	uj, _ := json.Marshal(o)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	fmt.Fprintf(w, "%s", uj)
}

// RemoveLocation
func (uc LocationCont) RemoveLoc(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("location_id")

	if !bson.IsObjectIdHex(id) {
		w.WriteHeader(404)
		return
	}
	// get id
	oid := bson.ObjectIdHex(id)

	// Remove user
	if err := uc.session.DB("locations").C("locationA").RemoveId(oid); err != nil {
		w.WriteHeader(404)
		return
	}

	w.WriteHeader(200)
}

//UpdateLocation
func (uc LocationCont) UpdateLoc(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var i Input
	var o Output

	id := p.ByName("location_id")
	if !bson.IsObjectIdHex(id) {
		w.WriteHeader(404)
		return
	}
	oid := bson.ObjectIdHex(id)

	if err := uc.session.DB("locations").C("locationA").FindId(oid).One(&o); err != nil {
		w.WriteHeader(404)
		return
	}

	json.NewDecoder(r.Body).Decode(&i)
	googResCoor := getGoogLocation(i.Address + "+" + i.City + "+" + i.State + "+" + i.Zip)
	fmt.Println("resp is: ", googResCoor.Coordinate.Lat, googResCoor.Coordinate.Lang)

	o.Address = i.Address
	o.City = i.City
	o.State = i.State
	o.Zip = i.Zip
	o.Coordinate.Lat = googResCoor.Coordinate.Lat
	o.Coordinate.Lang = googResCoor.Coordinate.Lang

	// Write the user to mongo
	c := uc.session.DB("locations").C("locationA")

	id2 := bson.M{"_id": oid}
	err := c.Update(id2, o)
	if err != nil {
		panic(err)
	}

	uj, _ := json.Marshal(o)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	fmt.Fprintf(w, "%s", uj)
}

// GetTrip
func (uc LocationCont) GetTrip(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("trip_id")
	if !bson.IsObjectIdHex(id) {
		w.WriteHeader(404)
		return
	}

	oid := bson.ObjectIdHex(id)
	var tO TripPostOutput
	if err := uc.session.DB("locations").C("trips").FindId(oid).One(&tO); err != nil {
		w.WriteHeader(404)
		return
	}
	uj, _ := json.Marshal(tO)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	fmt.Fprintf(w, "%s", uj)
}

// CreateTrip
func (uc LocationCont) CreateTrip(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var tI PostInput
	var tO PostOutput
	var cost_array []int
	var duration_array []int
	var distance_array []float64
	costtotal := 0
	durationtotal := 0
	distancetotal := 0.0

	json.NewDecoder(r.Body).Decode(&tI)

	startingid := bson.ObjectIdHex(tI.Starting_from_location_id)
	var start Output
	if err := uc.session.DB("locations").C("LocationA").FindId(starting_id).One(&start); err != nil {
		w.WriteHeader(404)
		return
	}
	start_Lat := start.Coordinate.Lat
	start_Lang := start.Coordinate.Lang

	for len(tI.Location_ids) > 0 {

		for _, loc := range tI.Location_ids {
			id := bson.ObjectIdHex(loc)
			var o Output
			if err := uc.session.DB("locations").C("LocationA").FindId(id).One(&o); err != nil {
				w.WriteHeader(404)
				return
			}
			loc_Lat := o.Coordinate.Lat
			loc_Lang := o.Coordinate.Lang

			getUberResponse := uber.GetUberPrice(start_Lat, start_Lang, loc_Lat, loc_Lang)
			fmt.Println("Response: ", getUberResponse.Cost, getUberResponse.Duration, getUberResponse.Distance)
			cost_array = append(cost_array, getUberResponse.Cost)
			duration_array = append(duration_array, getUberResponse.Duration)
			distance_array = append(distance_array, getUberResponse.Distance)

		}
		fmt.Println("Cost", cost_array)

		min_cost := cost_array[0]
		var indexNeeded int
		for index, value := range cost_array {
			if value < min_cost {
				min_cost = value
				indexNeeded = index
			}
		}

		cost_total += min_cost
		duration_total += duration_array[indexNeeded]
		distance_total += distance_array[indexNeeded]

		tO.Best_route_location_ids = append(tO.Best_route_location_ids, tI.Location_ids[indexNeeded])

		starting_id = bson.ObjectIdHex(tI.Location_ids[indexNeeded])
		if err := uc.session.DB("locations").C("LocationA").FindId(starting_id).One(&start); err != nil {
			w.WriteHeader(404)
			return
		}
		tI.Location_ids = append(tI.Location_ids[:indexNeeded], tI.Location_ids[indexNeeded+1:]...)

		start_Lat = start.Coordinate.Lat
		start_Lang = start.Coordinate.Lang

		cost_array = cost_array[:0]
		duration_array = duration_array[:0]
		distance_array = distance_array[:0]
	}

	Last_loc_id := bson.ObjectIdHex(tO.Best_route_location_ids[len(tO.Best_route_location_ids)-1])
	var o2 Output
	if err := uc.session.DB("locations").C("LocationA").FindId(Last_loc_id).One(&o2); err != nil {
		w.WriteHeader(404)
		return
	}
	last_loc_Lat := o2.Coordinate.Lat
	last_loc_Lang := o2.Coordinate.Lang

	ending_id := bson.ObjectIdHex(tI.Starting_from_location_id)
	var end Output
	if err := uc.session.DB("locations").C("LocationA").FindId(ending_id).One(&end); err != nil {
		w.WriteHeader(404)
		return
	}
	end_Lat := end.Coordinate.Lat
	end_Lang := end.Coordinate.Lang

	getUberResponse_last := uber.GetUberPrices(last_loc_Lat, last_loc_Lang, end_Lat, end_Lang)

	tO.Id = bson.NewObjectId()
	tO.Status = "planning"
	tO.Starting_from_location_id = tI.Starting_from_location_id
	tO.Total_uber_costs = cost_total + getUberResponse_last.Cost
	tO.Total_distance = distance_total + getUberResponse_last.Distance
	tO.Total_uber_duration = duration_total + getUberResponse_last.Duration

	// Write the user to mongo
	uc.session.DB("locations").C("trips").Insert(tO)

	uj, _ := json.Marshal(tO)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	fmt.Fprintf(w, "%s", uj)
}

type live_data struct {
	Id             string   `json:"_id" bson:"_id,omitempty"`
	Tripvisited    []string `json:"tripvisited"`
	Tripnotvisited []string `json:"tripnotvisited"`
	Tripcompleted  int      `json:"tripcompleted"`
}

func (uc LocationCont) UpdateTrip(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	var theStruct putStruct
	var final Finalstruct
	final.theMap = make(map[string]Struct_for_put)

	var tPO TripPutOutput
	var internal Internaldata

	id := p[0].Value
	if !bson.IsObjectIdHex(id) {
		w.WriteHeader(404)
		return
	}
	oid := bson.ObjectIdHex(id)
	if err := uc.session.DB("locations").C("Trips").FindId(oid).One(&tPO); err != nil {
		w.WriteHeader(404)
		return
	}

	theStruct.trip_route = tPO.Best_route_location_ids
	theStruct.trip_route = append([]string{tPO.Startinglocationid}, theStruct.trip_route...)
	fmt.Println("route  is: ", theStruct.trip_route)
	theStruct.trip_visits = make(map[string]int)

	var tripvisited []string
	var tripnotvisited []string

	if err := uc.session.DB("locations").C("err_data").FindId(id).One(&internal); err != nil {
		for index, loc := range theStruct.trip_route {
			if index == 0 {
				theStruct.trip_visits[loc] = 1
				trip_visited = append(trip_visited, loc)
			} else {
				theStruct.trip_visits[loc] = 0
				trip_not_visited = append(tripnotvisited, loc)
			}
		}
		internal.Id = id
		internal.Tripvisited = tripvisited
		internal.Tripnotvisited = tripnotvisited
		internal.Tripcompleted = 0
		uc.session.DB("locations").C("err_data").Insert(internal)

	} else {
		for _, loc_id := range internal.Trip_visited {
			theStruct.trip_visits[loc_id] = 1
		}
		for _, loc_id := range internal.Trip_not_visited {
			theStruct.trip_visits[loc_id] = 0
		}
	}

	fmt.Println("Trip visit route ", theStruct.trip_visits)
	final.theMap[id] = theStruct

	last_index := len(theStruct.triproute) - 1
	trip_completed := internal.Tripcompleted
	if trip_completed == 1 {

		tPO.Status = "completed"

		uj, _ := json.Marshal(tPO)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		fmt.Fprintf(w, "%s", uj)
		return
	}

	for i, location := range theStruct.trip_route {
		if theStruct.trip_visits[location] == 0 {
			tPO.Nextlocationid = location
			nextoid := bson.ObjectIdHex(location)
			var o Output
			if err := uc.session.DB("locations").C("LocationA").FindId(nextoid).One(&o); err != nil {
				w.WriteHeader(404)
				return
			}
			nlat := o.Coordinate.Lat
			nlang := o.Coordinate.Lang

			if i == 0 {
				starting_point := theStruct.trip_route[last_index]
				startingoid := bson.ObjectIdHex(starting_point)
				var o Output
				if err := uc.session.DB("locations").C("LocationA").FindId(startingoid).One(&o); err != nil {
					w.WriteHeader(404)
					return
				}
				slat := o.Coordinate.Lat
				slang := o.Coordinate.Lang

				eta := uber.GetUberEta(slat, slang, nlat, nlang)
				tPO.Uber_wait_time_eta = eta
				trip_completed = 1
			} else {
				starting_point2 := theStruct.trip_route[i-1]
				startingoid2 := bson.ObjectIdHex(starting_point2)
				var o Output
				if err := uc.session.DB("locations").C("LocationA").FindId(startingoid2).One(&o); err != nil {
					w.WriteHeader(404)
					return
				}
				slat := o.Coordinate.Lat
				slang := o.Coordinate.Lang
				eta := uber.GetUberEta(slat, slang, nlat, nlang)
				tPO.Uber_wait_time_eta = eta
			}

			fmt.Println("Starting Location: ", tPO.Starting_from_location_id)
			fmt.Println("Next location: ", tPO.Next_destination_location_id)
			theStruct.trip_visits[location] = 1
			if i == last_index {
				theStruct.trip_visits[theStruct.trip_route[0]] = 0
			}
			break
		}
	}

	tripvisited = tripvisited[:0]
	tripnotvisited = tripnotvisited[:0]
	for location, visit := range theStruct.trip_visits {
		if visit == 1 {
			tripvisited = append(tripvisited, location)
		} else {
			tripnotvisited = append(tripnotvisited, location)
		}
	}

	internal.Id = id
	internal.Tripvisited = tripvisited
	internal.Tripnotvisited = tripnotvisited
	fmt.Println("Trip Visisted", internal.Trip_visited)
	fmt.Println("Trip Not Visisted", internal.Tripnotvisited)
	internal.Trip_completed = trip_completed

	c := uc.session.DB("locations").C("err_data")
	id2 := bson.M{"_id": id}
	err := c.Update(id2, internal)
	if err != nil {
		panic(err)
	}

	uj, _ := json.Marshal(tPO)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	fmt.Fprintf(w, "%s", uj)

}
