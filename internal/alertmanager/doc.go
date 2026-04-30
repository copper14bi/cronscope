// Package alertmanager provides alert deduplication and dispatch for cronscope.
//
// It wraps the webhook client with a cooldown mechanism so that repeated
// failures for the same job do not flood the notification channel.
//
// Basic usage:
//
//	client := webhook.NewClient(cfg.WebhookURL)
//	mgr := alertmanager.New(client, 30*time.Minute)
//
//	// Dispatch an alert; suppressed if sent within the last 30 minutes.
//	sent, err := mgr.Send(alertmanager.Alert{
//		JobName:   "daily-backup",
//		Type:      alertmanager.AlertMissed,
//		Message:   "job has not run since expected time",
//		Timestamp: time.Now(),
//	})
//
//	// Reset deduplication state when a job recovers.
//	mgr.Reset("daily-backup")
package alertmanager
