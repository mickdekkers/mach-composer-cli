package plugins

import (
	"fmt"

	"github.com/elliotchance/pie/v2"
)

type PluginRepository struct {
	Plugins map[string]MachComposerPlugin
}

func NewPluginRepository() *PluginRepository {
	return &PluginRepository{
		Plugins: make(map[string]MachComposerPlugin),
	}
}

func (p *PluginRepository) Load(name string, version string) error {
	if plugin, ok := localPlugins[name]; ok {
		p.Plugins[name] = plugin
		return nil
	}

	// Load remote (TODO)

	return fmt.Errorf("Plugin not found")
}

func (p *PluginRepository) Get(name string) (MachComposerPlugin, error) {
	if name == "" {
		return nil, fmt.Errorf("invalid plugin name given, received: %#v", name)
	}
	plugin, ok := p.Plugins[name]
	if !ok {
		return nil, fmt.Errorf("plugin %s not found", name)
	}
	return plugin, nil
}

func (p *PluginRepository) All() []MachComposerPlugin {
	plugins := pie.Values(p.Plugins)
	return pie.Filter(plugins, func(p MachComposerPlugin) bool { return p.IsEnabled() })
}

func (p *PluginRepository) SetRemoteState(name string, data map[string]any) error {
	plugin, err := p.Get(name)
	if err != nil {
		return err
	}
	return plugin.SetRemoteStateBackend(data)
}

func (p *PluginRepository) SetGlobalConfig(name string, data map[string]any) error {
	plugin, err := p.Get(name)
	if err != nil {
		return err
	}
	return plugin.SetGlobalConfig(data)
}

func (p *PluginRepository) SetSiteConfig(name string, site string, data map[string]any) error {
	plugin, err := p.Get(name)
	if err != nil {
		return err
	}
	return plugin.SetSiteConfig(site, data)
}

func (p *PluginRepository) SetSiteComponentConfig(site, component, name string, data map[string]any) error {
	plugin, err := p.Get(name)
	if err != nil {
		return err
	}
	return plugin.SetSiteComponentConfig(site, component, data)
}

func (p *PluginRepository) SetComponentConfig(name, component string, data map[string]any) error {
	plugin, err := p.Get(name)
	if err != nil {
		return err
	}
	return plugin.SetComponentConfig(component, data)
}