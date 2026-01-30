package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const emailSettingsFileName = "email-settings.json"
const emailTemplatesDir = "email"

const defaultDigestTemplate = `Scoli Daily Digest - {{date}}

Overdue
{{tasks_overdue}}

Due Today
{{tasks_today}}

Upcoming
{{tasks_upcoming}}

Tasks by Project
{{tasks_by_project}}

Yesterday's Notes
{{notes_summary}}

Completed Yesterday
{{completed_yesterday}}
`

const defaultDueTemplate = `Scoli Tasks Due - {{date}}

Overdue
{{tasks_overdue}}

Due Today
{{tasks_today}}
`

type EmailSettings struct {
	Version   int                   `json:"version"`
	Enabled   bool                  `json:"enabled"`
	SMTP      EmailSMTPSettings     `json:"smtp"`
	Digest    EmailDigestSettings   `json:"digest"`
	Due       EmailDueSettings      `json:"due"`
	Templates EmailTemplateSettings `json:"templates"`
}

type EmailSMTPSettings struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	From     string `json:"from"`
	To       string `json:"to"`
	UseTLS   bool   `json:"useTLS"`
}

type EmailDigestSettings struct {
	Enabled bool   `json:"enabled"`
	Time    string `json:"time"`
}

type EmailDueSettings struct {
	Enabled        bool   `json:"enabled"`
	Time           string `json:"time"`
	WindowDays     int    `json:"windowDays"`
	IncludeOverdue bool   `json:"includeOverdue"`
}

type EmailTemplateSettings struct {
	Digest string `json:"digest"`
	Due    string `json:"due"`
}

type EmailSettingsResponse struct {
	Settings EmailSettings `json:"settings"`
	Notice   string        `json:"notice,omitempty"`
}

type EmailSMTPSettingsPayload struct {
	Host     *string `json:"host,omitempty"`
	Port     *int    `json:"port,omitempty"`
	Username *string `json:"username,omitempty"`
	Password *string `json:"password,omitempty"`
	From     *string `json:"from,omitempty"`
	To       *string `json:"to,omitempty"`
	UseTLS   *bool   `json:"useTLS,omitempty"`
}

type EmailDigestSettingsPayload struct {
	Enabled *bool   `json:"enabled,omitempty"`
	Time    *string `json:"time,omitempty"`
}

type EmailDueSettingsPayload struct {
	Enabled        *bool   `json:"enabled,omitempty"`
	Time           *string `json:"time,omitempty"`
	WindowDays     *int    `json:"windowDays,omitempty"`
	IncludeOverdue *bool   `json:"includeOverdue,omitempty"`
}

type EmailTemplateSettingsPayload struct {
	Digest *string `json:"digest,omitempty"`
	Due    *string `json:"due,omitempty"`
}

type EmailSettingsPayload struct {
	Enabled   *bool                         `json:"enabled,omitempty"`
	SMTP      *EmailSMTPSettingsPayload     `json:"smtp,omitempty"`
	Digest    *EmailDigestSettingsPayload   `json:"digest,omitempty"`
	Due       *EmailDueSettingsPayload      `json:"due,omitempty"`
	Templates *EmailTemplateSettingsPayload `json:"templates,omitempty"`
}

func (s *Server) handleEmailSettingsGet(w http.ResponseWriter, r *http.Request) {
	settings, notice, err := s.loadEmailSettings()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to load email settings")
		return
	}

	resp := EmailSettingsResponse{Settings: settings}
	if notice != "" {
		resp.Notice = notice
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleEmailSettingsUpdate(w http.ResponseWriter, r *http.Request) {
	payload, err := decodeJSON[EmailSettingsPayload](r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	settings, _, err := s.loadEmailSettings()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to load email settings")
		return
	}

	if err := applyEmailSettingsPayload(&settings, payload); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := validateEmailSettings(settings); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := s.ensureEmailTemplates(settings); err != nil {
		writeError(w, http.StatusInternalServerError, "unable to ensure email templates")
		return
	}

	if err := s.saveEmailSettings(settings); err != nil {
		writeError(w, http.StatusInternalServerError, "unable to save email settings")
		return
	}

	s.logger.Info("email settings updated")
	writeJSON(w, http.StatusOK, settings)
}

func (s *Server) handleEmailTest(w http.ResponseWriter, r *http.Request) {
	settings, _, err := s.loadEmailSettings()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to load email settings")
		return
	}
	if err := validateEmailSettings(settings); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := validateEmailSMTP(settings.SMTP); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := sendEmail(settings.SMTP, "Scoli test email", "This is a test email from Scoli."); err != nil {
		writeError(w, http.StatusInternalServerError, "unable to send test email")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "sent"})
}

