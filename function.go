package main

import "net/http"

func alreadyLoggedIn(req *http.Request) bool {
	c, err := req.Cookie("session")
	if err != nil {
		return false
	}
	un, oks := dbSession[c.Value]
	if !oks {
		return false
	}
	_, ok := dbUser[un]
	return ok
}

func getData(req *http.Request) user {
	cook, _ := req.Cookie("session")
	return dbUser[cook.Value]
}
