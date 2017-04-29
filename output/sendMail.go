package output

import (
	"fmt"
	"net/smtp"
	"strings"

	"github.com/alerting/alert"
)

const (
	MailServer = "localhost:25"
	FromMail   = "Alert-NoReply@alert.com"
)

func SendMail(aBody string, aTo, aSub string) {
	alert.Logger.Debug("Inside SendMail")
	to := strings.Split(aTo, ";")
	alert.Logger.Debug(fmt.Sprintln("To =", to))

	msg := []byte("To:" + aTo + "\r\n" +
		"Subject: " + aSub + "\r\n" +
		"\r\n" +
		aBody + "\r\n")

	alert.Logger.Debug(fmt.Sprintln("Send Mail Msg =", string(msg)))

	err := smtp.SendMail(MailServer, nil, FromMail, to, msg)
	if err != nil {
		alert.Logger.Error(fmt.Sprintln("Error in Sending mail =", err))
	}

	return
}
