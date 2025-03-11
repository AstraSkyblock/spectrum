package main

import (
	"image/color"
	"log/slog"

	"github.com/astraskyblock/spectrum"
	"github.com/astraskyblock/spectrum/server"
	"github.com/astraskyblock/spectrum/session/animation"
	"github.com/astraskyblock/spectrum/util"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

func main() {
	logger := slog.Default()
	proxy := spectrum.NewSpectrum(server.NewStaticDiscovery("127.0.0.1:19133", ""), logger, nil, nil)
	if err := proxy.Listen(minecraft.ListenConfig{StatusProvider: util.NewStatusProvider("Spectrum Proxy", "Spectrum")}); err != nil {
		return
	}

	for {
		s, err := proxy.Accept()
		if err != nil {
			continue
		}

		s.SetAnimation(&animation.Fade{
			Colour: color.RGBA{},
			Timing: protocol.CameraFadeTimeData{
				FadeInDuration:  0.25,
				WaitDuration:    4.50,
				FadeOutDuration: 0.25,
			},
		})
	}
}