func (s *Server) emailSettingsFilePath() string {
	return filepath.Join(s.notesDir, emailSettingsFileName)
}

func (s *Server) loadEmailSettings() (EmailSettings, string, error) {
	path := s.emailSettingsFilePath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			settings := defaultEmailSettings()
			if err := os.MkdirAll(s.notesDir, 0o755); err != nil {
				return settings, "", err
			}
			if err := s.ensureEmailTemplates(settings); err != nil {
				return settings, "", err
			}
			if err := s.saveEmailSettings(settings); err != nil {
				return settings, "", err
			}
			s.logger.Info("email settings created", "path", path)
			return settings, "Created email-settings.json", nil
		}
		return EmailSettings{}, "", err
	}

	var settings EmailSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return EmailSettings{}, "", err
	}
	applyEmailSettingsDefaults(&settings)
	if err := s.ensureEmailTemplates(settings); err != nil {
		return settings, "", err
	}
	return settings, "", nil
}

func (s *Server) saveEmailSettings(settings EmailSettings) error {
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(s.emailSettingsFilePath(), data, 0o644)
}

func defaultEmailSettings() EmailSettings {
	return EmailSettings{
		Version: 1,
		Enabled: false,
		SMTP: EmailSMTPSettings{
			Host:   "smtp.gmail.com",
			Port:   587,
			UseTLS: true,
		},
		Digest: EmailDigestSettings{
			Enabled: true,
			Time:    "08:00",
		},
		Due: EmailDueSettings{
			Enabled:        true,
			Time:           "07:30",
			WindowDays:     0,
			IncludeOverdue: true,
		},
		Templates: EmailTemplateSettings{
			Digest: filepath.ToSlash(filepath.Join(emailTemplatesDir, "digest.template")),
			Due:    filepath.ToSlash(filepath.Join(emailTemplatesDir, "due.template")),
		},
	}
}

func applyEmailSettingsDefaults(settings *EmailSettings) {
	if settings.Version == 0 {
		settings.Version = 1
	}
	if settings.SMTP.Host == "" {
		settings.SMTP.Host = "smtp.gmail.com"
	}
	if settings.SMTP.Port == 0 {
		settings.SMTP.Port = 587
	}
	if settings.Digest.Time == "" {
		settings.Digest.Time = "08:00"
	}
	if settings.Due.Time == "" {
		settings.Due.Time = "07:30"
	}
	if settings.Due.WindowDays < 0 {
		settings.Due.WindowDays = 0
	}
	if settings.Templates.Digest == "" {
		settings.Templates.Digest = filepath.ToSlash(filepath.Join(emailTemplatesDir, "digest.template"))
	}
	if settings.Templates.Due == "" {
		settings.Templates.Due = filepath.ToSlash(filepath.Join(emailTemplatesDir, "due.template"))
	}
}

