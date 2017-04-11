from flask import Flask, request, jsonify
import os

app = Flask(__name__)

@app.route('/create', methods=['POST'])
def create_project():
    id = 123
    return jsonify(id=id, url="Project-%s" % request.json.get('title'))

@app.route('/_status/healthcheck/')
def healtchechk():
    return 'OK'

if __name__ == '__main__':
    app.run()
