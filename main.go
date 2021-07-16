package main

import (
	"UntisAPI"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
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

const periodEventTypeLesson = 1
const periodEventTypeRecess = 2
const periodEventTypeDayBegin = 3
const periodEventTypeDayEnd = 4
const periodEventTypeMax = 4

type userConfig struct {
	USERNAME string
	PASSWORD string
	SCHOOL   string
	SERVER   string
}

type IperiodEvent interface {
	getType() int
}

type periodEvent struct {
	eventType int
	time      int
}

func (p periodEvent) getType() int {
	return p.eventType
}

type periodLessonEvent struct {
	periodEvent
	room      string
	teachers  string
	class     string
	startTime string
	endTime   string
}

var periodEvents map[int]map[int]IperiodEvent // [room id][untis time]

func main() {
	getConfiguration()
	getAPIData()
	getTimeTables()
	updateEvents()
	fmt.Printf("Initi done.\n")

	fmt.Print(getCurentEvent("2.312", 1200))
}

func getCurentEvent(room string, time int) (event IperiodEvent) {
	/*
		finds the newest event in the past.
	*/
	for t, period := range periodEvents[roomMapper[room]] {
		if t < time {
			event = period
		} else {
			break
		}
	}
	return event
}

func getConfiguration() {
	// read file
	data, err := ioutil.ReadFile("./config.json")
	if err != nil {
		fmt.Print(err)
	}
	// json data
	var config userConfig
	// unmarshall it
	err = json.Unmarshal(data, &config)
	if err != nil {
		fmt.Println("error:", err)
	}

	user = UntisAPI.NewUser(
		config.USERNAME,
		config.PASSWORD,
		config.SCHOOL,
		config.SERVER,
	)
}

func getAPIData() {
	/*
		login to Untis
	*/
	fmt.Printf("Logging in...")
	defer user.Logout()
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

}

func getTimeTables() {
	fmt.Printf("Logging in...")
	defer user.Logout()
	err := user.Login()
	if err != nil {
		fmt.Printf("\rLogin failed! error: %s\n", err.Error())
		return
	}
	fmt.Printf("\rLogged in!\n")

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
}

func updateEvents() {
	/*
		Create the lesson events
	*/
	periodEvents = map[int]map[int]IperiodEvent{}
	for name, id := range roomMapper {
		if id != -1 {
			periodEvents[id] = map[int]IperiodEvent{}
			for _, period := range periodsList[id] {
				event := periodLessonEvent{
					periodEvent: periodEvent{
						eventType: periodEventTypeLesson,
						time:      period.StartTime,
					},
					room:      name, // TODO how does untis save room changes
					startTime: strconv.Itoa(period.StartTime),
					endTime:   strconv.Itoa(period.EndTime),
				}
				for _, id := range period.Teacher { // adds all teachers to the event
					event.teachers += teachers[id].Name + "; "
				}
				for _, id := range period.Classes { // adds all classes to the event
					event.class += classes[id].Name + "; "
				}
				periodEvents[id][event.time] = event
			}
		}
	}
	/*
		Merge lessons so lessons are not split up in 45 min chunks
	*/
	for id, room := range periodEvents {
		var old = periodLessonEvent{}
		for _, event := range room {
			if old.getType() == periodEventTypeLesson {
				if event.getType() == periodEventTypeLesson {
					lessonEvent := event.(periodLessonEvent)
					// TODO use untis time grid to check time between events
					// checking if two lessons are the same
					if lessonEvent.room == old.room && lessonEvent.class == old.class && lessonEvent.teachers == old.teachers {
						delete(periodEvents[id], lessonEvent.time)
						old.endTime = lessonEvent.endTime
						periodEvents[id][old.time] = old
					} else {
						old = event.(periodLessonEvent)
					}
				}
			} else {
				old = event.(periodLessonEvent)
			}
		}
	}
	fmt.Printf("Created event list.\n")
}