func applyEmailSettingsPayload(settings *EmailSettings, payload EmailSettingsPayload) error {
	if payload.Enabled != nil {
		settings.Enabled = *payload.Enabled
	}
	if payload.SMTP != nil {
		if payload.SMTP.Host != nil {
			settings.SMTP.Host = strings.TrimSpace(*payload.SMTP.Host)
		}
		if payload.SMTP.Port != nil {
			settings.SMTP.Port = *payload.SMTP.Port
		}
		if payload.SMTP.Username != nil {
			settings.SMTP.Username = strings.TrimSpace(*payload.SMTP.Username)
		}
		if payload.SMTP.Password != nil {
			settings.SMTP.Password = strings.TrimSpace(*payload.SMTP.Password)
		}
		if payload.SMTP.From != nil {
			settings.SMTP.From = strings.TrimSpace(*payload.SMTP.From)
		}
		if payload.SMTP.To != nil {
			settings.SMTP.To = strings.TrimSpace(*payload.SMTP.To)
		}
		if payload.SMTP.UseTLS != nil {
			settings.SMTP.UseTLS = *payload.SMTP.UseTLS
		}
	}
	if payload.Digest != nil {
		if payload.Digest.Enabled != nil {
			settings.Digest.Enabled = *payload.Digest.Enabled
		}
		if payload.Digest.Time != nil {
			settings.Digest.Time = strings.TrimSpace(*payload.Digest.Time)
		}
	}
	if payload.Due != nil {
		if payload.Due.Enabled != nil {
			settings.Due.Enabled = *payload.Due.Enabled
		}
		if payload.Due.Time != nil {
			settings.Due.Time = strings.TrimSpace(*payload.Due.Time)
		}
		if payload.Due.WindowDays != nil {
			settings.Due.WindowDays = *payload.Due.WindowDays
		}
		if payload.Due.IncludeOverdue != nil {
			settings.Due.IncludeOverdue = *payload.Due.IncludeOverdue
		}
	}
	if payload.Templates != nil {
		if payload.Templates.Digest != nil {
			cleaned, err := cleanRelPath(*payload.Templates.Digest)
			if err != nil {
				return err
			}
			if cleaned == "" {
				return errors.New("digest template path is required")
			}
			settings.Templates.Digest = cleaned
		}
		if payload.Templates.Due != nil {
			cleaned, err := cleanRelPath(*payload.Templates.Due)
			if err != nil {
				return err
			}
			if cleaned == "" {
				return errors.New("due template path is required")
			}
			settings.Templates.Due = cleaned
		}
	}
	applyEmailSettingsDefaults(settings)
	return nil
}

func validateEmailSettings(settings EmailSettings) error {
	if settings.Digest.Time != "" {
		if _, err := time.Parse("15:04", settings.Digest.Time); err != nil {
			return errors.New("digest time must be HH:MM")
		}
	}
	if settings.Due.Time != "" {
		if _, err := time.Parse("15:04", settings.Due.Time); err != nil {
			return errors.New("due time must be HH:MM")
		}
	}
	if settings.SMTP.Port <= 0 || settings.SMTP.Port > 65535 {
		return errors.New("smtp port must be between 1 and 65535")
	}
	if settings.Enabled {
		if err := validateEmailSMTP(settings.SMTP); err != nil {
			return err
		}
	}
	return nil
}

func validateEmailSMTP(settings EmailSMTPSettings) error {
	if settings.Host == "" {
		return errors.New("smtp host is required")
	}
	if settings.Username == "" {
		return errors.New("smtp username is required")
	}
	if settings.Password == "" {
		return errors.New("smtp password is required")
	}
	if settings.From == "" {
		return errors.New("smtp from address is required")
	}
	if settings.To == "" {
		return errors.New("smtp to address is required")
	}
	return nil
}

func (s *Server) ensureEmailTemplates(settings EmailSettings) error {
	templates := []struct {
		path    string
		content string
	}{
		{path: settings.Templates.Digest, content: defaultDigestTemplate},
		{path: settings.Templates.Due, content: defaultDueTemplate},
	}

	for _, tmpl := range templates {
		cleaned, err := cleanRelPath(tmpl.path)
		if err != nil {
			return err
		}
		if cleaned == "" {
			continue
		}
		absPath := filepath.Join(s.notesDir, filepath.FromSlash(cleaned))
		if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
			return err
		}
		if _, err := os.Stat(absPath); err == nil {
			continue
		} else if !os.IsNotExist(err) {
			return err
		}
		if err := os.WriteFile(absPath, []byte(tmpl.content), 0o644); err != nil {
			return err
		}
	}
	return nil
}
