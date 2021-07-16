package main

import (
	"fmt"
)

func goThroughDay() {
	for i := 0; i < 2400; i++ {
		publishCurentEvents(i)
	}
}

func publishCurentEvents(time int) {
	for _, room := range periodEvents {
		if room[time] != nil {
			fmt.Printf("Time: %d \n", time)
			fmt.Println(room[time])
		}
	}
}
