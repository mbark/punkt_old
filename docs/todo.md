# Roadmap to better managers

* [ ] configuration
  * [x] use toml to configure `punkt`
  * [x] provide manager's via configuration
* [ ] managers
  * [ ] git
    * [ ] migrate to new manager interface
    * [ ] include path in configuration
    * [ ] `add` command
    * [ ] `remove` command
  * [ ] symlink
    * [ ] migrate to new manager interface
    * [ ] use array of tables configuration format
    * [ ] `remove` commmand
  * [ ] generic
    * [x] add a generic manager which proxies via the given configuration
    * [ ] add a generic test suite which can be applied to `symlink` and `git`
          as well
