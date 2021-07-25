package future

import (
	"errors"
	"time"
)

// in memory copy of appointment data grouped by trainer
// we can do this bc our file is small - if this was coming from a db
// we could still hydrate it with some value on app load, but this wouldn't scale
// forever. Would then move to a model where we only hydrate this in memory map
// as items pass through the code, but for now this will do :)
//
// if we changed implementation away from a file to a db - we could theoretically
// keep most of the code in this file the same, and write similar methods on
// a new service, to keep up the abstraction.
var trainerAppointments = groupByTrainer(loadAppointments())

// Returns all scheduled appointments for the provided trainer
// returns an empty array if there are no appointments found for the trainer
func GetScheduledAppointmentsFor(trainerID int) []Appointment {
	appointments := trainerAppointments.GetAppointmentsFor(trainerID)
	if appointments == nil {
		return []Appointment{}
	}
	return appointments
}

// Returns all open appointment windows for the provided trainer
func GetAvailableAppointmentsFor(trainerID int, startDate, endDate time.Time) []AppointmentWindow {
	// convert times to UTC
	startDate = startDate.UTC()
	endDate = endDate.UTC()

	// set minutes to next 30 minute block 00 or 30
	adjustedStart := startDate
	if startDate.Minute()%30 != 0 {
		minuteBase := 30
		if startDate.Minute() > 30 {
			minuteBase = 60
		}
		adjustedStart = startDate.Add(time.Minute * time.Duration((minuteBase - startDate.Minute())))
	}

	adjustedEnd := endDate
	if endDate.Minute()%30 != 0 {
		minuteBase := 0
		if endDate.Minute() > 30 {
			minuteBase = 30
		}
		adjustedEnd = endDate.Add(-time.Minute * time.Duration(endDate.Minute()-minuteBase))
	}

	aptWindows := generateAppointmentWindowsBetween(adjustedStart, adjustedEnd, 30)

	whereAlreadyBooked := func(window AppointmentWindow) bool {
		return !trainerAppointments.TrainerFreeBetween(trainerID, window.StartsAt, window.EndsAt)
	}

	return discardWindows(whereAlreadyBooked, aptWindows)
}

func BookAppointmentFor(trainerID int, userID int, startTime time.Time) (*Appointment, error) {
	// convert start time to UTC for consistent server processing
	startTime = startTime.UTC()

	// what do we do if minute is not a clean 00 or 30? kick back an error? automatically round up for them?
	// I say kick back an error, lest we become an unpredictable black box - better to be explicit
	if startTime.Minute()%30 != 0 {
		return nil, errors.New("apt times can only start on the hour or half hour marks")
	}

	// apts must fall within operating hours
	// check for holidays?
	if !withinOperatingHours(startTime) {
		return nil, errors.New("time is outside of operating hours. Please select a time between 8am and 4:30pm PST M-F")
	}

	if !trainerAppointments.TrainerFreeBetween(trainerID, startTime, startTime.Add(time.Minute*30)) {
		return nil, errors.New("trainer is already booked for that time slot")
	}

	appointment := Appointment{
		ID:        trainerAppointments.GetNextID(),
		TrainerID: trainerID,
		UserID:    userID,
		StartsAt:  startTime,
		EndsAt:    startTime.Add(time.Minute * 30),
	}

	trainerAppointments.AddAppointment(appointment)
	err := trainerAppointments.Save()
	return &appointment, err
}
