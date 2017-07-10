package pagerduty

import "time"

/*
{
  "users": [
    {
      "name": "Alex Cunningham",
      "email": "acunningham@pagerduty.com",
      "id": "PGJ36Z3",
    }
  ],
  "query": "Cunningham",
  "limit": 25,
  "offset": 0,
  "total": null,
  "more": false
}*/

type PDUser struct {
	Name  string `json:"summary"`
	Email string
	ID    string `json:"id"`
}

type UsersReply struct {
	Users  []PDUser
	Limit  int
	Offset int
	More   bool
}

type Oncall struct {
	User  PDUser
	Start time.Time
	End   time.Time
}

type OncallsReply struct {
	Oncalls []Oncall
	Limit   int
	Offset  int
	More    bool
}
