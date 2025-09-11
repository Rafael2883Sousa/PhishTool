package controllers

import (
	"net/http"
)

func (as *AdminServer) SMSProfilesPage(w http.ResponseWriter, r *http.Request) {
	params := newTemplateParams(r)
	params.Title = "SMS Profiles"
	if err := getTemplate(w, "sms_profiles").ExecuteTemplate(w, "base", params); err != nil {
		http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
