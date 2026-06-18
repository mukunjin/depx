import requests
from flask import Flask, jsonify
import numpy as np
import pandas as pd

def main():
    app = Flask(__name__)
    
    # Make HTTP request
    response = requests.get("https://api.example.com")
    
    # Use numpy
    arr = np.array([1, 2, 3])
    
    # Use pandas
    df = pd.DataFrame({"col": [1, 2, 3]})
    
    return jsonify({"status": "ok"})

if __name__ == "__main__":
    main()
