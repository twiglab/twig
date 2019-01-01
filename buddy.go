package twig

type Partner interface {
	ID() string
	Name() string
	Type() string
}

func GetPartner(id string, c Ctx) Partner {
	t := c.Twig()
	if p, ok := t.Partner(id); ok {
		return p
	}

	c.Logger().Panicf("Twig: Partner (%s) is not exist!", id)

	return nil
}
