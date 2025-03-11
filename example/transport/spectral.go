package main

import (
	"log/slog"

	"github.com/astraskyblock/spectrum"
	"github.com/astraskyblock/spectrum/server"
	"github.com/astraskyblock/spectrum/transport"
	"github.com/astraskyblock/spectrum/util"
	"github.com/sandertv/gophertunnel/minecraft"
)

func main() {
	logger := slog.Default()
	proxy := spectrum.NewSpectrum(server.NewStaticDiscovery("127.0.0.1:19133", ""), logger, nil, transport.NewSpectral(logger))
	if err := proxy.Listen(minecraft.ListenConfig{StatusProvider: util.NewStatusProvider("Spectrum Proxy", "Spectrum")}); err != nil {
		return
	}

	for {
		_, _ = proxy.Accept()
	}
}
