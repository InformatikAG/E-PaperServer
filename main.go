package main

import (
	"UntisAPI"
	"fmt"
	"time"
)

var roomNames = []string{
	"2.306",
	"2.312",
}

var user *UntisAPI.User                         // user information to login to the UntisAPI
var rooms map[int]UntisAPI.Room                 // maps room id to a Untis Room object
var teachers map[int]UntisAPI.Teacher           // maps teachre id to Untis Teacher object
var classes map[int]UntisAPI.Class              // maps classe id to Untis class object
var subjects map[int]UntisAPI.Subject           // maps subject id to Untis subject object
var periodsList map[int]map[int]UntisAPI.Period // room id to map of periods
var roomMapper map[string]int                   // maps room names to room ids

const periodEventTypeLessonBegin = 1
const periodEventTypeLessonEnd = 2
const periodEventTypeDayBegin = 3
const periodEventTypeDayEnd = 4
const periodEventTypeMax = 4

type periodEvent struct {
	eventType int
	time      int
}

var periodEvents map[int]map[int]periodEvent // [room id][untis time]

func main() {
	/*
		login to Untis
	*/
	user = UntisAPI.NewUser(
		//"maarten8",
		"niklas351",
		//"behn500",
		"C1oben",
		"TBZ Mitte Bremen",
		"https://tipo.webuntis.com")

	fmt.Printf("Logging in...")
	//time.Sleep(time.Second)
	err := user.Login()
	if err != nil {
		fmt.Printf("\rLogin failed! error: %s\n", err.Error())
		return
	}
	fmt.Printf("\rLogged in!\n")

	/*
		saves basic information about rooms into rooms map
	*/
	fmt.Printf("Loading rooms...")
	rooms, err = user.GetRooms()
	if err != nil {
		fmt.Printf("\rLoading rooms failed! error: %s\n", err.Error())
		return
	}
	fmt.Printf("\rLoaded rooms.\n")

	/*
		saves basic information about teachers into teachers map
	*/
	fmt.Printf("Loading teachers...")
	teachers, err = user.GetTeachers()
	if err != nil {
		fmt.Printf("\rLoading teachers failed! error: %s\n", err.Error())
		fmt.Printf("Skipping teachers.\n")
	} else {
		fmt.Printf("\rLoaded teachers.\n")
	}

	/*
		saves basic information about classes into classes map
	*/
	fmt.Printf("Loading classes...")
	classes, err = user.GetClasses()
	if err != nil {
		fmt.Printf("\rLoading classes failed! error: %s\n", err.Error())
		fmt.Printf("Skipping classes.\n")
	} else {
		fmt.Printf("\rLoaded classes.\n")
	}

	/*
		saves basic information about subjects into subjects map
	*/
	fmt.Printf("Loading subjects...")
	subjects, err = user.GetSubjectes()
	if err != nil {
		fmt.Printf("\rLoading subjects failed! error: %s\n", err.Error())
		fmt.Printf("Skipping subjects.\n")
	} else {
		fmt.Printf("\rLoaded subjects.\n")
	}

	fmt.Printf("Mapping rooms...\r")
	roomMapper = map[string]int{}
	for _, usedRoom := range roomNames {
		found := false
		var room UntisAPI.Room
		for i, _ := range rooms {
			if rooms[i].Name == usedRoom {
				found = true
				room = rooms[i]
			}
		}
		if found {
			roomMapper[usedRoom] = room.Id
			fmt.Printf("Room %s has id %d \n", room.Name, room.Id)
		} else {
			roomMapper[usedRoom] = -1
			fmt.Printf("Room %s not found!\nSkipping room.\n", usedRoom)
		}
	}

	/*
	   saves the periods of the current day of every room into the periods List map
	*/
	fmt.Printf("Loading date...")
	date := UntisAPI.ToUntisDate(time.Now())
	fmt.Printf("\rToday is the: %d\n", date)

	fmt.Printf("Loading periods...\r")
	periodsList = map[int]map[int]UntisAPI.Period{}
	for _, id := range roomMapper {
		if id != -1 {
			periodsList[id], err = user.GetTimeTable(id, 4, date, date)
			if err != nil {
				fmt.Printf("Loading periodsList of roomId %d\n failed! error: %s\n", id, err.Error())
			}
			fmt.Printf("Loading periodsList of roomId %d\n", id)
		}
	}

	/*
		fmt.Printf("Creating events...\r")
		for id, _ := range periodsList {
			currentTime := 0
		}
	*/

	fmt.Printf("Initi done.\n")

}

func getCurentHour(room int) UntisAPI.Period {
	for _, period := range periodsList[room] {
		if period.StartTime < UntisAPI.ToUnitsTime(time.Now()) && period.EndTime > UntisAPI.ToUnitsTime(time.Now()) {
			return nil, period
		}
	}
	return
}
