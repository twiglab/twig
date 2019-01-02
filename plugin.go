package twig

type Plugin interface {
	ID() string
	Name() string
	Type() string
}

func GetPartner(id string, c Ctx) Plugin {
	t := c.Twig()
	return t.Plugin(id)

}
