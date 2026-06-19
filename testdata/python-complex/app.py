from flask import Flask, jsonify
import requests
from database import get_db_session
from models import User

app = Flask(__name__)

@app.route('/api/users')
def get_users():
    session = get_db_session()
    users = session.query(User).all()
    return jsonify([{'id': u.id, 'name': u.name} for u in users])

@app.route('/api/external')
def fetch_external():
    response = requests.get('https://api.example.com/data')
    return jsonify(response.json())

if __name__ == '__main__':
    app.run(debug=True)
