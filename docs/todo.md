# Roadmap to better managers

* [x] configuration
  * [x] use toml to configure `punkt`
  * [x] provide manager's via configuration
* [x] managers
  * [x] hook up symlinks and repositories in toml
  * [x] `punkt {add,remove} {git,symlink}`
  * [x] git
    * [x] migrate to new manager interface
    * [x] include path in configuration
    * [x] `add` command
    * [x] `remove` command
    * [x] updat test suite
  * [x] symlink
    * [x] migrate to new manager interface
    * [x] use array of tables configuration format
    * [x] `remove` commmand
    * [x] updat test suite
  * [x] generic
    * [x] add a generic manager which proxies via the given configuration
  * [x] clean up code and remove code duplication
* [ ] printing
  * [x] use a user-facing logger that looks good
  * [x] look at `yarn` for other ideas (timing, emojis etc)
  * [ ] spread log messages throughout the app
