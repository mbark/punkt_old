# Roadmap to better managers

* [x] configuration
  * [x] use toml to configure `punkt`
  * [x] provide manager's via configuration
* [ ] managers
  * [x] hook up symlinks and repositories in toml
  * [x] `punkt {add,remove} {git,symlink}`
  * [x] git
    * [x] migrate to new manager interface
    * [x] include path in configuration
    * [x] `add` command
    * [x] `remove` command
    * [ ] updat test suite
  * [x] symlink
    * [x] migrate to new manager interface
    * [x] use array of tables configuration format
    * [x] `remove` commmand
    * [ ] updat test suite
  * [ ] generic
    * [x] add a generic manager which proxies via the given configuration
    * [ ] add a generic test suite which can be applied to `symlink` and `git`
          as well
  * [ ] clean up code and remove code duplication
