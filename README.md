# Appointment Scheduling

## Duration
Around 60-90 minutes (please make sure to send across your code around the 90-minute mark at the most).

## Motivation
Clients need to be able to schedule an appointment with their trainer through an HTTP API.

## Instructions

The client should be able to pick from a list of available times. Appointments for a trainer should not overlap.
Appointments are 30 minutes long.
Appointments should be scheduled at :00, :30 minutes after the hour during business hours.
Business hours are M-F 8am-5pm Pacific Time

Your job is to create an HTTP JSON API written in Go with the following endpoints:

* Get a list of available appointment times for a trainer between two dates
  Parameters:
    trainer_id
    starts_at
    ends_at
  Returns:
    list of available appointment times
* Post an appointment (as JSON)
  Fields:
    trainer_id
    user_id
    starts_at
    ends_at
* Get a list of scheduled appointments for a trainer
  Parameters:
    trainer_id

appointments.json contains the current list of appointments in this format:

 [
	{
		"id": 1
		"trainer_id": 1
		"user_id": 2,
		"starts_at": "2019-01-25T09:00:00-08:00",
		"ends_at": "2019-01-25T09:30:00-08:00"
	}
]

You can store appointments in this file, a database or any back end storage you prefer.

## CURL commands
### Create new appointment
```shell
curl -X POST -H "Content-Type: application/json" -d '{"user_id": 1,"starts_at": "2020-01-26T11:40:00-08:00"}' "http://localhost:3000/trainers/1/appointments"
```

### Get list of scheduled appointments
```shell
curl -X GET "http://localhost:3000/trainers/1/appointments/scheduled"
```

### Get list of available appointments between 2 dates
```shell
curl -X GET -G \
'http://localhost:3000/trainers/2/appointments' \
-d starts_at=2020-01-24T08:00:00-08:00 \
-d ends_at=2020-01-24T23:00:00-08:00
```
