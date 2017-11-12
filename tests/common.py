import yaml


def create_conf_file(d, conf, name="config"):
    conf_file = d.join("%s.yaml" % name)
    conf_file.write(yaml.dump(conf))
    return conf_file
