package main

import (
	"controllers"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"gopkg.in/mgo.v2"
)

func main() {
	r := httprouter.New()

	uc := controllers.NewLocationCont(getSession())

	r.GET("/locations/:location_id", uc.GetLoc)

	r.GET("/trips/:trip_id", uc.GetTrip)

	r.POST("/locations", uc.CreateLoc)

	r.POST("/trips", uc.CreateTrip)

	r.PUT("/locations/:location_id", uc.UpdateLoc)

	r.PUT("/trips/:trip_id/request", uc.UpdateTrip)

	r.DELETE("/locations/:location_id", uc.RemoveLoc)

	http.ListenAndServe("localhost:8080", r)
}

func getSession() *mgo.Session {
	// Connect to mongo
	s, err := mgo.Dial("mongodb://admin:admin@ds045064.mongolab.com:45064/locations")
	if err != nil {
		panic(err)
	}
	s.SetMode(mgo.Monotonic, true)
	return s
}
