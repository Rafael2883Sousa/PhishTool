package controllers

import "net/http"

func (as *AdminServer) SMSTemplatesPage(w http.ResponseWriter, r *http.Request) {
	p := newTemplateParams(r)
	p.Title = "SMS Templates"
	if err := getTemplate(w, "sms_templates").ExecuteTemplate(w, "base", p); err != nil {
		http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
	}
}
