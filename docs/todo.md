# Roadmap to better managers

* [ ] configuration
  * [ ] use toml to configure `punkt`
  * [ ] provide manager's via configuration
* [ ] managers
  * [ ] git
    * [ ] include path in configuration
    * [ ] `add` command
    * [ ] `remove` command
  * [ ] symlink
    * [ ] use array of tables configuration format
    * [ ] `remove` commmand
  * [ ] generic
    * [ ] add a generic manager which proxies via the given configuration
    * [ ] add a generic test suite which can be applied to `symlink` and `git`
          as well
