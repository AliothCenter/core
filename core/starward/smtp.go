package starward

import "html/template"

var (
	userRegistrationEmailTemplate                *template.Template
	userChangePasswordNoticeEmailTemplate        *template.Template
	userChangePhoneNoticeEmailTemplate           *template.Template
	applicationChangeSecretNoticeEmailTemplate   *template.Template
	applicationChangeCallbackNoticeEmailTemplate *template.Template
)
