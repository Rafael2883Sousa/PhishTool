package api

import (
	"net/http"
	"log"

	mid "github.com/gophish/gophish/middleware"
	"github.com/gophish/gophish/middleware/ratelimit"
	"github.com/gophish/gophish/models"
	"github.com/gophish/gophish/worker"
	"github.com/gorilla/mux"
)

// ServerOption is an option to apply to the API server.
type ServerOption func(*Server)

// Server represents the routes and functionality of the Gophish API.
// It's not a server in the traditional sense, in that it isn't started and
// stopped. Rather, it's meant to be used as an http.Handler in the
// AdminServer.
type Server struct {
	handler http.Handler
	worker  worker.Worker
	limiter *ratelimit.PostLimiter
}

// NewServer returns a new instance of the API handler with the provided
// options applied.
func NewServer(options ...ServerOption) *Server {
	defaultWorker, _ := worker.New()
	defaultLimiter := ratelimit.NewPostLimiter()
	as := &Server{
		worker:  defaultWorker,
		limiter: defaultLimiter,
	}
	for _, opt := range options {
		opt(as)
	}
	as.registerRoutes()
	return as
}

// WithWorker is an option that sets the background worker.
func WithWorker(w worker.Worker) ServerOption {
	return func(as *Server) {
		as.worker = w
	}
}

func WithLimiter(limiter *ratelimit.PostLimiter) ServerOption {
	return func(as *Server) {
		as.limiter = limiter
	}
}

func (as *Server) registerRoutes() {
	root := mux.NewRouter()
	root = root.StrictSlash(true)
	router := root.PathPrefix("/api/").Subrouter()
	router.Use(mid.RequireAPIKey)
	router.Use(mid.EnforceViewOnly)
	router.HandleFunc("/imap/", as.IMAPServer)
	router.HandleFunc("/imap/validate", as.IMAPServerValidate)
	router.HandleFunc("/reset", as.Reset)
	router.HandleFunc("/campaigns/", as.Campaigns)
	router.HandleFunc("/campaigns/summary", as.CampaignsSummary)
	router.HandleFunc("/campaigns/{id:[0-9]+}", as.Campaign)
	router.HandleFunc("/campaigns/{id:[0-9]+}/results", as.CampaignResults)
	router.HandleFunc("/campaigns/{id:[0-9]+}/summary", as.CampaignSummary)
	router.HandleFunc("/campaigns/{id:[0-9]+}/complete", as.CampaignComplete)
	router.HandleFunc("/groups/", as.Groups)
	router.HandleFunc("/groups/summary", as.GroupsSummary)
	router.HandleFunc("/groups/{id:[0-9]+}", as.Group)
	router.HandleFunc("/groups/{id:[0-9]+}/summary", as.GroupSummary)
	router.HandleFunc("/templates/", as.Templates)
	router.HandleFunc("/templates/{id:[0-9]+}", as.Template)
	router.HandleFunc("/pages/", as.Pages)
	router.HandleFunc("/pages/{id:[0-9]+}", as.Page)
	router.HandleFunc("/smtp/", as.SendingProfiles)
	router.HandleFunc("/smtp/{id:[0-9]+}", as.SendingProfile)
	router.HandleFunc("/users/", mid.Use(as.Users, mid.RequirePermission(models.PermissionModifySystem)))
	router.HandleFunc("/users/{id:[0-9]+}", mid.Use(as.User))
	router.HandleFunc("/util/send_test_email", as.SendTestEmail)
	router.HandleFunc("/import/group", as.ImportGroup)
	router.HandleFunc("/import/email", as.ImportEmail)
	router.HandleFunc("/import/site", as.ImportSite)
	router.HandleFunc("/webhooks/", mid.Use(as.Webhooks, mid.RequirePermission(models.PermissionModifySystem)))
	router.HandleFunc("/webhooks/{id:[0-9]+}/validate", mid.Use(as.ValidateWebhook, mid.RequirePermission(models.PermissionModifySystem)))
	router.HandleFunc("/webhooks/{id:[0-9]+}", mid.Use(as.Webhook, mid.RequirePermission(models.PermissionModifySystem)))

	// SMS Profiles (API)
	router.HandleFunc("/sms/profiles/", as.GetSMSProfiles).Methods("GET")
	router.HandleFunc("/sms/profiles/", as.CreateSMSProfile).Methods("POST")
	router.HandleFunc("/sms/profiles/{id:[0-9]+}", as.UpdateSMSProfile).Methods("PUT")
	router.HandleFunc("/sms/profiles/{id:[0-9]+}", as.DeleteSMSProfile).Methods("DELETE")

	//SMS Tamplates
	router.HandleFunc("/sms/templates/", as.GetSMSTemplates).Methods("GET")
	router.HandleFunc("/sms/templates/", as.CreateSMSTemplate).Methods("POST")
	router.HandleFunc("/sms/templates/{id:[0-9]+}", as.UpdateSMSTemplate).Methods("PUT")
	router.HandleFunc("/sms/templates/{id:[0-9]+}", as.DeleteSMSTemplate).Methods("DELETE")

	// ==== DB diagnostics (path + columns) ====
	type dbrow struct{ Seq int; Name, File string }
	var dbl []dbrow
	if err := models.DB().Raw("PRAGMA database_list;").Scan(&dbl).Error; err != nil {
		log.Printf("PRAGMA database_list error: %v", err)
	} else if len(dbl) > 0 {
		log.Printf("SQLite file in use: %s", dbl[0].File)
	}

	type colInfo struct{ Name string `gorm:"column:name"` }
	var cols []colInfo
	var dbres = models.DB().Raw("PRAGMA table_info(sms_profiles);").Scan(&cols)
	if dbres.Error == nil {
		names := make([]string, 0, len(cols))
		for _, c := range cols { names = append(names, c.Name) }
		log.Printf("sms_profiles columns: %v", names)
	} else {
		log.Printf("PRAGMA table_info(sms_profiles) error: %v", dbres.Error)
	}

	cols = nil
	dbres = models.DB().Raw("PRAGMA table_info(sms_templates);").Scan(&cols)
	if dbres.Error == nil {
		names := make([]string, 0, len(cols))
		for _, c := range cols { names = append(names, c.Name) }
		log.Printf("sms_templates columns: %v", names)
	} else {
		log.Printf("PRAGMA table_info(sms_templates) error: %v", dbres.Error)
	}

	cols = nil
	dbres = models.DB().Raw("PRAGMA table_info(sms_campaigns);").Scan(&cols)
	if dbres.Error == nil {
		names := make([]string, 0, len(cols))
		for _, c := range cols { names = append(names, c.Name) }
		log.Printf("sms_campaigns columns: %v", names)
	} else {
		log.Printf("PRAGMA table_info(sms_campaigns) error: %v", dbres.Error)
	}

	cols = nil
	dbres = models.DB().Raw("PRAGMA table_info(sms_campaign_recipients);").Scan(&cols)
	if dbres.Error == nil {
		names := make([]string, 0, len(cols))
		for _, c := range cols { names = append(names, c.Name) }
		log.Printf("sms_campaign_recipients columns: %v", names)
	} else {
		log.Printf("PRAGMA table_info(sms_campaign_recipients) error: %v", dbres.Error)
	}
	// ==== end diagnostics ====


	as.handler = router
}

func (as *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	as.handler.ServeHTTP(w, r)
}
