package twig

import (
	"github.com/twiglab/twig/internal/uuid"
)

type IdGenerator interface {
	NextID() string
}

const uuidPluginID = "_twig_uuid_plugin_id_"

type uuidPlugin struct {
}

func (id uuidPlugin) ID() string {
	return uuidPluginID
}

func (id uuidPlugin) Name() string {
	return uuidPluginID
}

func (id uuidPlugin) Type() string {
	return "plugin"
}

func (id uuidPlugin) NextID() string {
	return uuid.NewV1().String()
}

func newUuidPlugin() uuidPlugin {
	return uuidPlugin{}
}

func GetIdGenerator(c Ctx) IdGenerator {
	p, _ := GetPlugin(uuidPluginID, c)
	return p.(IdGenerator)
}
