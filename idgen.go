package twig

import "github.com/twiglab/twig/internal/uuid"

const uuidPluginID = "_twig_uuid_plugin_id_"

type uuidGen struct {
}

func (id uuidGen) ID() string {
	return uuidPluginID
}

func (id uuidGen) Name() string {
	return uuidPluginID
}

func (id uuidGen) Type() string {
	return "plugin"
}

func (id uuidGen) NextID() string {
	return uuid.NewV1().String()
}

func GetUUIDGen(c Ctx) (gen IdGenerator) {
	gen, _ = GetIdGenerator(uuidPluginID, c)
	return
}
