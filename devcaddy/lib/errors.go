package lib

const (
	ERROR_CONFIG_FILE = `
Config file was not found.
* You need to have a "devcaddy.json configuration file at
  the root of where you run the devcaddy command."
`
	ERROR_CONFIG_PARSE = `
Please check your configuration JSON syntax.
`

	ERROR_CONFIG_ROOT = `
Unable to determine your project root.
* If you did not specify a "root" in your config file, it is
  determined from your current working directory. There was
  a problem determining it. Please try running devcaddy again
  or specify the "root" manually.
* Note: If you specify a project root, it should be a relative
  path, ie. "../" or "../sub_folder"
`
	ERROR_PLUGIN_COMMAND_UNKNOWN = `
Could not determine the command for a plugin.
* If you did not specify a command, if will be determined
  from your plugin "path".
* Commands are mapped for the following file extensions:
  - .go (go)
  - .js (node)
  - .rb (ruby)
`
	ERROR_PLUGIN_DUPLICATE = `
Duplicate plugins detected.
* If you have not provide a "name" parameter for your
  plugin, it is determined from your "args".
* You cannot have multiple plugins with the same name.
  If you have plugins with the same file path in "args",
  you must provide a "name" to the plugins.
`
	ERROR_PLUGIN_NOT_DEFINED = `
Plugin not defined.
* You specified "plugins" for a file definition. However,
  this plugin was not defined. The plugin should have the
  same name as in the file definition.
`
)
