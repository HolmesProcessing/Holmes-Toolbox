import json

import os
from os import path
import traceback

import tornado
from tornado import web, httpserver, ioloop

import time

# reading configuration file

Metadata = {
        "Name"  :   "{$name}",
        "Version" : "{$version}",
        "Description" : "./README.md",
        "Copyright": "Copyright 2017 Holmes Group LLC",
        "License" : "./LICENSE"
    }

def ServiceConfig(filename):
    configPath = filename
    try:
        config = json.loads(open(configPath).read())
        return config
    except FileNotFoundError:
        raise tornado.web.HTTPError(500)

Config = ServiceConfig("./service.conf")

def {$name}Run(obj):
	### ADD YOUR SERVICE LOGIC HERE
    data = {}
    data["Hello"] = "HelloWorld"
    return data

class {$name}Process(tornado.web.RequestHandler):
    def get(self):
        try:
            filename = self.get_argument("obj", strip=False)
            fullPath = os.path.join('/tmp', filename)
            start_time = time.time()
            data = {$name}Run(fullPath)
            self.write(data)
        except tornado.web.MissingArgumentError:
            raise tornado.web.HTTPError(400)
        except TypeError as e:
            raise tornado.web.HTTPError(500)
        except Exception as e:
            self.write({"error": traceback.format_exc(e)})

class Info(tornado.web.RequestHandler):
    # Emits a string which describes the purpose of the analytics
    def get(self):
        info = """
            <p>{name:s} - {version:s}</p>
            <hr>
            <p>{description:s}</p>
            <hr>
            <p>{license:s}
        """.format(
                name = str(Metadata["Name"]).replace("\n", "<br>"),
                version = str(Metadata["Version"]).replace("\n", "<br>"),
                description = str(Metadata["Description"]).replace("\n", "<br>"),
                license = str(Metadata["License"]).replace("\n", "<br>")
        )
        self.write(info)

class {$name}App(tornado.web.Application):
    def __init__(self):
        for key in ["Description", "License"]:
            fpath = Metadata[key]
            if os.path.isfile(fpath):
                with open(fpath) as file:
                    Metadata[key] = file.read()

        handlers = [
                (r'/', Info),
                (r'/analyze/', {$name}Process),
            ]
        settings = dict(
            template_path = path.join(path.dirname(__file__), 'templates'),
            static_path = path.join(path.dirname(__file__), 'static'),
        )
        tornado.web.Application.__init__(self, handlers, **settings)
        self.engine = None

def main():
    server = tornado.httpserver.HTTPServer({$name}App())
    server.listen(Config["settings"]["port"])
    try:
        tornado.ioloop.IOLoop.current().start()
    except KeyboardInterrupt:
        tornado.ioloop.IOLoop.current().stop()

if __name__ == '__main__':
    main()
