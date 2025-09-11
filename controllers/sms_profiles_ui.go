package controllers

import (
	"net/http"
)

func (as *AdminServer) SMSProfilesPage(w http.ResponseWriter, r *http.Request) {
	params := newTemplateParams(r)
	params.Title = "SMS Profiles"
	getTemplate(w, "sms_profiles").ExecuteTemplate(w, "base", params)
}
