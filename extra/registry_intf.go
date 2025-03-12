package extra

type RegistryInterface interface {
	GetSession(xuid string) SessionInterface
}
