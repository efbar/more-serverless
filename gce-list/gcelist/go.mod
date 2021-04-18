module github.com/efbar/more-serverless/gce-list/gcelist

go 1.16

replace github.com/efbar/more-serverless/slack-message/slackmessage => ../../slack-message/slackmessage

require (
	github.com/efbar/more-serverless/slack-message/slackmessage v0.0.0-00010101000000-000000000000
	github.com/ryanuber/columnize v2.1.2+incompatible
	google.golang.org/api v0.44.0
)
