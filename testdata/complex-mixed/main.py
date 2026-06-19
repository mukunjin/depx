import requests
from flask import Flask
import numpy as np

app = Flask(__name__)

@app.route('/')
def index():
    data = np.array([1, 2, 3])
    response = requests.get('https://api.example.com')
    return {'data': data.tolist()}
