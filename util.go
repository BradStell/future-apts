package future

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"time"
)

func loadAppointments() []Appointment {
	fileData, err := ioutil.ReadFile("appointments.json")
	if err != nil {
		log.Fatal(err)
	}

	appointments := []Appointment{}
	err = json.Unmarshal(fileData, &appointments)
	if err != nil {
		fmt.Printf("Error loading appointments from file, starting off with no data\n")
	}

	// convert all times to UTC for consistent data processing
	// TODO uncouple this from this function
	// create slice.map util or something
	for i := 0; i < len(appointments); i++ {
		appointments[i].StartsAt = appointments[i].StartsAt.UTC()
		appointments[i].EndsAt = appointments[i].EndsAt.UTC()
	}

	return appointments
}

func groupByTrainer(appointments []Appointment) TrainerAppointmentDictionary {
	trainerMap := make(TrainerAppointmentDictionary)

	for _, appointment := range appointments {
		trainerMap.AddAppointment(appointment)
	}

	return trainerMap
}

func discardWindows(where func(AppointmentWindow) bool, list []AppointmentWindow) []AppointmentWindow {
	windows := make([]AppointmentWindow, 0)
	for _, window := range list {
		if !where(window) {
			windows = append(windows, window)
		}
	}
	return windows
}

func generateAppointmentWindowsBetween(start, end time.Time, minutesLong int) []AppointmentWindow {
	windows := make([]AppointmentWindow, 0)
	current := start
	aptLength := time.Minute * time.Duration(minutesLong)

	for {
		if current.Equal(end) || current.After(end) || current.Add(aptLength).After(end) {
			// if we are here we have reached the end of our time window
			break
		} else if current.Before(end) && (current.Add(aptLength).Before(end) || current.Add(aptLength).Equal(end)) && withinOperatingHours(current) {
			// if we are here we should generate the next time slot
			windows = append(windows, AppointmentWindow{
				StartsAt: current,
				EndsAt:   current.Add(aptLength),
			})
		}
		current = current.Add(aptLength)
	}

	return windows
}

// Times all come in in UTC - we need to inspect it in PST due to
// local business operating times
func withinOperatingHours(t time.Time) bool {
	// Convert to PST
	pst, _ := time.LoadLocation("America/Los_Angeles")
	t = t.In(pst)

	if t.Weekday() == time.Sunday || t.Weekday() == time.Saturday {
		return false
	} else {
		if t.Hour() < 8 || t.Hour() >= 17 {
			return false
		}
	}
	return true
}
