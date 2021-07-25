package future

import (
	"encoding/json"
	"io/ioutil"
	"sort"
	"time"
)

type Appointment struct {
	ID        int       `json:"id"`
	TrainerID int       `json:"trainer_id"`
	UserID    int       `json:"user_id"`
	StartsAt  time.Time `json:"starts_at"`
	EndsAt    time.Time `json:"ends_at"`
}

type AppointmentWindow struct {
	StartsAt time.Time `json:"starts_at"`
	EndsAt   time.Time `json:"ends_at"`
}

// type alias so we can add methods to it
type TrainerAppointmentDictionary map[int][]Appointment

// Adds an appointment to the in memory dictionary of appointments
// then sorts the appointments by startDate
func (dict TrainerAppointmentDictionary) AddAppointment(apt Appointment) {
	if dict.TrainerExists(apt.TrainerID) {
		// add apointment to existing list
		dict[apt.TrainerID] = append(dict[apt.TrainerID], apt)
	} else {
		// create new appointment list with new appointment
		dict[apt.TrainerID] = []Appointment{apt}
	}

	// sort each trainers appointments
	// could always sort these conditionally based on prop?
	// or store them in a binary tree rather than an array
	// or not sort them at all if our dataset is huge
	// but to simplify online processing rather than this offline
	// processing - I'm going to pre sort the data so the GET request
	// speeds are prioritized
	for trainerID, apts := range dict {
		sort.Slice(apts, func(i, j int) bool {
			return apts[i].StartsAt.Before(apts[j].StartsAt)
		})
		dict[trainerID] = apts
	}
}

func (dict TrainerAppointmentDictionary) TrainerExists(trainerID int) bool {
	_, ok := dict[trainerID]
	return ok
}

func (dict TrainerAppointmentDictionary) GetAppointmentsFor(trainerID int) []Appointment {
	return dict[trainerID]
}

func (dict TrainerAppointmentDictionary) Save() error {
	// convert dict to slice of appointments
	apts := make([]Appointment, 0, len(dict))
	for _, trainerApts := range dict {
		apts = append(apts, trainerApts...)
	}

	// Convert back to PST before saving
	// wouldn't do this normally - but since the file came with PST times
	// I'll respect that for saving. Maybe a human is reading that file
	pst, _ := time.LoadLocation("America/Los_Angeles")
	for i := 0; i < len(apts); i++ {
		apts[i].StartsAt = apts[i].StartsAt.In(pst)
		apts[i].EndsAt = apts[i].EndsAt.In(pst)
	}

	// convert Appointment slice to json data as byte[]
	bytes, err := json.Marshal(apts)
	if err != nil {
		return err
	}

	// save byte slice to file
	err = ioutil.WriteFile("appointments.json", bytes, 0644)

	// either nil or an error
	return err
}

// Returns the next ID
// TODO - re-implement TrainerAppointmentDictionary to not be a map, but
// rather to contain a map, and also contain an int field which represents
// the next ID. It can be calculated on the fly during other method operations
// and can just be returned.
func (dict TrainerAppointmentDictionary) GetNextID() int {
	nextID := 0

	// go find the highest ID in our data
	for _, trainerAppointments := range dict {
		for _, appointment := range trainerAppointments {
			if appointment.ID > nextID {
				nextID = appointment.ID
			}
		}
	}

	return nextID + 1
}

func (dict TrainerAppointmentDictionary) TrainerFreeBetween(trainerID int, start, end time.Time) bool {
	trainerApts := dict.GetAppointmentsFor(trainerID)

	// if trainer doesn't have any appointments - then they are free
	if trainerApts == nil {
		return true
	}

	for _, apt := range trainerApts {
		switch {
		case start.Equal(apt.StartsAt):
			return false
		case start.Before(apt.StartsAt) && (end.Before(apt.StartsAt) || end.Equal(apt.StartsAt)):
			// valid
		case start.After(apt.StartsAt) && (start.After(apt.EndsAt) || start.Equal(apt.EndsAt)):
			// valid
		default:
			return false
		}
	}

	// we didn't fall into the false steps above so they must be free
	return true
}
