#!/usr/bin/env python
# Reflects the requests from HTTP methods GET, POST, PUT, and DELETE
# Written by Nathan Hamiel (2010)
# Copied from https://gist.github.com/huyng/814831

from BaseHTTPServer import HTTPServer, BaseHTTPRequestHandler
from optparse import OptionParser

import consulate
import os
import signal
import sys

port = int(os.environ.get('PORT', 5000))
service_id = 'projects_%s' % port

consul = consulate.Consul()
# Add "projects" service to the local agent
consul.agent.service.register('projects', service_id=service_id, port=port)


def signal_handler(signal, frame):
    consul.agent.service.deregister(service_id)
    sys.exit(0)
signal.signal(signal.SIGINT, signal_handler)


class RequestHandler(BaseHTTPRequestHandler):
    
    def do_GET(self):
        
        request_path = self.path
        
        print("\n----- Request Start ----->\n")
        print(request_path)
        print(self.headers)
        print("<----- Request End -----\n")
        
        self.send_response(200)
        self.send_header("Set-Cookie", "foo=bar")
        
    def do_POST(self):
        
        request_path = self.path
        
        print("\n----- Request Start ----->\n")
        print(request_path)
        
        request_headers = self.headers
        #content_length = request_headers.getheaders('content-length')
        content_length = request_headers.getheaders('Content-Length')
        length = int(content_length[0]) if content_length else 0
        
        print(request_headers)
        print("Bytes received: %s\n" %  length)
        print(self.rfile.read(length))
        print("<----- Request End -----\n")
        
        self.send_response(200)
    
    do_PUT = do_POST
    do_DELETE = do_GET
        
def main():
    print('Listening on localhost:%s' % port)
    server = HTTPServer(('', port), RequestHandler)
    server.serve_forever()

        
if __name__ == "__main__":
    parser = OptionParser()
    parser.usage = ("Creates an http-server that will echo out any GET or POST parameters\n"
                    "Run:\n\n"
                    "   reflect")
    (options, args) = parser.parse_args()
    
    main()
