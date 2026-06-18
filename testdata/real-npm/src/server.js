const express = require('express');
const axios = require('axios');
const _ = require('lodash');
require('dotenv').config();

const app = express();

app.get('/api/users', async (req, res) => {
  try {
    const response = await axios.get('https://api.example.com/users');
    const users = _.sortBy(response.data, 'name');
    res.json(users);
  } catch (error) {
    res.status(500).json({ error: 'Failed to fetch users' });
  }
});

module.exports = app;
