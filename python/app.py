from flask import Flask, request, jsonify

app =  Flask(__name__)

@app.route('/create', methods=['POST'])
def create_project():
    print request.get_data()
    return jsonify({"id":12323, "title": "Hello"})

if __name__ == '__main__':
    app.run(debug=True)
