package utils


import (
    "github.com/mssola/user_agent"
)

type ParsedUserAgent struct {
    Browser         string `json:"browser"`
    OS              string `json:"os"`
    Device          string `json:"device"` 
    Platform        string `json:"platform"` 
}

func ParseUserAgent(uaString string) ParsedUserAgent {
    ua := user_agent.New(uaString)

    browser, _ := ua.Browser()

    os := ua.OS()

    platform := ua.Platform()

    deviceType := "desktop"
    if ua.Mobile() {
        deviceType = "mobile"
    }

    return ParsedUserAgent{
        Browser:        browser,
        OS:             os,
        Device:         deviceType,
        Platform:       platform,
    }
}