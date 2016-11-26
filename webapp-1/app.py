from flask import Flask, request, jsonify
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

app = Flask(__name__)


@app.route('/create', methods=['POST'])
def create_project():
    id = 123
    return jsonify(id=id, url="Project-%s" % request.json.get('title'))

if __name__ == '__main__':
    app.run(debug=True, port=port)
