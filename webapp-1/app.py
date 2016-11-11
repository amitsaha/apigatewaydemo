from flask import Flask, request, jsonify

app =  Flask(__name__)

@app.route('/create', methods=['POST'])
def create_project():
    print request.headers
    print request.json
    id = 123
    return jsonify(id=id, url="Project-%s" % id)

if __name__ == '__main__':
    app.run(debug=True)
