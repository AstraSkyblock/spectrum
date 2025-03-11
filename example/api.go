package main

import (
	"log/slog"

	"github.com/astraskyblock/spectrum"
	"github.com/astraskyblock/spectrum/api"
	"github.com/astraskyblock/spectrum/server"
	"github.com/astraskyblock/spectrum/util"
	"github.com/sandertv/gophertunnel/minecraft"
)

func main() {
	logger := slog.Default()
	proxy := spectrum.NewSpectrum(server.NewStaticDiscovery("127.0.0.1:19133", ""), logger, nil, nil)
	if err := proxy.Listen(minecraft.ListenConfig{StatusProvider: util.NewStatusProvider("Spectrum Proxy", "Spectrum")}); err != nil {
		return
	}

	go func() {
		a := api.NewAPI(proxy.Registry(), logger, nil)
		if err := a.Listen("127.0.0.1:19132"); err != nil {
			return
		}

		for {
			_ = a.Accept()
		}
	}()

	for {
		_, _ = proxy.Accept()
	}
}
