package services

import (
	_ "embed"
	"bytes"
	"fmt"
	"text/template"
	"time"

	"github.com/abiosoft/colima/cli"
	"github.com/abiosoft/colima/util/downloader"
	"github.com/abiosoft/colima/embedded"
	"github.com/abiosoft/colima/environment"
)

const metallbVersion = "v0.13.4"

func InstallMetallb(
	host environment.HostActions, 
	guest environment.GuestActions, 
	a *cli.ActiveCommandChain,
	cidrBlock string,
) {
	metallbConfigPath := "/tmp/metallb-config.yaml"

	downloadPath := "/tmp/metallb-native.yaml"
	url := "https://raw.githubusercontent.com/metallb/metallb/" + metallbVersion + "/config/manifests/metallb-native.yaml"
	a.Stage("installing MetalLB")
	a.Retry("", time.Second*5, 30, func(retryCount int) error {
		return downloader.Download(host, guest, url, downloadPath)
	})
	a.Retry("", time.Second*5, 30, func(retryCount int) error {
		return guest.Run("kubectl", "apply", "-f", downloadPath)
	})

	a.Add(func() error {
		var availableData = map[string]string{
			"IpAddressRange": cidrBlock,
		}
		install, err := embedded.ReadString("metallb/config.yaml")
		if err != nil {
			return fmt.Errorf("error reading embedded metallb config: %w", err)
		}
		tmpl, err := template.New("config.yaml").Parse(install)
		if err != nil {
			return fmt.Errorf("error parsing embedded metallb config: %w", err)
		}
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, availableData); err != nil {
			return fmt.Errorf("error parsing embedded metallb config: %w", err)
		}
		return guest.Write(metallbConfigPath, buf.String())
	})

	a.Retry("", time.Second*5, 30, func(retryCount int) error {
		return guest.Run("kubectl", "apply", "-f", metallbConfigPath)
	})
}